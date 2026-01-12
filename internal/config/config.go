package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config armazena todas as configurações da aplicação.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Room     RoomConfig
}

// ServerConfig contém configurações do servidor HTTP.
type ServerConfig struct {
	Port        string
	Environment string
}

// DatabaseConfig contém configurações do PostgreSQL.
type DatabaseConfig struct {
	URL string
}

// RedisConfig contém configurações do Redis.
type RedisConfig struct {
	URL string
}

// JWTConfig contém configurações de autenticação.
type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// RoomConfig contém configurações das salas.
type RoomConfig struct {
	IdleTimeoutSeconds int
	MaxSeats           int
}

// Load carrega as configurações do arquivo .env e variáveis de ambiente.
// Retorna um ponteiro para Config preenchido.
func Load() *Config {
	// Tenta carregar o arquivo .env
	// Em produção, as variáveis vêm do ambiente (não do arquivo)
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	return &Config{
		Server: ServerConfig{
			Port:        getEnv("HTTP_PORT", "8080"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", ""),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", ""),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", ""),
			AccessTokenTTL:  getDurationEnv("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL: getDurationEnv("JWT_REFRESH_TOKEN_TTL", 168*time.Hour),
		},
		Room: RoomConfig{
			IdleTimeoutSeconds: getIntEnv("ROOM_IDLE_TIMEOUT_SECONDS", 120),
			MaxSeats:           getIntEnv("ROOM_MAX_SEATS", 16),
		},
	}
}

// getEnv busca uma variável de ambiente.
// Se não existir, retorna o valor padrão (defaultValue).
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getIntEnv busca uma variável de ambiente e converte para int.
// Se não existir ou for inválida, retorna o valor padrão.
func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Warning: %s is not a valid number, using default %d", key, defaultValue)
		return defaultValue
	}

	return intValue
}

// getDurationEnv busca uma variável de ambiente e converte para time.Duration.
// Aceita formatos como "15m", "1h", "168h".
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	// time.ParseDuration entende "15m", "1h30m", "24h", etc.
	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("Warning: %s is not a valid duration, using default value %v", key, defaultValue)
		return defaultValue
	}

	return duration
}
