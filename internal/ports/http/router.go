package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vinib1903/cineus-api/internal/app/auth"
	approom "github.com/vinib1903/cineus-api/internal/app/room"
	"github.com/vinib1903/cineus-api/internal/domain/user"
	infraauth "github.com/vinib1903/cineus-api/internal/infra/auth"
	"github.com/vinib1903/cineus-api/internal/ports/http/handlers"
)

// RouterConfig contém as dependências do router.
type RouterConfig struct {
	AuthService *auth.Service
	RoomService *approom.Service
	UserRepo    user.Repository
	JWTManager  *infraauth.JWTManager
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
	userHandler := handlers.NewUserHandler(cfg.UserRepo)
	roomHandler := handlers.NewRoomHandler(cfg.RoomService)

	// Rotas públicas
	r.Get("/health", healthHandler.Health)

	// Rotas da API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (públicas)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})

		// Room routes (algumas públicas)
		r.Route("/rooms", func(r chi.Router) {
			// Rotas públicas
			r.Get("/", roomHandler.ListPublic)
			r.Get("/{id}", roomHandler.GetByID)

			// Rotas protegidas
			r.Group(func(r chi.Router) {
				r.Use(AuthMiddleware(cfg.JWTManager))
				r.Post("/", roomHandler.Create)
				r.Get("/my", roomHandler.ListMy)
				r.Post("/join", roomHandler.JoinByCode)
				r.Delete("/{id}", roomHandler.Delete)
			})
		})

		// Rotas protegidas gerais
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(cfg.JWTManager))
			r.Get("/me", userHandler.Me)
		})
	})

	return r
}
