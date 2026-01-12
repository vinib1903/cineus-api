package chat

import (
	"errors"
	"strings"
	"time"

	"github.com/vinib1903/cineus-api/internal/domain/room"
	"github.com/vinib1903/cineus-api/internal/domain/user"
)

// MessageID é o identificador único da mensagem.
type MessageID string

func (id MessageID) String() string {
	return string(id)
}

// Message representa uma mensagem de chat em uma sala.
type Message struct {
	ID        MessageID
	RoomID    room.ID
	UserID    user.ID
	Content   string
	CreatedAt time.Time
}

// Erros de chat.
var (
	ErrMessageTooLong  = errors.New("message too long (max 500 characters)")
	ErrMessageEmpty    = errors.New("message cannot be empty")
)

// Constantes.
const (
	MaxMessageLength = 500
)

// NewMessage cria uma nova mensagem com validações.
func NewMessage(id MessageID, roomID room.ID, userID user.ID, content string) (*Message, error) {
	content = strings.TrimSpace(content)

	if content == "" {
		return nil, ErrMessageEmpty
	}

	if len(content) > MaxMessageLength {
		return nil, ErrMessageTooLong
	}

	return &Message{
		ID:        id,
		RoomID:    roomID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}, nil
}

// DirectMessageID é o identificador único da mensagem direta.
type DirectMessageID string

func (id DirectMessageID) String() string {
	return string(id)
}

// DirectMessage representa uma mensagem direta entre dois usuários.
type DirectMessage struct {
	ID         DirectMessageID
	FromUserID user.ID
	ToUserID   user.ID
	Content    string
	ReadAt     *time.Time // nil = não lida
	CreatedAt  time.Time
}

// NewDirectMessage cria uma nova mensagem direta com validações.
func NewDirectMessage(id DirectMessageID, fromUserID, toUserID user.ID, content string) (*DirectMessage, error) {
	content = strings.TrimSpace(content)

	if content == "" {
		return nil, ErrMessageEmpty
	}

	if len(content) > MaxMessageLength {
		return nil, ErrMessageTooLong
	}

	return &DirectMessage{
		ID:         id,
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Content:    content,
		ReadAt:     nil,
		CreatedAt:  time.Now(),
	}, nil
}

// IsRead verifica se a mensagem foi lida.
func (dm *DirectMessage) IsRead() bool {
	return dm.ReadAt != nil
}

// MarkAsRead marca a mensagem como lida.
func (dm *DirectMessage) MarkAsRead() {
	if dm.ReadAt == nil {
		now := time.Now()
		dm.ReadAt = &now
	}
}
