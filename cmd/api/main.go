package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/vinib1903/cineus-api/internal/config"
	"github.com/vinib1903/cineus-api/internal/infra/db"
	"github.com/vinib1903/cineus-api/internal/infra/repo"
)

func main() {
	cfg := config.Load()

	printLogo()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Connecting to database...")
	dbPool, err := db.NewPostgresPool(ctx, db.DefaultPostgresConfig(cfg.Database.URL))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connected successfully!")

	userRepo := repo.NewUserRepository(dbPool)
	roomRepo := repo.NewRoomRepository(dbPool)

	testRepository(ctx, userRepo)

	log.Printf("Repositories initialized: userRepo=%T, roomRepo=%T", userRepo, roomRepo)

	fmt.Printf("\n-> Server ready on port %s\n", cfg.Server.Port)
	fmt.Printf("-> Environment: %s\n", cfg.Server.Environment)

	waitForShutdown(cancel)
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

func waitForShutdown(cancel context.CancelFunc) {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Printf("\nReceived signal: %v. Shutting down...", sig)

	cancel()

	time.Sleep(1 * time.Second)
	log.Println("Server stopped gracefully.")
}

func testRepository(ctx context.Context, userRepo *repo.UserRepository) {
	log.Println("\n=== Testing User Repository ===")

	_, err := userRepo.GetByEmail(ctx, "test@example.com")
	if err != nil {
		log.Printf("GetByEmail (not found): %v ✓", err)
	}

	exists, err := userRepo.ExistsByEmail(ctx, "test@example.com")
	if err != nil {
		log.Printf("ExistsByEmail error: %v", err)
	} else {
		log.Printf("ExistsByEmail: %v ✓", exists)
	}

	log.Println("=== Repository Test Complete ===\n")
}
