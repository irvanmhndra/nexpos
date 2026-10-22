# nexpos tasks. `just --list` to see them.

test_db := env_var_or_default("TEST_DATABASE_URL", "")

# Postgres for local dev (backend/docker-compose.yml, reads backend/.env)
dev-db:
    cd backend && docker compose up -d postgres

# API on :8080 with hot reload (reads backend/.env)
dev-api:
    cd backend && air

# Web on :5173, talks to VITE_API_BASE_URL
dev-web:
    npm run dev --prefix frontend

# Backend unit + integration tests. Integration uses testcontainers, or TEST_DATABASE_URL (a throwaway DB; its schema is reset)
test-api:
    cd backend && TEST_DATABASE_URL={{test_db}} go test -race -count=1 ./...

# Lint both sides
lint:
    cd backend && golangci-lint run ./...
    npm run lint --prefix frontend

# Production build of the web app
build-web:
    npm run build --prefix frontend

# Everything CI checks before deploying
check: lint test-api build-web

migrate-up:
    make -C backend migrate-up

migrate-create name:
    make -C backend migrate-create name={{name}}
