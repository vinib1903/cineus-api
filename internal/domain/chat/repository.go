package chat

import (
	"context"
	"errors"
	"time"

	"github.com/vinib1903/cineus-api/internal/domain/room"
	"github.com/vinib1903/cineus-api/internal/domain/user"
)

// Erros de repositório.
var (
	ErrMessageNotFound = errors.New("message not found")
)

// MessageRepository define as operações de persistência para mensagens de sala.
type MessageRepository interface {
	// Create salva uma nova mensagem.
	Create(ctx context.Context, msg *Message) error

	// ListByRoom retorna mensagens de uma sala.
	// Ordenadas por data (mais recentes primeiro).
	// before: retorna mensagens anteriores a este timestamp (para paginação).
	// limit: quantidade máxima de mensagens.
	ListByRoom(ctx context.Context, roomID room.ID, before *time.Time, limit int) ([]*Message, error)
}

// DirectMessageRepository define as operações para mensagens diretas.
type DirectMessageRepository interface {
	// Create salva uma nova mensagem direta.
	Create(ctx context.Context, dm *DirectMessage) error

	// ListConversation retorna mensagens entre dois usuários.
	// Ordenadas por data (mais recentes primeiro).
	ListConversation(ctx context.Context, userA, userB user.ID, before *time.Time, limit int) ([]*DirectMessage, error)

	// MarkAsRead marca mensagens como lidas.
	// Marca todas as mensagens de fromUserID para toUserID como lidas.
	MarkAsRead(ctx context.Context, fromUserID, toUserID user.ID) error

	// CountUnread conta mensagens não lidas para um usuário.
	CountUnread(ctx context.Context, userID user.ID) (int, error)
}
