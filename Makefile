# Shortcuts for common tasks. Run `make help` to list them.

.PHONY: help run tidy docker-up docker-down migrate-up migrate-down test

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  %-14s %s\n", $$1, $$2}'

run: ## Run the API locally (needs postgres+redis up)
	go run ./cmd/api

tidy: ## Sync go.mod/go.sum with imports
	go mod tidy

docker-up: ## Start postgres + redis + api in Docker
	docker compose up --build

docker-down: ## Stop and remove containers
	docker compose down

migrate-up: ## Apply DB migrations (requires golang-migrate CLI)
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down: ## Roll back the last migration
	migrate -path migrations -database "$$DATABASE_URL" down 1

test: ## Run tests
	go test ./...
