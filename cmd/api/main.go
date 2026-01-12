package main

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/vinib1903/cineus-api/internal/config"
)

func main() {
	// Carrega as configurações do .env
	cfg := config.Load()

	logo := `                                        
 ▄▄▄▄▄▄▄                 ▄▄▄  ▄▄▄       
███▀▀▀▀▀ ▀▀              ███  ███       
███      ██  ████▄ ▄█▀█▄ ███  ███ ▄█▀▀▀ 
███      ██  ██ ██ ██▄█▀ ███▄▄███ ▀███▄ 
▀███████ ██▄ ██ ██ ▀█▄▄▄ ▀██████▀ ▄▄▄█▀ 
                                        
                                        `
	color.Blue(logo)
	fmt.Printf("\n-> Listening on port %s...\n", cfg.Server.Port)
}
