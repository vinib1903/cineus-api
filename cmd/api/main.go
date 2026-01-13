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
	"github.com/vinib1903/cineus-api/internal/config"
	"github.com/vinib1903/cineus-api/internal/infra/db"
	"github.com/vinib1903/cineus-api/internal/infra/repo"
	httpport "github.com/vinib1903/cineus-api/internal/ports/http"
)

func main() {
	// Carrega as configurações do .env
	cfg := config.Load()

	// Exibe o logo
	printLogo()

	// Cria um contexto que será cancelado quando a aplicação receber sinal de término
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Conecta ao banco de dados
	log.Println("Connecting to database...")
	dbPool, err := db.NewPostgresPool(ctx, db.DefaultPostgresConfig(cfg.Database.URL))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connected successfully!")

	// Cria os repositórios
	_ = repo.NewUserRepository(dbPool)
	_ = repo.NewRoomRepository(dbPool)

	// Cria o router HTTP
	router := httpport.NewRouter(httpport.RouterConfig{})

	// Configura o servidor HTTP
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Inicia o servidor em uma goroutine
	go func() {
		log.Printf("Server starting on port %s...", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Exibe informações
	fmt.Printf("\n-> Server ready on http://localhost:%s\n", cfg.Server.Port)
	fmt.Printf("-> Health check: http://localhost:%s/health\n", cfg.Server.Port)
	fmt.Printf("-> Environment: %s\n\n", cfg.Server.Environment)

	// Aguarda sinal de término
	waitForShutdown(server, cancel)
}

// printLogo exibe o logo do CineUs.
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

// waitForShutdown aguarda sinais de término e faz shutdown gracioso.
func waitForShutdown(server *http.Server, cancel context.CancelFunc) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Printf("\nReceived signal: %v. Shutting down...", sig)

	// Cancela o contexto
	cancel()

	// Cria contexto com timeout para o shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Faz shutdown gracioso do servidor HTTP
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped gracefully.")
}
