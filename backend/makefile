.PHONY: all build run test lint mocks

include .env
export

DATABASE_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)
MIGRATE=migrate -path migrations -database "$(DATABASE_URL)"

# ==================== Build & Run ====================

build:
	go build -o bin/api ./cmd/api/main.go

run:
	go run cmd/api/main.go

dev:
	air

# ==================== Docker ====================

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# ==================== Database ====================

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-version:
	$(MIGRATE) version

migrate-force:
	$(MIGRATE) force $(version)

# ==================== Testing ====================

test: test-unit test-integration

test-unit:
	@echo "Running unit tests..."
	go test -v -race -short ./internal/...

test-integration:
	@echo "Running integration tests..."
	go test -v -race ./tests/integration/...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage-func:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# ==================== Mocks ====================

mocks:
	@echo "Generating mocks..."
	mockery

mocks-clean:
	rm -rf internal/repository/mocks
	rm -rf internal/client/mocks
	rm -rf internal/event/mocks

# ==================== Linting ====================

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

# ==================== Clean ====================

clean:
	rm -rf bin coverage.out coverage.html

# ==================== Help ====================

help:
	@echo "Available commands:"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build              Build the application"
	@echo "  make run                Run the application"
	@echo "  make dev                Run with hot reload (requires air)"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up          Start all containers"
	@echo "  make docker-down        Stop all containers"
	@echo "  make docker-logs        View container logs"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up         Run all migrations"
	@echo "  make migrate-down       Rollback last migration"
	@echo "  make migrate-create name=xxx  Create new migration"
	@echo ""
	@echo "Testing:"
	@echo "  make test               Run all tests"
	@echo "  make test-unit          Run unit tests only"
	@echo "  make test-integration   Run integration tests only"
	@echo "  make test-coverage      Generate coverage report"
	@echo ""
	@echo "Mocks:"
	@echo "  make mocks              Generate mocks (requires mockery)"
	@echo "  make mocks-clean        Remove generated mocks"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint               Run linter"
	@echo "  make lint-fix           Run linter and fix issues"