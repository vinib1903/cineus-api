package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/vinib1903/cineus-api/internal/infra/auth"
	"github.com/vinib1903/cineus-api/internal/ports/http/httputil"
)

// AuthMiddleware cria um middleware que valida tokens JWT.
func AuthMiddleware(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extrair o token do header Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httputil.Unauthorized(w, "Authorization header is required")
				return
			}

			// O formato esperado é: "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				httputil.Unauthorized(w, "Invalid authorization header format")
				return
			}

			tokenString := parts[1]

			// Validar o token
			claims, err := jwtManager.ValidateToken(tokenString)
			if err != nil {
				if err == auth.ErrExpiredToken {
					httputil.Unauthorized(w, "Token has expired")
					return
				}
				httputil.Unauthorized(w, "Invalid token")
				return
			}

			// Verificar se é um access token (não refresh token)
			if claims.TokenType != auth.AccessToken {
				httputil.Unauthorized(w, "Invalid token type")
				return
			}

			// Adicionar informações do usuário ao contexto
			ctx := context.WithValue(r.Context(), httputil.UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, httputil.UserEmailKey, claims.Email)

			// Chamar o próximo handler com o contexto atualizado
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
