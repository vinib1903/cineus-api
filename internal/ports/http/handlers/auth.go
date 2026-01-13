package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/vinib1903/cineus-api/internal/app/auth"
	"github.com/vinib1903/cineus-api/internal/domain/user"
	"github.com/vinib1903/cineus-api/internal/ports/http/httputil"
)

// AuthHandler gerencia as rotas de autenticação.
type AuthHandler struct {
	authService *auth.Service
}

// NewAuthHandler cria uma nova instância do handler.
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRequest é o corpo da requisição de registro.
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// AuthResponse é a resposta de autenticação (registro ou login).
type AuthResponse struct {
	User   UserResponse   `json:"user"`
	Tokens TokensResponse `json:"tokens"`
}

// UserResponse é a representação do usuário na resposta.
type UserResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	XP          int64  `json:"xp"`
}

// TokensResponse é a representação dos tokens na resposta.
type TokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Register cria uma nova conta de usuário.
// POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Decodificar o corpo da requisição
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "Invalid request body")
		return
	}

	// Validar campos obrigatórios
	if req.Email == "" {
		httputil.BadRequest(w, "Email is required")
		return
	}
	if req.Password == "" {
		httputil.BadRequest(w, "Password is required")
		return
	}
	if req.DisplayName == "" {
		httputil.BadRequest(w, "Display name is required")
		return
	}

	// Chamar o serviço
	output, err := h.authService.Register(r.Context(), auth.RegisterInput{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})

	if err != nil {
		handleAuthError(w, err)
		return
	}

	// Montar resposta
	response := AuthResponse{
		User: UserResponse{
			ID:          string(output.User.ID),
			Email:       output.User.Email,
			DisplayName: output.User.DisplayName,
			XP:          output.User.XP,
		},
		Tokens: TokensResponse{
			AccessToken:  output.Tokens.AccessToken,
			RefreshToken: output.Tokens.RefreshToken,
		},
	}

	httputil.JSON(w, http.StatusCreated, response)
}

// LoginRequest é o corpo da requisição de login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login autentica um usuário existente.
// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Decodificar o corpo da requisição
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "Invalid request body")
		return
	}

	// Validar campos obrigatórios
	if req.Email == "" {
		httputil.BadRequest(w, "Email is required")
		return
	}
	if req.Password == "" {
		httputil.BadRequest(w, "Password is required")
		return
	}

	// Chamar o serviço
	output, err := h.authService.Login(r.Context(), auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		handleAuthError(w, err)
		return
	}

	// Montar resposta
	response := AuthResponse{
		User: UserResponse{
			ID:          string(output.User.ID),
			Email:       output.User.Email,
			DisplayName: output.User.DisplayName,
			XP:          output.User.XP,
		},
		Tokens: TokensResponse{
			AccessToken:  output.Tokens.AccessToken,
			RefreshToken: output.Tokens.RefreshToken,
		},
	}

	httputil.JSON(w, http.StatusOK, response)
}

// handleAuthError trata erros do serviço de autenticação.
func handleAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrEmailAlreadyExists):
		httputil.Conflict(w, "Email already registered")
	case errors.Is(err, auth.ErrInvalidCredentials):
		httputil.Unauthorized(w, "Invalid email or password")
	case errors.Is(err, user.ErrInvalidEmail):
		httputil.BadRequest(w, "Invalid email format")
	case errors.Is(err, user.ErrPasswordTooShort):
		httputil.BadRequest(w, "Password must be at least 8 characters")
	case errors.Is(err, user.ErrDisplayNameTooShort):
		httputil.BadRequest(w, "Display name must be at least 3 characters")
	case errors.Is(err, user.ErrDisplayNameTooLong):
		httputil.BadRequest(w, "Display name must be at most 50 characters")
	default:
		httputil.InternalServerError(w, "An unexpected error occurred")
	}
}
