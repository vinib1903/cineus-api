package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vinib1903/cineus-api/internal/ports/http/handlers"
)

// RouterConfig contém as dependências do router.
type RouterConfig struct {
	// Aqui vamos adicionar os serviços/repositórios depois
}

// NewRouter cria e configura o router HTTP.
func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	// Middlewares globais
	r.Use(middleware.RequestID) // Adiciona ID único a cada requisição
	r.Use(middleware.RealIP)    // Obtém o IP real do cliente
	r.Use(Logger)               // Nosso logger customizado
	r.Use(Recoverer)            // Nosso recoverer customizado
	r.Use(CORS)                 // Permite requisições cross-origin

	// Handlers
	healthHandler := handlers.NewHealthHandler()

	// Rotas públicas (sem autenticação)
	r.Get("/health", healthHandler.Health)

	// Rotas da API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (públicas)
		r.Route("/auth", func(r chi.Router) {
			// TODO: adicionar handlers de auth
			// r.Post("/register", authHandler.Register)
			// r.Post("/login", authHandler.Login)
		})

		// Room routes (algumas públicas, outras autenticadas)
		r.Route("/rooms", func(r chi.Router) {
			// TODO: adicionar handlers de rooms
			// r.Get("/", roomHandler.ListPublic)
			// r.Post("/", roomHandler.Create)
		})
	})

	return r
}
