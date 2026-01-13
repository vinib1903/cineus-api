package auth

import (
	"github.com/google/uuid"
)

// IDGenerator gera IDs únicos.
type IDGenerator struct{}

// NewIDGenerator cria uma nova instância do gerador.
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

// NewID gera um novo UUID v4.
func (g *IDGenerator) NewID() string {
	return uuid.New().String()
}
