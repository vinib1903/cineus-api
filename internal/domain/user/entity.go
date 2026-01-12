package user

import (
	"errors"
	"strings"
	"time"

	"github.com/badoux/checkmail"
)

// ID é o identificador único do usuário.
// Usamos um tipo próprio (não string puro) para evitar confusão.
// Exemplo: não misturar UserID com RoomID acidentalmente.
type ID string

// String retorna o ID como string.
// Isso implementa a interface fmt.Stringer.
func (id ID) String() string {
	return string(id)
}

// IsEmpty verifica se o ID está vazio.
func (id ID) IsEmpty() bool {
	return id == ""
}

// User representa um usuário da plataforma.
type User struct {
	ID            ID
	Email         string
	PasswordHash  string
	DisplayName   string
	XP            int64
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastLoginAt   *time.Time // Ponteiro porque pode ser nulo (nunca logou)
}

// Erros de domínio do usuário.
// Definimos como variáveis para poder comparar depois.
var (
	ErrInvalidEmail        = errors.New("invalid email")
	ErrDisplayNameTooLong  = errors.New("display name too long (max 50 characters)")
	ErrDisplayNameTooShort = errors.New("display name too short (min 3 characters)")
	ErrEmptyPassword       = errors.New("password cannot be empty")
	ErrPasswordTooShort    = errors.New("password too short (min 8 characters)")
)

// Constantes de validação.
const (
	MaxDisplayNameLength = 50
	MinDisplayNameLength = 3
	MinPasswordLength    = 8
)

// NewUser cria um novo usuário com validações.
// Esta é a única forma de criar um User válido.
func NewUser(id ID, email, passwordHash, displayName string) (*User, error) {
	// Validar email
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	// Validar nome
	if err := validateDisplayName(displayName); err != nil {
		return nil, err
	}

	// Validar que passwordHash não está vazio
	// (a validação da senha em si é feita antes de gerar o hash)
	if passwordHash == "" {
		return nil, ErrEmptyPassword
	}

	now := time.Now()

	return &User{
		ID:            id,
		Email:         strings.ToLower(strings.TrimSpace(email)),
		PasswordHash:  passwordHash,
		DisplayName:   strings.TrimSpace(displayName),
		XP:            0,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastLoginAt:   nil,
	}, nil
}

// validateEmail verifica se o email é válido.
func validateEmail(email string) error {
	email = strings.TrimSpace(email)

	if err := checkmail.ValidateFormat(email); err != nil {
		return ErrInvalidEmail
	}

	return nil
}

// validateDisplayName verifica se o nome é válido.
func validateDisplayName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) < MinDisplayNameLength {
		return ErrDisplayNameTooShort
	}

	if len(name) > MaxDisplayNameLength {
		return ErrDisplayNameTooLong
	}

	return nil
}

// ValidatePassword verifica se a senha atende os requisitos.
// Chamada ANTES de gerar o hash.
func ValidatePassword(password string) error {
	if password == "" {
		return ErrEmptyPassword
	}

	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	return nil
}

// UpdateDisplayName atualiza o nome do usuário.
func (u *User) UpdateDisplayName(name string) error {
	if err := validateDisplayName(name); err != nil {
		return err
	}

	u.DisplayName = strings.TrimSpace(name)
	u.UpdatedAt = time.Now()
	return nil
}

// AddXP adiciona pontos de experiência ao usuário.
func (u *User) AddXP(amount int64) {
	if amount > 0 {
		u.XP += amount
		u.UpdatedAt = time.Now()
	}
}

// VerifyEmail marca o email como verificado.
func (u *User) VerifyEmail() {
	u.EmailVerified = true
	u.UpdatedAt = time.Now()
}

// RecordLogin registra o momento do login.
func (u *User) RecordLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}
