package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/vinib1903/cineus-api/internal/app/auth"
	approom "github.com/vinib1903/cineus-api/internal/app/room"
	"github.com/vinib1903/cineus-api/internal/config"
	infraauth "github.com/vinib1903/cineus-api/internal/infra/auth"
	"github.com/vinib1903/cineus-api/internal/infra/db"
	"github.com/vinib1903/cineus-api/internal/infra/repo"
	httpport "github.com/vinib1903/cineus-api/internal/ports/http"
	"github.com/vinib1903/cineus-api/internal/ports/ws"
)

func main() {
	cfg := config.Load()
	printLogo()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database
	log.Println("Connecting to database...")
	dbPool, err := db.NewPostgresPool(ctx, db.DefaultPostgresConfig(cfg.Database.URL))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connected successfully!")

	// Repositories
	userRepo := repo.NewUserRepository(dbPool)
	roomRepo := repo.NewRoomRepository(dbPool)

	// Infrastructure services
	passwordHasher := infraauth.NewPasswordHasher(10)
	jwtManager := infraauth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	idGenerator := infraauth.NewIDGenerator()

	// Application services
	authService := auth.NewService(userRepo, passwordHasher, jwtManager, idGenerator)
	roomService := approom.NewService(roomRepo, idGenerator)

	// WebSocket hub
	wsHub := ws.NewHub()
	wsHandler := ws.NewHandler(wsHub, roomRepo)

	// HTTP Router
	router := httpport.NewRouter(httpport.RouterConfig{
		AuthService: authService,
		RoomService: roomService,
		UserRepo:    userRepo,
		JWTManager:  jwtManager,
		WSHandler:   wsHandler,
	})

	// HTTP Server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s...", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	fmt.Printf("\n-> Server ready on http://localhost:%s\n", cfg.Server.Port)
	fmt.Printf("-> Health check: http://localhost:%s/health\n", cfg.Server.Port)
	fmt.Printf("-> WebSocket: ws://localhost:%s/ws/room/{roomId}\n", cfg.Server.Port)
	fmt.Printf("-> Environment: %s\n\n", cfg.Server.Environment)

	waitForShutdown(server, cancel)
}

func printLogo() {
	logo := `                                        
 ▄▄▄▄▄▄▄                 ▄▄▄  ▄▄▄       
███▀▀▀▀▀ ▀▀              ███  ███       
███      ██  ████▄ ▄█▀█▄ ███  ███ ▄█▀▀▀ 
███      ██  ██ ██ ██▄█▀ ███▄▄███ ▀███▄ 
▀███████ ██▄ ██ ██ ▀█▄▄▄ ▀██████▀ ▄▄▄█▀ 
                                        `
	color.Blue(logo)
}

func waitForShutdown(server *http.Server, cancel context.CancelFunc) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Printf("\nReceived signal: %v. Shutting down...", sig)
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped gracefully.")
}
