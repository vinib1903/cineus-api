package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher gerencia o hash de senhas.
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher cria uma nova instância do hasher.
// cost define a "dificuldade" do hash (maior = mais seguro, mais lento).
// Recomendado: 10-12 para produção.
func NewPasswordHasher(cost int) *PasswordHasher {
	if cost < bcrypt.MinCost {
		cost = bcrypt.DefaultCost
	}
	return &PasswordHasher{cost: cost}
}

// Hash gera um hash seguro da senha.
func (h *PasswordHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Compare verifica se a senha corresponde ao hash.
// Retorna nil se a senha estiver correta.
func (h *PasswordHasher) Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
