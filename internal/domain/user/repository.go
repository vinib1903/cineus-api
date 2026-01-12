package user

import (
	"context"
	"errors"
)

// Erros de repositório.
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

// Repository define as operações de persistência para User.
type Repository interface {
	// Create salva um novo usuário no banco.
	// Retorna ErrUserAlreadyExists se o email já existir.
	Create(ctx context.Context, user *User) error

	// GetByID busca um usuário pelo ID.
	// Retorna ErrUserNotFound se não existir.
	GetByID(ctx context.Context, id ID) (*User, error)

	// GetByEmail busca um usuário pelo email.
	// Retorna ErrUserNotFound se não existir.
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Update atualiza os dados de um usuário existente.
	// Retorna ErrUserNotFound se não existir.
	Update(ctx context.Context, user *User) error

	// ExistsByEmail verifica se já existe um usuário com este email.
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
