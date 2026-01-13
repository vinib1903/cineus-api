package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vinib1903/cineus-api/internal/domain/room"
	"github.com/vinib1903/cineus-api/internal/domain/user"
)

// RoomRepository implementa room.Repository
type RoomRepository struct {
	pool *pgxpool.Pool
}

// NewRoomRepository cria uma nova instância do repositório.
func NewRoomRepository(pool *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{pool: pool}
}

// Create salva uma nova sala no banco.
func (r *RoomRepository) Create(ctx context.Context, rm *room.Room) error {
	query := `
		INSERT INTO rooms (id, owner_id, name, theme, visibility, access_code, max_seats, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(ctx, query,
		rm.ID,
		rm.OwnerID,
		rm.Name,
		rm.Theme,
		rm.Visibility,
		rm.AccessCode,
		rm.MaxSeats,
		rm.CreatedAt,
		rm.UpdatedAt,
		rm.DeletedAt,
	)

	return err
}

// GetByID busca uma sala pelo ID.
func (r *RoomRepository) GetByID(ctx context.Context, id room.ID) (*room.Room, error) {
	query := `
		SELECT id, owner_id, name, theme, visibility, access_code, max_seats, created_at, updated_at, deleted_at
		FROM rooms
		WHERE id = $1 AND deleted_at IS NULL
	`

	return r.scanRoom(r.pool.QueryRow(ctx, query, id))
}

// GetByAccessCode busca uma sala pelo código de acesso.
func (r *RoomRepository) GetByAccessCode(ctx context.Context, code string) (*room.Room, error) {
	query := `
		SELECT id, owner_id, name, theme, visibility, access_code, max_seats, created_at, updated_at, deleted_at
		FROM rooms
		WHERE UPPER(access_code) = UPPER($1) AND deleted_at IS NULL
	`

	return r.scanRoom(r.pool.QueryRow(ctx, query, code))
}

// Update atualiza os dados de uma sala existente.
func (r *RoomRepository) Update(ctx context.Context, rm *room.Room) error {
	query := `
		UPDATE rooms
		SET name = $2,
		    theme = $3,
		    visibility = $4,
		    access_code = $5,
		    max_seats = $6,
		    updated_at = $7,
		    deleted_at = $8
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		rm.ID,
		rm.Name,
		rm.Theme,
		rm.Visibility,
		rm.AccessCode,
		rm.MaxSeats,
		rm.UpdatedAt,
		rm.DeletedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return room.ErrRoomNotFound
	}

	return nil
}

// ListPublic retorna todas as salas públicas não deletadas.
func (r *RoomRepository) ListPublic(ctx context.Context, limit, offset int) ([]*room.Room, error) {
	query := `
		SELECT id, owner_id, name, theme, visibility, access_code, max_seats, created_at, updated_at, deleted_at
		FROM rooms
		WHERE visibility = 'public' AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRooms(rows)
}

// ListByOwner retorna todas as salas de um usuário.
func (r *RoomRepository) ListByOwner(ctx context.Context, ownerID user.ID) ([]*room.Room, error) {
	query := `
		SELECT id, owner_id, name, theme, visibility, access_code, max_seats, created_at, updated_at, deleted_at
		FROM rooms
		WHERE owner_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRooms(rows)
}

// CountByOwner conta quantas salas ativas um usuário possui.
func (r *RoomRepository) CountByOwner(ctx context.Context, ownerID user.ID) (int, error) {
	query := `SELECT COUNT(*) FROM rooms WHERE owner_id = $1 AND deleted_at IS NULL`

	var count int
	err := r.pool.QueryRow(ctx, query, ownerID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// scanRoom converte uma linha do banco em um Room.
func (r *RoomRepository) scanRoom(row pgx.Row) (*room.Room, error) {
	var rm room.Room

	err := row.Scan(
		&rm.ID,
		&rm.OwnerID,
		&rm.Name,
		&rm.Theme,
		&rm.Visibility,
		&rm.AccessCode,
		&rm.MaxSeats,
		&rm.CreatedAt,
		&rm.UpdatedAt,
		&rm.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, room.ErrRoomNotFound
		}
		return nil, err
	}

	return &rm, nil
}

// scanRooms converte múltiplas linhas em uma lista de Rooms.
func (r *RoomRepository) scanRooms(rows pgx.Rows) ([]*room.Room, error) {
	var rooms []*room.Room

	for rows.Next() {
		var rm room.Room
		err := rows.Scan(
			&rm.ID,
			&rm.OwnerID,
			&rm.Name,
			&rm.Theme,
			&rm.Visibility,
			&rm.AccessCode,
			&rm.MaxSeats,
			&rm.CreatedAt,
			&rm.UpdatedAt,
			&rm.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, &rm)
	}

	// Verificar erros após iterar
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}
