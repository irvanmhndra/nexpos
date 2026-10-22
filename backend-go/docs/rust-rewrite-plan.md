# NexPOS backend rewrite: Go → Rust (React stays)

**Status:** proposal · **Scope:** `nexpos-api` only. `nexpos-web` keeps React; it only switches to a generated API client.
**Why nexpos rather than fnb:** one frontend instead of three, no SSE/QR customer flow, and a black-box HTTP integration suite that can be reused as a conformance test. Its money, stock, and tenancy rules were also just made explicit (transactions, decimal, composite FKs), so there is a precise spec to port.

---

## 1. Goals and non-goals

**Goals**
- Rust backend that is **wire-compatible** with today's API: same routes, JSON shapes, envelope, error codes, and status codes. `nexpos-web` must work unchanged against either backend.
- Same database, same schema, same session tokens during the transition, so Go and Rust can serve traffic side by side.
- Business rules ported 1:1, with tests proving it: tax, cash rounding, promotions, stock locking, refunds, tenancy.

**Non-goals**
- No new features during the rewrite.
- No schema redesign, and no frontend redesign.
- No "big bang" cutover.

**Be honest about the payoff.** Users get nothing new from this. Do it for type safety, learning Rust, lower memory, and a generated contract. Stop if it drags past the time box in §10.

---

## 2. Target stack

| Concern | Go today | Rust target |
|---|---|---|
| HTTP | Echo v5 | `axum` 0.8 + `tokio` |
| Middleware | Echo middleware | `tower-http` (cors, trace, request-id, compression, body limit, timeout) |
| DB | `sqlx` + `lib/pq` | `sqlx` (postgres, `macros`, `migrate`, offline mode with a committed `.sqlx/`) |
| Migrations | golang-migrate SQL files | **Keep golang-migrate** until cutover (§6); then `sqlx migrate` or keep golang-migrate |
| Money | `shopspring/decimal` | `rust_decimal` with the `serde-with-float` feature, so JSON stays numbers |
| Validation | go-playground/validator | `validator` derive, plus custom checks for `Decimal` ranges |
| Errors | `apperror` + `httputil.Response` | `enum AppError` + `impl IntoResponse` (§4.2) |
| Auth | opaque token, SHA-256 digest in `user_sessions` | same: `sha2` + `hex`, `rand` for tokens |
| Passwords | existing bcrypt hashes (confirm in `service/auth.go`) | `bcrypt` crate, same cost, so existing users log in unchanged |
| Storage (R2) | S3 client | `aws-sdk-s3` with the custom R2 endpoint (account-id host) |
| Logging | slog | `tracing` + `tracing-subscriber` (JSON), request-id span |
| OpenAPI | none | `utoipa` + `utoipa-axum`, spec generated and committed |
| Tests | testify mocks + testcontainers | pure-domain unit tests, `#[sqlx::test]` (per-test DB), conformance suite (§7) |
| Lint / security | golangci-lint, gosec, govulncheck | `cargo fmt`, `clippy -D warnings`, `cargo-deny`, `cargo-audit` |
| Container | multi-stage Go build | `cargo-chef` cache layers → static musl or distroless image |

---

## 3. Repository layout

A Cargo workspace inside the existing repo (`nexpos-api/rust/`), or a new `nexpos-rs` repo. Same-repo is recommended while both backends live side by side:

```
nexpos-api/
├── (existing Go code — untouched until cutover)
├── migrations/                  # stays the single schema source of truth
└── rust/
    ├── Cargo.toml               # [workspace]
    ├── crates/
    │   ├── domain/              # pure logic, no I/O: money, tax, rounding, promotions,
    │   │                        # order state machine, stock math, errors
    │   ├── db/                  # sqlx queries per domain, transactions, row locks
    │   ├── http/                # axum routers per domain, extractors, AppError, OpenAPI
    │   └── server/              # bin: config, pool, wiring, graceful shutdown
    ├── .sqlx/                   # sqlx offline query cache (committed)
    ├── openapi.json             # generated, committed, CI checks it is fresh
    └── Dockerfile
```

Every crate is split by domain module: `auth`, `users`, `branches`, `customers`, `catalog` (categories, products, variants), `orders`, `payments`, `promotions`, `inventory` (stocks, movements), `purchasing` (suppliers, POs), `shifts`, `expenses`, `stock_opname`, `settlements`, `receipts`, `reports`, `settings`, `uploads`.

The split keeps compile times sane: editing a handler recompiles `http` and `server`, not `domain`.

---

## 4. Porting the Go patterns

### 4.1 Layers
- **Handler → axum handler.** It extracts `AuthUser`, `Path`, and `Json<T>`, validates, calls the service, and returns `Result<Json<Envelope<T>>, AppError>`.
- **Service → plain functions/structs** that take `&mut PgConnection` or `&mut Transaction<'_, Postgres>`.
- **Repository → sqlx query functions** in `db::<domain>`.

Drop the mock-heavy style. Pure rules move to `domain` (unit-tested with no DB). Anything touching SQL is tested against a real database with `#[sqlx::test]`, which gives each test a fresh migrated database.

### 4.2 Errors and the envelope (must be identical)
```rust
pub enum AppError {
    BadRequest(String),                 // 400 BAD_REQUEST
    Validation(String, Vec<FieldError>), // 422 VALIDATION_ERROR
    Unauthorized(String),               // 401 UNAUTHORIZED
    TokenExpired,                       // 401 TOKEN_EXPIRED
    Forbidden(String),                  // 403 FORBIDDEN
    NotFound(String),                   // 404 NOT_FOUND
    Conflict(String),                   // 409 CONFLICT
    Internal(anyhow::Error),            // 500 INTERNAL_ERROR, detail only in logs
}
```
- Map sqlx database errors the way `httputil` does today: `23505` → Conflict, `23503` → Validation/NotFound, serialization/deadlock → Conflict.
- Copy the exact `{ success, message, data, error_code, errors }` shape from `pkg/httputil`, and snapshot-test it.

### 4.3 Transactions and locking
Go carries the transaction in `context`. Rust passes it explicitly, which is clearer:

```rust
let mut tx = pool.begin().await?;
let order = db::orders::get_for_update(&mut tx, company_id, id).await?;   // SELECT … FOR UPDATE
domain::orders::complete(&order)?;                                        // state machine
db::stock::apply_sale(&mut tx, &order).await?;  // lock stock rows in variant-id order, clamp at 0
tx.commit().await?;
```

Port these rules exactly (they are the reason for the rewrite's tests):
- Order state transitions lock the order row, and refunds lock the payment row.
- Stock changes use `LockForUpdate`: insert at zero if missing, then `SELECT … FOR UPDATE`. Lock in **ascending variant id** to avoid deadlocks. A sale that exceeds stock floors it at 0 and logs a warning.
- Stock deduction on complete and auto-complete, and restore on void, happen **in the same transaction** as the state change.
- PO receive and stock-opname complete lock their header row and every stock row.

### 4.4 Money
- `Decimal` everywhere money or rates appear. Display-only percentages (margin, period changes) stay `f64`.
- `line_amounts(unit_price, qty, discount, settings) -> (gross, subtotal, tax)`, with tax rounded to 2 dp **per line**.
- `order_grand_total(gross, discount, tax, settings)`: add tax **only for exclusive pricing**, then cash rounding (`round_to_nearest`, halves up).
- Percentage promotions are rounded to 2 dp and capped by `max_discount`. Fixed promotions are capped at the subtotal. The best promotion wins.
- These are `domain` functions with table-driven tests, copied from the Go unit tests and the inclusive/exclusive tax cases.

### 4.5 Auth and tenancy
- Tokens: 32 random bytes, hex encoded. The DB stores only the `sha256` hex digest (`access_token_hash`, `refresh_token_hash`).
- Refresh is single-use: a conditional revoke plus insert in one transaction (`Rotate`).
- **The same table and the same hashing mean a token issued by Go works on Rust and vice versa.** This is what makes the side-by-side run in §6 possible.
- `AuthUser` extractor: bearer → digest → session (not revoked, not expired) → `{ user_id, company_id, branch_id }`, then update `last_used_at` best-effort.
- Every variant/product ID from a request is looked up **scoped to `company_id`**, like `GetByIDForCompany`. The composite FKs from migration 000037 are the backstop.

---

## 5. Contract first, then code

1. **Write down today's contract** before porting anything:
   - hand-author `openapi.yaml` for the Go API from `internal/router/router.go`, the DTOs, and `docs/api_documentation.md`;
   - record golden responses for representative calls from the current integration tests.
2. The Rust server generates its own spec with `utoipa`.
3. CI diffs the generated spec against the frozen contract. Any change must be intentional.
4. `nexpos-web` generates types and a client from the spec (`openapi-typescript` + a thin fetch wrapper) into `src/api/generated/`, then replaces hand-written types in `src/types/*.ts` screen by screen. This works with the Go backend too, so it can start in Phase 0.

---

## 6. Migration strategy: strangler, route group by route group

Both services run against the same Postgres. nginx on `api.nexpos.irvanmahendra.com` sends migrated route groups to Rust (`127.0.0.1:8090`) and everything else to Go (`127.0.0.1:8080`):

```nginx
location /api/v1/customers { proxy_pass http://127.0.0.1:8090; }
location /api/v1/          { proxy_pass http://127.0.0.1:8080; }   # default: Go
```

Rules while both run:
- **The Go repo keeps owning migrations** (the golang-migrate step in the Go deploy). Rust never migrates. Rust compiles its queries against the migrated schema in sqlx offline mode.
- Any schema change must be backward compatible for both services (expand, then contract).
- Rollback of a route group is a one-line nginx change plus a reload.

---

## 7. Test plan

| Layer | What | Tool |
|---|---|---|
| Domain | tax, rounding, promotions, state machine, stock math. Port every case from `internal/service/*_test.go` | `cargo test` / `nextest` |
| DB | each query, locks, transactions, tenancy scoping | `#[sqlx::test]` |
| **Conformance** | the existing Go HTTP integration suite, pointed at a base URL, run against **both** servers | Go `tests/integration`, made target-agnostic |
| Concurrency | port `concurrency_test.go`: parallel completes, parallel sales, parallel refunds, parallel adjustments, rollback on failed create | conformance suite |
| Contract | spec freshness, JSON snapshots of envelopes and errors | CI + `insta` |

**Conformance is the key idea.** `tests/integration` already talks HTTP. Make `testutil` accept a `TARGET_BASE_URL`: skip starting the Go app and point requests at the running Rust server with the same database. A route group is "done" when the suite for it passes against Rust.

---

## 8. Phases

| # | Phase | Deliverable | Exit criteria |
|---|---|---|---|
| 0 | Contract & harness | `openapi.yaml`, target-agnostic integration suite, generated TS client in nexpos-web | Suite green against Go via `TARGET_BASE_URL` |
| 1 | Skeleton | workspace, config, pool, `/health`, AppError envelope, tracing, request-id, CORS, CI (fmt, clippy, nextest, deny/audit, sqlx offline check, spec check), Docker, deploy workflow on port 8090 | Rust `/health` live in prod behind nginx, receiving no API traffic yet |
| 2 | Auth & RBAC | login, register, refresh, logout, `AuthUser`, permissions | Auth conformance green; tokens interchangeable across Go and Rust |
| 3 | Read-mostly domains | branches, customers, categories, products/variants (read + lookup), company settings, reports | Routed to Rust in prod for a week with no error-rate change |
| 4 | CRUD domains | products write + R2 uploads + orphan cleanup, promotions, suppliers, expenses, users | Conformance green; routed |
| 5 | Money & stock core | orders (create, preview, update, confirm, pay, complete, cancel, void, refund), receipts, inventory adjust / min-stock / movements, POs + receive, stock opname, shifts, daily settlements | Full suite **including concurrency tests** green on Rust; shadow period passed |
| 6 | Cutover | nginx sends all of `/api/v1` to Rust; Go kept deployable for 2 weeks | Two quiet weeks, then delete the Go code and move migrations ownership |

**Before routing orders:** run a shadow week.
- Mirror a sample of `POST /orders/preview` to Rust (it is read-only) and log any response diff.
- Nothing else in the money path is mirrored, because writes cannot be duplicated.

---

## 9. CI/CD

- Rust jobs are filtered to `rust/**` paths; Go jobs to their own paths.
- Required checks: `fmt`, `clippy -D warnings`, `nextest`, `cargo deny check`, `cargo sqlx prepare --check`, OpenAPI freshness, conformance suite against a Rust container.
- Deploy the same way as today: build → GHCR → ssh. Keep the migrate-before-swap order, and remember **migrations still run from the Go deploy** until Phase 6.
- Image: `cargo-chef` for dependency caching, a musl static binary, and a `FROM scratch`/distroless runtime. Expect roughly 15–30 MB, against the current Go image plus the migrate binary.

---

## 10. Risks and time box

| Risk | Mitigation |
|---|---|
| Rules drift while two implementations live | Conformance suite runs against both on every PR; the Go side is frozen except for bug fixes, which are ported the same week |
| Decimal/JSON mismatch (`100` vs `100.0` vs `"100.00"`) | `serde-with-float`, plus snapshot tests on money fields; `nexpos-web` already parses numbers |
| Compile times slow the loop | workspace split (§3), `mold`/`lld` linker, `cargo-nextest`, sqlx offline mode |
| Rust learning curve on async + lifetimes | Pass `&mut Transaction` explicitly; avoid generic repository traits; keep `domain` pure and synchronous |
| Dual-run schema changes | expand/contract only; Go owns migrations until cutover |
| Time sink with no user value | **Time box:** if Phase 3 isn't in prod within ~6 weeks of part-time work, stop and keep Go. The contract, generated client, and conformance suite from Phase 0 are still worth having. |

Rough effort, solo part-time: Phase 0 ≈ 1 wk, 1 ≈ 1 wk, 2 ≈ 1 wk, 3 ≈ 1–2 wk, 4 ≈ 2 wk, 5 ≈ 3–4 wk, 6 ≈ 2 wk of watching. About **10–13 weeks**.

---

## 11. Definition of done

- 100% of `/api/v1` is served by Rust, and the Go code plus its deploy job are removed.
- The conformance suite, concurrency tests included, is green against Rust in CI.
- `nexpos-web` builds only against generated types; no hand-written API types remain.
- `openapi.json` is generated, committed, and checked in CI.
- p95 latency and error rate are at least as good as Go over two weeks. Memory use is recorded for the before/after.
