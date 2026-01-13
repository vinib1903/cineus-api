package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vinib1903/cineus-api/internal/domain/user"
)

// UserRepository implementa user.Repository
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository cria uma nova instância do repositório.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create salva um novo usuário no banco.
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, display_name, xp, email_verified, created_at, updated_at, last_login_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		u.ID,
		u.Email,
		u.PasswordHash,
		u.DisplayName,
		u.XP,
		u.EmailVerified,
		u.CreatedAt,
		u.UpdatedAt,
		u.LastLoginAt,
	)

	if err != nil {
		if isDuplicateKeyError(err) {
			return user.ErrUserAlreadyExists
		}
		return err
	}

	return nil
}

// GetByID busca um usuário pelo ID.
func (r *UserRepository) GetByID(ctx context.Context, id user.ID) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, display_name, xp, email_verified, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`

	return r.scanUser(r.pool.QueryRow(ctx, query, id))
}

// GetByEmail busca um usuário pelo email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, display_name, xp, email_verified, created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1
	`

	return r.scanUser(r.pool.QueryRow(ctx, query, email))
}

// Update atualiza os dados de um usuário existente.
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	query := `
		UPDATE users
		SET email = $2,
		    password_hash = $3,
		    display_name = $4,
		    xp = $5,
		    email_verified = $6,
		    updated_at = $7,
		    last_login_at = $8
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		u.ID,
		u.Email,
		u.PasswordHash,
		u.DisplayName,
		u.XP,
		u.EmailVerified,
		u.UpdatedAt,
		u.LastLoginAt,
	)

	if err != nil {
		return err
	}

	// Verificar se alguma linha foi atualizada
	if result.RowsAffected() == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// ExistsByEmail verifica se já existe um usuário com este email.
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// scanUser converte uma linha do banco em um User.
func (r *UserRepository) scanUser(row pgx.Row) (*user.User, error) {
	var u user.User

	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.DisplayName,
		&u.XP,
		&u.EmailVerified,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoginAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

// isDuplicateKeyError verifica se o erro é de chave duplicada.
func isDuplicateKeyError(err error) bool {
	// O código de erro do PostgreSQL para unique violation é 23505
	return err != nil && contains(err.Error(), "23505")
}

// contains verifica se uma string contém outra (helper simples).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
