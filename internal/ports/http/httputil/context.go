package httputil

import (
	"context"
)

// ContextKey é um tipo para chaves de contexto.
type ContextKey string

const (
	// UserIDKey é a chave para o ID do usuário no contexto.
	UserIDKey ContextKey = "user_id"
	// UserEmailKey é a chave para o email do usuário no contexto.
	UserEmailKey ContextKey = "user_email"
)

// GetUserID extrai o ID do usuário do contexto.
func GetUserID(ctx context.Context) string {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}

// GetUserEmail extrai o email do usuário do contexto.
func GetUserEmail(ctx context.Context) string {
	email, ok := ctx.Value(UserEmailKey).(string)
	if !ok {
		return ""
	}
	return email
}
