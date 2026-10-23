# nexpos Rust API

The rewrite of `../backend-go`, taking over production route group by route group (strangler). Plan: `../backend-go/docs/rust-rewrite-plan.md`.

## Rules

- **Wire-compatible.** Same routes, the same JSON envelope (`crates/http/src/envelope.rs`), the same error codes (`crates/http/src/error.rs`), and the same status codes. The web app must not notice which backend answered.
- **One entry point.** With `GO_UPSTREAM_URL` set, any route Rust doesn't serve yet is forwarded to the Go API (`crates/http/src/proxy.rs`), so clients can talk to Rust only while route groups move over.
- **Same database, no migrations here.** The Go API owns the migrations until the cutover. This side only reads and writes the existing schema.
- **Done means conformant.** A route group is finished when the Go integration suite for it passes against this server:

  ```bash
  just conformance        # Rust server on :8090, tests from conformance-tests.txt
  ```

  Add a group's Go test names to `conformance-tests.txt` once it's ported, and CI keeps checking them.

## Layout

| Crate | What |
|---|---|
| `crates/domain` | pure rules: session tokens, bcrypt (Go's `$2a$` format), company codes |
| `crates/db` | sqlx queries on the Go schema |
| `crates/http` | routers, envelope, `AppError`, extractors (`JsonBody`, `Validator`, `AuthUser`, `ClientInfo`) that fail exactly like Echo + go-playground, the Go fallback proxy, middleware |
| `crates/server` | binary: env config (same `POSTGRES_*` and `JWT_*_EXPIRES_*` as Go, `SERVER_PORT` default 8090, `GO_UPSTREAM_URL`), pool, tracing, graceful shutdown |

## Ported

| Group | Routes | Conformance |
|---|---|---|
| health | `/health`, `/health/live`, `/health/ready` | `TestHealth_*` |
| auth | `POST /api/v1/auth/{login,register,refresh,logout}` + the `AuthUser` extractor | `TestAuth*`, `TestMigration036_*` |

Sessions and passwords are interchangeable with Go: a token or account made by one backend works on the other.

## Commands

```bash
just dev-api     # run on :8090
just test-api    # fmt --check, clippy -D warnings, tests (set DATABASE_URL to also test a live DB)
```
