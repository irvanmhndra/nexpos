# nexpos tasks. `just --list` to see them.
#
# backend/     Rust API (axum + sqlx) — the rewrite, taking over route group by route group
# backend-go/  Go API — serves production until the cutover, owns the migrations
# frontend/    React web app

test_db := env_var_or_default("TEST_DATABASE_URL", "")

# Postgres for local dev (backend-go/docker-compose.yml, reads backend-go/.env)
dev-db:
    cd backend-go && docker compose up -d postgres

# Go API on :8080 with hot reload (reads backend-go/.env)
dev-go:
    cd backend-go && air

# Rust API on :8090 (reads the same POSTGRES_* env as the Go API)
dev-api:
    cd backend && SERVER_PORT=8090 cargo run --bin nexpos-server

# Web on :5173, talks to VITE_API_BASE_URL
dev-web:
    npm run dev --prefix frontend

# Rust: format check, clippy, tests
test-api:
    cd backend && cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test

# Go unit + integration tests. Integration uses testcontainers, or TEST_DATABASE_URL (a throwaway DB; its schema is reset)
test-go:
    cd backend-go && TEST_DATABASE_URL={{test_db}} go test -race -count=1 ./...

# Run the Go integration suite against a running server (default: the Rust API on :8090,
# the tests listed in backend/conformance-tests.txt). The server must use TEST_DATABASE_URL's database.
conformance target="http://127.0.0.1:8090" tests=`paste -sd'|' backend/conformance-tests.txt`:
    cd backend-go && TARGET_BASE_URL={{target}} TEST_DATABASE_URL={{test_db}} go test -count=1 -p 1 -run '{{tests}}' ./tests/integration/

# Lint everything
lint:
    cd backend && cargo clippy --all-targets -- -D warnings
    cd backend-go && golangci-lint run ./...
    npm run lint --prefix frontend

# Production build of the web app
build-web:
    npm run build --prefix frontend

migrate-up:
    make -C backend-go migrate-up

migrate-create name:
    make -C backend-go migrate-create name={{name}}
