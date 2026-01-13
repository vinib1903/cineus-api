.PHONY: help run build test clean docker-up docker-down migrate-up migrate-down migrate-create

# Variáveis
APP_NAME=cineus-api
BUILD_DIR=./bin
DATABASE_URL=postgres://cineus:cineus@localhost:5432/cineus?sslmode=disable

help: ## Mostra esta ajuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Roda a aplicação em modo desenvolvimento
	go run cmd/api/main.go

build: ## Compila a aplicação
	go build -o $(BUILD_DIR)/$(APP_NAME) cmd/api/main.go

test: ## Roda os testes
	go test -v ./...

test-coverage: ## Roda os testes com cobertura
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Limpa arquivos de build
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

docker-up: ## Sobe os containers (Postgres + Redis)
	docker-compose up -d

docker-down: ## Para os containers
	docker-compose down

docker-logs: ## Mostra logs dos containers
	docker-compose logs -f

migrate-up: ## Aplica todas as migrations
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down: ## Reverte a última migration
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-reset: ## Reverte TODAS as migrations
	migrate -path migrations -database "$(DATABASE_URL)" down -all

migrate-create: ## Cria uma nova migration (usar: make migrate-create name=nome_da_migration)
	migrate create -ext sql -dir migrations -seq $(name)

deps: ## Baixa as dependências
	go mod download
	go mod tidy
