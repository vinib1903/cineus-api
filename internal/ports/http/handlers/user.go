package handlers

import (
	"net/http"

	"github.com/vinib1903/cineus-api/internal/domain/user"
	"github.com/vinib1903/cineus-api/internal/ports/http/httputil"
)

// UserHandler gerencia as rotas de usuário.
type UserHandler struct {
	userRepo user.Repository
}

// NewUserHandler cria uma nova instância do handler.
func NewUserHandler(userRepo user.Repository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

// MeResponse é a resposta do endpoint /me.
type MeResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"display_name"`
	XP            int64  `json:"xp"`
	EmailVerified bool   `json:"email_verified"`
}

// Me retorna os dados do usuário autenticado.
// GET /api/v1/me
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	// Extrair o ID do usuário do contexto (colocado pelo AuthMiddleware)
	userID := httputil.GetUserID(r.Context())
	if userID == "" {
		httputil.Unauthorized(w, "User not authenticated")
		return
	}

	// Buscar o usuário no banco
	u, err := h.userRepo.GetByID(r.Context(), user.ID(userID))
	if err != nil {
		if err == user.ErrUserNotFound {
			httputil.NotFound(w, "User not found")
			return
		}
		httputil.InternalServerError(w, "Failed to get user")
		return
	}

	// Montar resposta
	response := MeResponse{
		ID:            string(u.ID),
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		XP:            u.XP,
		EmailVerified: u.EmailVerified,
	}

	httputil.JSON(w, http.StatusOK, response)
}
