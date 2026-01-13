package handlers

import (
	"net/http"

	"github.com/vinib1903/cineus-api/internal/ports/http/httputil"
)

// HealthHandler gerencia as rotas de health check.
type HealthHandler struct{}

// NewHealthHandler cria uma nova instância do handler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthResponse é a resposta do health check.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// Health retorna o status da aplicação.
// GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:  "healthy",
		Version: "0.1.0",
	}

	httputil.JSON(w, http.StatusOK, response)
}
