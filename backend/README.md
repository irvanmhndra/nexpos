# nexpos Rust API

The rewrite of `../backend-go`. It serves all 105 routes and runs in production next to the Go API, without traffic until the cutover below. Plan: `../backend-go/docs/rust-rewrite-plan.md`.

## Rules

- **Wire-compatible.** Same routes, the same JSON envelope (`crates/http/src/envelope.rs`), the same error codes (`crates/http/src/error.rs`), and the same status codes. The web app must not notice which backend answered.
- **Nothing left to forward.** `GO_UPSTREAM_URL` (`crates/http/src/proxy.rs`) forwarded routes not ported yet while groups moved over. It is unset everywhere now and goes away with the Go code.
- **Same database, no migrations here.** The Go API owns the migrations until the cutover. This side only reads and writes the existing schema.
- **Done means conformant.** The whole Go integration suite passes against this server alone:

  ```bash
  just conformance-local  # throwaway DB, Go + Rust servers, the whole suite, cleanup
  ```

  Rust runs with no Go upstream, so a route it doesn't serve fails its test. Go only answers the tests that compare the two byte for byte.

## Layout

| Crate | What |
|---|---|
| `crates/domain` | pure rules: session tokens, bcrypt (Go's `$2a$` format), company codes |
| `crates/db` | sqlx queries on the Go schema |
| `crates/http` | routers, envelope, `AppError`, extractors (`JsonBody`, `Validator`, `AuthUser`, `ClientInfo`, `QueryParams`, `RawId`) that fail exactly like Echo + go-playground + strconv, Go-style timestamps, the Go fallback proxy, middleware |
| `crates/server` | binary: env config (same `POSTGRES_*` and `JWT_*_EXPIRES_*` as Go, `SERVER_PORT` default 8090, `GO_UPSTREAM_URL`, `DB_TIMEZONE`), pool, tracing, graceful shutdown |

## Ported

| Group | Routes | Conformance |
|---|---|---|
| health | `/health`, `/health/live`, `/health/ready` | `TestHealth_*` |
| auth | `POST /api/v1/auth/{login,register,refresh,logout}` + the `AuthUser` extractor | `TestAuth*`, `TestMigration036_*` |
| branches | `/api/v1/branches`, `/api/v1/branches/{id}` (list, create, get, update, delete) | `TestBranch*` |
| customers | `/api/v1/customers`, `/api/v1/customers/{id}` | `TestCustomer*` |
| product categories | `/api/v1/product-categories`, `/all`, `/{id}` | `TestCategory*` |
| company settings | `GET`/`PUT /api/v1/company-settings` | `TestCompanySettings*` |
| reports | `/api/v1/reports/{summary,sales-trend,top-products,category-revenue,payment-methods,hourly-sales}` | `TestReport*` |
| products | `/api/v1/products`, `/lookup`, `/{id}` (writes run in one transaction) | `TestProductRead*`, `TestProductWrite*`, `TestProduct_*` |
| suppliers | `/api/v1/suppliers`, `/{id}` | `TestSupplier*` |
| users | `/api/v1/users`, `/{id}`, `/{id}/status` | `TestUser_*` |
| expenses | `/api/v1/expense-categories`, `/api/v1/expenses`, `/summary` | `TestExpense*` |
| promotions | `/api/v1/promotions`, `/{id}` | `TestPromotion*` |
| uploads | `POST /api/v1/uploads` (R2, signed with SigV4 in `storage.rs`) | `TestUpload*`; the real R2 path is unverified |
| orders, receipts | `/api/v1/orders`, `/preview`, `/{id}`, `/{id}/{confirm,payments,complete,cancel,void,refund,receipt}` | `TestOrder*`, `TestReceipt*`, `TestConcurrency_*` |
| inventory | `/api/v1/inventory`, `/stats`, `/adjust`, `/{variantId}/min-stock`, `/movements`, `/movements/stats` | `TestInventory*` |
| purchase orders | `/api/v1/purchase-orders`, `/{id}`, `/{id}/receive` | `TestPurchaseOrder*` |
| shifts | `/api/v1/shifts`, `/current`, `/{id}`, `/{id}/close` | `TestShift*` |
| stock opnames | `/api/v1/stock-opnames`, `/{id}`, `/{id}/items`, `/{id}/items/{itemId}`, `/{id}/{complete,cancel}` | `TestStockOpname*` |
| daily settlements | `/api/v1/daily-settlements`, `/report`, `/{id}`, `/{id}/items`, `/{id}/items/{itemId}`, `/{id}/finalize` | `TestDailySettlement*` |

`TestCompare_ReadsAreByteIdentical` and `TestCompare_MoneyCoreIsByteIdentical` also send the same reads and refused writes to both servers on the same data and compare status and raw bytes. Differences on purpose: internal errors say only "internal server error" here, a list sent as a non-list says "must be a sequence" where Go names its own type, and a wrong method is 405 where Echo 5.4 says 404. A successful write's response can also differ from Go's in `ordered_at`, `opened_at`, and `counted_at`: same instant, but Go prints its process clock with nanoseconds where this prints the database zone with microseconds.

Sessions and passwords are interchangeable with Go: a token or account made by one backend works on the other.

Timestamps are rendered like Go's `time.Time` in the database's TimeZone. sqlx starts sessions in UTC, so set `DB_TIMEZONE` (IANA name, default `UTC`, which is what production uses) to the database's zone; locally that is usually `Asia/Jakarta`. It sets each session's time zone, which reports depend on, and the zone timestamps are shown in.

Go quirks kept on purpose, because the web app sees them: `"<X> not found not found"` messages, pagination `count` always 0, an empty-string optional email failing validation, a refund leaving the order's `payment_status` as it was, a purchase order going to the session's branch rather than `branch_id` in the body, and only the complete endpoint saving a receipt (auto-completing doesn't).

## Cutover

Each push to `main` deploys this next to Go on `127.0.0.1:8090` (`docker-compose.prod.yml`, project `nexpos-rust`, sharing `~/nexpos/.env` and the database). Nginx still sends everything to Go.

1. On the server, check it against production data without affecting it: `curl -s 127.0.0.1:8090/health/ready`, then sign in and read a few lists with a real account.
2. In `/etc/nginx/sites-available/nexpos`, change the `api.nexpos.irvanmahendra.com` `proxy_pass` from `http://localhost:8080` to `http://localhost:8090`, then `nginx -t && systemctl reload nginx`. Sessions carry over, since both backends share them.
3. To go back, restore the port and reload. Both backends read and write the same schema, so nothing needs migrating either way.

The Go deploy keeps applying migrations until Go is removed. After two quiet weeks, delete `backend-go/` except what migrations and the test suite need, and move migrations here.

## Commands

```bash
just dev-api     # run on :8090
just test-api    # fmt --check, clippy -D warnings, tests (set DATABASE_URL to also test a live DB)
```
