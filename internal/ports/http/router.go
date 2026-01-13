package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vinib1903/cineus-api/internal/app/auth"
	"github.com/vinib1903/cineus-api/internal/ports/http/handlers"
)

// RouterConfig contém as dependências do router.
type RouterConfig struct {
	AuthService *auth.Service
}

// NewRouter cria e configura o router HTTP.
func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	// Middlewares globais
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(Logger)
	r.Use(Recoverer)
	r.Use(CORS)

	// Handlers
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(cfg.AuthService)

	// Rotas públicas
	r.Get("/health", healthHandler.Health)

	// Rotas da API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})
	})

	return r
}
