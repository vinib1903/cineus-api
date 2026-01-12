package room

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/vinib1903/cineus-api/internal/domain/user"
)

// ID é o identificador único da sala.
type ID string

func (id ID) String() string {
	return string(id)
}

func (id ID) IsEmpty() bool {
	return id == ""
}

// Visibility define se a sala é pública ou privada.
type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)

// Theme representa o tema visual da sala.
type Theme string

const (
	ThemeDefault Theme = "default" // Cinema
	ThemeFarm    Theme = "farm"    // Fazenda
	ThemeHorror  Theme = "horror"  // Terror
	ThemeFun     Theme = "fun"     // Divertida
	ThemeSpace   Theme = "space"   // Espacial
)

// Room representa uma sala de cinema virtual.
type Room struct {
	ID         ID
	OwnerID    user.ID
	Name       string
	Theme      Theme
	Visibility Visibility
	AccessCode *string // Código para entrar (salas privadas)
	MaxSeats   int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time // Soft delete
}

// Erros de domínio da sala.
var (
	ErrNameTooShort       = errors.New("room name too short (min 3 characters)")
	ErrNameTooLong        = errors.New("room name too long (max 25 characters)")
	ErrInvalidTheme       = errors.New("invalid theme")
	ErrInvalidVisibility  = errors.New("invalid visibility")
	ErrRoomDeleted        = errors.New("room has been deleted")
	ErrNotOwner           = errors.New("only the owner can perform this action")
	ErrRoomNotEmpty       = errors.New("room is not empty")
	ErrInvalidAccessCode  = errors.New("invalid access code")
	ErrAccessCodeRequired = errors.New("access code is required for private rooms")
)

// Constantes de validação.
const (
	MinNameLength    = 3
	MaxNameLength    = 25
	AccessCodeLength = 4
	DefaultMaxSeats  = 16
)

// NewRoom cria uma nova sala com validações.
func NewRoom(id ID, ownerID user.ID, name string, theme Theme, visibility Visibility) (*Room, error) {
	// Validar nome
	if err := validateName(name); err != nil {
		return nil, err
	}

	// Validar tema
	if !isValidTheme(theme) {
		return nil, ErrInvalidTheme
	}

	// Validar visibilidade
	if !isValidVisibility(visibility) {
		return nil, ErrInvalidVisibility
	}

	now := time.Now()

	room := &Room{
		ID:         id,
		OwnerID:    ownerID,
		Name:       strings.TrimSpace(name),
		Theme:      theme,
		Visibility: visibility,
		MaxSeats:   DefaultMaxSeats,
		CreatedAt:  now,
		UpdatedAt:  now,
		DeletedAt:  nil,
	}

	// Gerar código de acesso para salas privadas
	if visibility == VisibilityPrivate {
		code, err := generateAccessCode()
		if err != nil {
			return nil, err
		}
		room.AccessCode = &code
	}

	return room, nil
}

// validateName verifica se o nome da sala é válido.
func validateName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) < MinNameLength {
		return ErrNameTooShort
	}

	if len(name) > MaxNameLength {
		return ErrNameTooLong
	}

	return nil
}

// isValidTheme verifica se o tema é válido.
func isValidTheme(theme Theme) bool {
	switch theme {
	case ThemeFarm, ThemeHorror, ThemeFun, ThemeSpace, ThemeDefault:
		return true
	default:
		return false
	}
}

// isValidVisibility verifica se a visibilidade é válida.
func isValidVisibility(visibility Visibility) bool {
	switch visibility {
	case VisibilityPublic, VisibilityPrivate:
		return true
	default:
		return false
	}
}

// generateAccessCode gera um código aleatório de 4 caracteres.
func generateAccessCode() (string, error) {
	// Gerar 2 bytes aleatórios (= 4 caracteres em hex)
	bytes := make([]byte, 2)

	// crypto/rand é seguro para criptografia
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Converter para hexadecimal e deixar maiúsculo
	return strings.ToUpper(hex.EncodeToString(bytes)), nil
}

// IsDeleted verifica se a sala foi deletada.
func (r *Room) IsDeleted() bool {
	return r.DeletedAt != nil
}

// IsPublic verifica se a sala é pública.
func (r *Room) IsPublic() bool {
	return r.Visibility == VisibilityPublic
}

// IsPrivate verifica se a sala é privada.
func (r *Room) IsPrivate() bool {
	return r.Visibility == VisibilityPrivate
}

// IsOwner verifica se o usuário é o dono da sala.
func (r *Room) IsOwner(userID user.ID) bool {
	return r.OwnerID == userID
}

// ValidateAccess verifica se o código de acesso está correto.
// Para salas públicas, sempre retorna true.
func (r *Room) ValidateAccess(code string) bool {
	if r.IsPublic() {
		return true
	}

	if r.AccessCode == nil {
		return false
	}

	return strings.EqualFold(*r.AccessCode, code)
}

// Delete marca a sala como deletada (soft delete).
// Apenas o dono pode deletar, e a sala deve estar vazia.
func (r *Room) Delete(requesterID user.ID, isEmpty bool) error {
	if r.IsDeleted() {
		return ErrRoomDeleted
	}

	if !r.IsOwner(requesterID) {
		return ErrNotOwner
	}

	if !isEmpty {
		return ErrRoomNotEmpty
	}

	now := time.Now()
	r.DeletedAt = &now
	r.UpdatedAt = now
	return nil
}

// UpdateName atualiza o nome da sala.
func (r *Room) UpdateName(requesterID user.ID, name string) error {
	if r.IsDeleted() {
		return ErrRoomDeleted
	}

	if !r.IsOwner(requesterID) {
		return ErrNotOwner
	}

	if err := validateName(name); err != nil {
		return err
	}

	r.Name = strings.TrimSpace(name)
	r.UpdatedAt = time.Now()
	return nil
}

// UpdateTheme atualiza o tema da sala.
func (r *Room) UpdateTheme(requesterID user.ID, theme Theme) error {
	if r.IsDeleted() {
		return ErrRoomDeleted
	}

	if !r.IsOwner(requesterID) {
		return ErrNotOwner
	}

	if !isValidTheme(theme) {
		return ErrInvalidTheme
	}

	r.Theme = theme
	r.UpdatedAt = time.Now()
	return nil
}

// RegenerateAccessCode gera um novo código de acesso.
func (r *Room) RegenerateAccessCode(requesterID user.ID) error {
	if r.IsDeleted() {
		return ErrRoomDeleted
	}

	if !r.IsOwner(requesterID) {
		return ErrNotOwner
	}

	if r.IsPublic() {
		return nil // Salas públicas não têm código
	}

	code, err := generateAccessCode()
	if err != nil {
		return err
	}

	r.AccessCode = &code
	r.UpdatedAt = time.Now()
	return nil
}
