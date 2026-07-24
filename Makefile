.PHONY: help dev test test-go test-nlp test-frontend lint build docker-up docker-down docker-rebuild clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Start development environment
	docker compose up -d
	@echo "Services starting..."
	@echo "  Web UI:     http://localhost:3000"
	@echo "  API:        http://localhost:8080"
	@echo "  Swagger:    http://localhost:8080/swagger/index.html"
	@echo "  NLP:        http://localhost:8000"
	@echo "  Postgres:   localhost:5432"
	@echo "  Redis:      localhost:6379"

test: test-go test-nlp ## Run all tests

test-go: ## Run Go tests
	go test ./... -v -count=1

test-go-cover: ## Run Go tests with coverage
	go test ./... -v -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-nlp: ## Run Python NLP tests
	cd nlp && python -m pytest tests/ -v

test-frontend: ## Run frontend tests
	cd frontend && npx vitest run

lint: ## Run linters
	@echo "Go vet..."
	go vet ./...
	@echo "Go fmt check..."
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting:" && gofmt -l . && exit 1)
	@echo "All checks passed!"

build: ## Build all Docker images
	docker compose build

docker-up: ## Start all services
	docker compose up -d

docker-down: ## Stop all services
	docker compose down

docker-rebuild: ## Rebuild and restart all services
	docker compose up -d --build

docker-logs: ## Tail logs from all services
	docker compose logs -f

docker-logs-api: ## Tail API logs
	docker compose logs -f api

docker-logs-nlp: ## Tail NLP logs
	docker compose logs -f nlp

clean: ## Remove build artifacts
	rm -f api.exe api coverage.out coverage.html
	rm -rf frontend/dist frontend/node_modules/.vite

db-import: ## Import OpenStreetMap data
	DATABASE_URL="postgres://postgres:postgres@localhost:5432/geosearch?sslmode=disable" \
		python3 scripts/ingest_osm.py

swagger: ## Regenerate Swagger docs
	swag init -g cmd/api/main.go -o docs
