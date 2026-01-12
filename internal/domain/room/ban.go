package room

import (
	"errors"
	"time"

	"github.com/vinib1903/cineus-api/internal/domain/user"
)

// BanID é o identificador único do banimento.
type BanID string

func (id BanID) String() string {
	return string(id)
}

// Ban representa um banimento de usuário em uma sala.
type Ban struct {
	ID        BanID
	RoomID    ID
	UserID    user.ID
	BannedBy  user.ID
	Reason    string
	ExpiresAt *time.Time
	CreatedAt time.Time
}

// Erros de banimento.
var (
	ErrUserBanned     = errors.New("user is banned from this room")
	ErrBanNotFound    = errors.New("ban not found")
	ErrCannotBanSelf  = errors.New("cannot ban yourself")
	ErrCannotBanOwner = errors.New("cannot ban the room owner")
)

// Constantes.
const (
	MaxBanReasonLength = 200
)

// NewBan cria um novo banimento.
func NewBan(id BanID, roomID ID, userID, bannedBy user.ID, reason string, expiresAt *time.Time) (*Ban, error) {
	// Não pode se auto-banir
	if userID == bannedBy {
		return nil, ErrCannotBanSelf
	}

	// Limitar tamanho do motivo
	if len(reason) > MaxBanReasonLength {
		reason = reason[:MaxBanReasonLength]
	}

	return &Ban{
		ID:        id,
		RoomID:    roomID,
		UserID:    userID,
		BannedBy:  bannedBy,
		Reason:    reason,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}

// IsExpired verifica se o banimento expirou.
func (b *Ban) IsExpired() bool {
	if b.ExpiresAt == nil {
		return false // Ban permanente nunca expira
	}
	return time.Now().After(*b.ExpiresAt)
}

// IsActive verifica se o banimento ainda está ativo.
func (b *Ban) IsActive() bool {
	return !b.IsExpired()
}
