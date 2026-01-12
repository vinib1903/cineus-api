package room

import (
	"context"
	"errors"

	"github.com/vinib1903/cineus-api/internal/domain/user"
)

// Erros de repositório.
var (
	ErrRoomNotFound = errors.New("room not found")
)

// Repository define as operações de persistência para Room.
type Repository interface {
	// Create salva uma nova sala no banco.
	Create(ctx context.Context, room *Room) error

	// GetByID busca uma sala pelo ID.
	// Retorna ErrRoomNotFound se não existir.
	// Não retorna salas deletadas (DeletedAt != nil).
	GetByID(ctx context.Context, id ID) (*Room, error)

	// GetByAccessCode busca uma sala pelo código de acesso.
	// Retorna ErrRoomNotFound se não existir.
	// Não retorna salas deletadas.
	GetByAccessCode(ctx context.Context, code string) (*Room, error)

	// Update atualiza os dados de uma sala existente.
	// Retorna ErrRoomNotFound se não existir.
	Update(ctx context.Context, room *Room) error

	// ListPublic retorna todas as salas públicas não deletadas.
	// Ordenadas por data de criação (mais recentes primeiro).
	// Suporta paginação com limit e offset.
	ListPublic(ctx context.Context, limit, offset int) ([]*Room, error)

	// ListByOwner retorna todas as salas de um usuário.
	// Inclui públicas e privadas, mas não deletadas.
	ListByOwner(ctx context.Context, ownerID user.ID) ([]*Room, error)

	// CountByOwner conta quantas salas ativas um usuário possui.
	// Usado para verificar o limite de 2 salas por usuário.
	CountByOwner(ctx context.Context, ownerID user.ID) (int, error)
}
