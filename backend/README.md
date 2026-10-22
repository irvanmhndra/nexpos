# nexpos Rust API

The rewrite of `../backend-go`, taking over production route group by route group (strangler). Plan: `../backend-go/docs/rust-rewrite-plan.md`.

## Rules

- **Wire-compatible.** Same routes, the same JSON envelope (`crates/http/src/envelope.rs`), the same error codes (`crates/http/src/error.rs`), and the same status codes. The web app must not notice which backend answered.
- **Same database, no migrations here.** The Go API owns the migrations until the cutover. This side only reads and writes the existing schema.
- **Done means conformant.** A route group is finished when the Go integration suite for it passes against this server:

  ```bash
  just conformance        # Rust server on :8090, tests from conformance-tests.txt
  ```

  Add a group's Go test names to `conformance-tests.txt` once it's ported, and CI keeps checking them.

## Layout

| Crate | What |
|---|---|
| `crates/http` | routers, envelope, `AppError`, middleware (CORS, request id, timeout, panic → 500 envelope) |
| `crates/server` | binary: env config (same `POSTGRES_*` as Go, `SERVER_PORT` default 8090), pool, tracing, graceful shutdown |

`domain` (pure rules: money, tax, promotions, stock math) and `db` (sqlx queries) crates are added with the first route group that needs them.

## Commands

```bash
just dev-api     # run on :8090
just test-api    # fmt --check, clippy -D warnings, tests (set DATABASE_URL to also test a live DB)
```
