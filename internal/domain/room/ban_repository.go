package room

import (
	"context"

	"github.com/vinib1903/cineus-api/internal/domain/user"
)

// BanRepository define as operações de persistência para Ban.
type BanRepository interface {
	// Create salva um novo banimento.
	Create(ctx context.Context, ban *Ban) error

	// GetActiveBan busca um banimento ativo de um usuário em uma sala.
	// Retorna ErrBanNotFound se não houver ban ativo.
	GetActiveBan(ctx context.Context, roomID ID, userID user.ID) (*Ban, error)

	// IsUserBanned verifica se um usuário está banido de uma sala.
	// Considera apenas bans ativos (não expirados).
	IsUserBanned(ctx context.Context, roomID ID, userID user.ID) (bool, error)

	// Delete remove um banimento (unban).
	Delete(ctx context.Context, id BanID) error

	// ListByRoom lista todos os bans ativos de uma sala.
	ListByRoom(ctx context.Context, roomID ID) ([]*Ban, error)
}
