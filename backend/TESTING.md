# Testing Guide

## Testing Strategy

This project uses a layered testing approach:

```
┌──────────────────────────────────────────────────┐
│  Integration Tests (tests/integration/)          │
│  Real HTTP requests → full stack → real database │
│  Tests: auth, customer, product, order, health   │
└──────────────────────┬───────────────────────────┘
                       │
┌──────────────────────┴───────────────────────────┐
│  Service Unit Tests (internal/service/*_test.go) │
│  Mock repositories, test business logic          │
│  Tests: all services                             │
└──────────────────────────────────────────────────┘
```

### Why this structure?

| Layer | What it tests | Mock boundary | Value |
|-------|--------------|---------------|-------|
| **Service unit tests** | Business logic, validation, error handling | Repository interfaces | **High** — this is where the real logic lives |
| **Integration tests** | Full request → response → database flow, including side effects | Nothing (real stack) | **High** — verifies everything works together |

### Key design decisions

1. **Services use repository interfaces** — enables mocking for service unit tests
2. **Integration tests live in `tests/integration/`** — a separate package is required because integration tests need to import `internal/app` to boot the real application, but `internal/app` already imports `internal/handler`. Placing integration tests inside `internal/handler/` would create a circular import (`handler_test → app → handler`). The separate package can import anything without this issue, giving it access to the real app, database, middleware, and route registration
3. **Integration tests share setup via `TestMain`** in `setup_test.go` — database, app, and test server are initialized once
4. **Integration tests use DB assertions for side effects** — when an operation produces database changes not visible in the API response (e.g. stock movements, payment status transitions), we query the database directly to verify

## Project Test Structure

```
internal/
  service/
    branch.go
    branch_test.go        ← service unit test (mocks repository)
    customer.go
    customer_test.go
    ...
  repository/
    mocks/                ← repository mocks (for service tests)
      branch.go
      customer.go
      ...
tests/
  testutil/
    db.go               ← PostgreSQL testcontainer + migrations + truncate helpers
    http.go             ← test HTTP client (GET, POST, PUT, DELETE)
    fixtures.go         ← test data factories (company, branch, user, product, stock, etc.)
    setup.go            ← TestEnv: boots full app wiring
  integration/
    setup_test.go         ← TestMain: start containers, boot app, teardown
    helpers_test.go       ← doPost/doGet/doPut/doDelete/decodeResponse/registerTestUser
    auth_test.go          ← full-stack integration tests
    customer_test.go
    order_test.go         ← includes DB assertions for stock/payment
    receipt_test.go       ← receipt snapshot tests (Postgres JSONB)
    ...
```

## Running Tests

### Unit tests only (fast, no database needed)
```bash
go test -short ./internal/...
```

### Unit tests with race detection
```bash
go test -race -short ./internal/...
```

### Integration tests (no external database needed)
Testcontainers automatically starts a PostgreSQL container during the test run.
Requires Docker to be running.

```bash
go test -v -race ./tests/integration/...
```

### All tests
```bash
go test ./...
```

### Specific package
```bash
go test -v ./internal/service/...
```

### Specific test
```bash
go test -v ./internal/service/ -run TestCustomerService_Create
```

### Coverage report
```bash
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out -o coverage.html
```

## Writing Tests

### Service unit test pattern

Service tests mock repository dependencies using `testify/mock`:

```go
package service

func setupCustomerTest(t *testing.T) (*CustomerService, *repoMocks.MockCustomerRepository) {
    t.Helper()
    mockRepo := repoMocks.NewMockCustomerRepository(t)
    svc := NewCustomerService(mockRepo)
    return svc, mockRepo
}

func TestCustomerService_Create_Success(t *testing.T) {
    svc, mockRepo := setupCustomerTest(t)

    // Arrange
    mockRepo.EXPECT().GetByCode(mock.Anything, int64(1), "CUST-001").
        Return(nil, nil)
    mockRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Customer")).
        Return(nil)

    // Act
    result, err := svc.Create(ctx, 1, dto.CreateCustomerRequest{
        Code: "CUST-001",
        Name: "Test",
    })

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "CUST-001", result.Code)
}
```

### Integration test pattern

Integration tests use the real HTTP stack and database. All protected routes require authentication via `registerTestUser()`, which registers a user via the API and returns an `AuthContext` with tokens and IDs:

```go
package integration_test

func TestCustomer_CreateAndGet(t *testing.T) {
    cleanupDatabase(t)
    auth := registerTestUser(t)

    // Create — doPost wraps testEnv.Server.POST and asserts no transport error
    resp := doPost(t, "/api/v1/customers", map[string]interface{}{
        "code":  "CUST-001",
        "name":  "John Customer",
        "phone": "+1234567890",
    }, auth.Token)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    // decodeResponse parses the standard API envelope
    r := decodeResponse(t, resp)
    var data map[string]interface{}
    require.NoError(t, json.Unmarshal(r.Data, &data))
    customerID := int64(data["id"].(float64))

    // Get
    resp = doGet(t, fmt.Sprintf("/api/v1/customers/%d", customerID), auth.Token)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

### Integration test with order setup and DB assertions

Order tests require stock to be seeded before creating orders (the API validates stock availability). Use `setupOrderTest()` which creates all prerequisites via the API and seeds stock directly in the database:

```go
type orderTestData struct {
    auth       *authContext
    customerID int64
    categoryID int64
    productID  int64
    variantID  int64
}

func setupOrderTest(t *testing.T) *orderTestData {
    cleanupDatabase(t)
    auth := registerTestUser(t)

    // Create customer, category, product via API
    resp := doPost(t, "/api/v1/customers", map[string]interface{}{"code": "ORD-CUST", "name": "Order Test Customer"}, auth.Token)
    r := decodeResponse(t, resp)
    var custData map[string]interface{}
    require.NoError(t, json.Unmarshal(r.Data, &custData))
    customerID := int64(custData["id"].(float64))

    categoryID := createTestCategory(t, auth.Token)
    productID, variantID := createTestProduct(t, auth.Token, categoryID, "Test Product", "SKU-001", 100.00, 50.00)

    // Seed stock via fixtures (required — order creation validates stock availability)
    err := testEnv.Fixtures.CreateStock(context.Background(), variantID, auth.BranchID, 100, 0)
    require.NoError(t, err)

    return &orderTestData{auth: auth, customerID: customerID, productID: productID, variantID: variantID}
}
```

**Important: Auto-complete counter orders.** By default, `CompanySettings.AutoCompleteCounterOrders` is `true`. When a counter-type order is fully paid, the system auto-completes it. To test the explicit confirm → pay → complete flow, use `"delivery"` as the fulfillment type:

```go
// Use "delivery" to test explicit complete step (counter orders auto-complete on payment)
orderBody := map[string]interface{}{
    "fulfillment_type": "delivery",
    "items": []map[string]interface{}{...},
}
```

For DB assertions (stock movements, payment status), query the database directly:

```go
// DB assertion: verify stock was deducted
var stockQty int
err = testEnv.DB.QueryRowContext(ctx,
    "SELECT quantity FROM stocks WHERE product_variant_id = $1 AND branch_id = $2",
    variantID, auth.BranchID,
).Scan(&stockQty)
require.NoError(t, err)
assert.Equal(t, 9, stockQty) // started at 10, ordered 1
```

### When to use DB assertions

Most integration tests only need to verify via HTTP responses (POST then GET back). Use direct DB assertions when:

| Scenario | Why DB assertion is needed |
|----------|--------------------------|
| **Stock deduction on order complete** | Stock changes commit in the same transaction as the completion but are not returned in the order response |
| **Stock restoration on order void** | Need to verify IN movements were created and stock quantity restored |
| **Payment refund status** | Payment status transitions happen in a separate table, verify the actual DB state |
| **No stock change on non-completed void** | Verify that voiding a confirmed (not completed) order does NOT create stock movements |

### When NOT to use DB assertions

- CRUD operations — verify through GET after POST/PUT/DELETE
- Pagination, filtering, search — verify through the list API response
- Validation errors — verify HTTP status code is sufficient
- Authentication flows — verify token and response status

## Naming Convention

Format: `Test<Type>_<Method>_<Scenario>`

```go
// Service tests
func TestCustomerService_Create_Success(t *testing.T)
func TestCustomerService_Create_CodeExists(t *testing.T)
func TestCustomerService_Create_RepoError(t *testing.T)

// Integration tests (flow-based naming)
func TestOrderFlow_CreateAndGet(t *testing.T)
func TestOrderFlow_CreateConfirmPayComplete(t *testing.T)
func TestOrderFlow_VoidCompleted(t *testing.T)
func TestOrder_ValidationErrors(t *testing.T)
```

## What to test at each layer

### Service tests — focus on:
- Business logic and validation rules
- Error handling (not found, conflict, repo errors)
- Edge cases (nil values, boundary conditions)
- State transitions (order lifecycle)

### Integration tests — focus on:
- End-to-end flows (create → get → update → delete)
- Database constraints (unique, foreign keys)
- Pagination and search
- Authentication flows
- Side effects via DB assertions (stock movements, payment status)

## How Integration Tests Work End-to-End

This section explains the full lifecycle of an integration test run: from `go test ./...` to teardown.

### 1. Test entry point — `TestMain`

`tests/integration/setup_test.go` declares `TestMain(m *testing.M)`. Go calls this before any test in the package runs. It is the single place where shared infrastructure (database, app, HTTP server) is initialized and later torn down:

```
go test ./tests/integration/...
    └── TestMain(m)
          ├── NewTestDB(ctx)       → starts PostgreSQL container
          ├── RunMigrations()      → applies all SQL migrations
          ├── app.New(cfg)         → boots the full application
          ├── NewTestServer(echo)  → wraps echo in httptest.Server
          └── m.Run()              → runs all Test* functions
                ├── TestAuth_Login
                ├── TestOrderFlow_CreateConfirmPayComplete
                ├── TestReceipt_GetAfterComplete
                └── ...
          └── env.Cleanup()        → teardown (containers + connections)
```

The infrastructure is created **once** for the entire package and shared across all tests, which keeps the test suite fast — container startup happens once, not per test.

### 2. Testcontainers — real databases, no manual setup

[testcontainers-go](https://github.com/testcontainers/testcontainers-go) starts real Docker containers programmatically from Go test code. No external database or Docker Compose file is needed — `go test` is the only command required.

**PostgreSQL container** (`tests/testutil/db.go`):

```go
pgContainer, err := postgres.Run(ctx, "postgres:18-alpine",
    postgres.WithDatabase("pos_test_db"),
    postgres.WithUsername("pos_test_user"),
    postgres.WithPassword("pos_test_password"),
    testcontainers.WithWaitStrategy(
        wait.ForLog("database system is ready to accept connections").
            WithOccurrence(2).
            WithStartupTimeout(60*time.Second),
    ),
)
```

- `postgres.Run` pulls the image (first run only; cached after) and starts a container.
- `WithWaitStrategy` blocks until the log line appears twice — this is the PostgreSQL readiness signal. The test will not proceed until the database is truly ready to accept connections, preventing flaky "connection refused" errors.
- `pgContainer.ConnectionString(ctx, "sslmode=disable")` returns a dynamic DSN like `postgres://pos_test_user:pos_test_password@localhost:49821/pos_test_db?sslmode=disable`. The port is random (Docker assigns it), so there are no port conflicts.

The container is terminated in `env.Cleanup()`, which is called after `m.Run()` returns.

### 3. golang-migrate — schema management

[golang-migrate](https://github.com/golang-migrate/migrate) applies SQL migration files from `migrations/` in sequence. It uses a `schema_migrations` table to track which migrations have already run.

**How `RunMigrations()` works** (`tests/testutil/db.go`):

```go
// 1. Drop and recreate the public schema for a clean slate
_, err := t.DB.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")

// 2. Find the migrations/ directory by walking up from the working directory
migrationsDir, err := findMigrationsDir()
// walks up until go.mod is found, returns <root>/migrations

// 3. Create a migrator targeting the test database
m, err := migrate.New("file://"+migrationsDir, t.DSN)

// 4. Apply all UP migrations in sequence
if err := m.Up(); err != nil && err != migrate.ErrNoChange {
    return fmt.Errorf("run migrations: %w", err)
}
```

`findMigrationsDir()` solves the problem that `go test` sets the working directory to the package being tested (e.g. `tests/integration/`), not the project root. It walks up directories until it finds `go.mod`, then appends `migrations/`. This means migration files are always found regardless of which package runs the tests.

**Migration files** (`migrations/`) follow the naming convention:
```
000001_create_companies.up.sql
000001_create_companies.down.sql
000002_create_branches.up.sql
...
```

`m.Up()` applies all unapplied files in numeric order. Each migration runs in a transaction — if any file fails, the migration halts and reports which file failed.

### 4. App boot — `app.New(cfg)`

The test passes a `*config.Config` built entirely from testcontainer outputs — no environment variables needed:

```go
cfg := &config.Config{
    Postgres: config.PostgresConfig{
        DSNOverride: tdb.DSN,  // e.g. "postgres://...@localhost:49821/..."
    },
    JWT: config.JWTConfig{
        Secret: "test-secret-key-for-integration-tests",
        ...
    },
}
a, err := app.New(cfg)
```

`app.New` initializes the full stack in the same way production does: Postgres connection pool, repository layer, service layer, handler layer, Echo routes, and middleware. The test app is **identical** to production — not a stripped-down version.

### 5. HTTP test server — `httptest.Server`

```go
server := NewTestServer(env.App.Echo())
// internally: httptest.NewServer(handler) → real TCP listener on a random port
```

`httptest.NewServer` starts a real TCP listener on a random local port. Tests make real `http.Client` requests to this server — they go through the full Echo middleware stack (auth middleware, request logger, CORS, recover). This is different from using `httptest.NewRecorder`, which bypasses the HTTP layer and calls handlers directly.

### 6. Per-test isolation — `cleanupDatabase(t)`

Tests share the same containers throughout the package run. To prevent test pollution, each test calls `cleanupDatabase(t)` at the start:

```go
func cleanupDatabase(t *testing.T) {
    t.Helper()
    if err := testEnv.TestDB.TruncateAllTables(); err != nil {
        t.Fatalf("Failed to truncate tables: %v", err)
    }
}
```

`TruncateAllTables()` truncates all application tables (including `receipts`) with `RESTART IDENTITY CASCADE` (resets auto-increment sequences too) and re-seeds system data (roles, permissions, role_permissions). This brings the database back to a known state in milliseconds — much faster than restarting containers or re-running migrations.

### 7. End-to-end request flow in a test

When a test calls `doPost(t, "/api/v1/orders", body, auth.Token)`:

```
doPost
  └── http.NewRequest(POST, "http://127.0.0.1:<port>/api/v1/orders", body)
        └── httptest.Server (TCP listener)
              └── Echo router
                    ├── middleware.RequestLogger
                    ├── middleware.Recover
                    ├── middleware.CORS
                    ├── middleware.RequestID
                    └── authmiddleware.Auth   ← validates JWT, sets company_id/user_id/branch_id in context
                          └── OrderHandler.Create
                                └── OrderService.Create
                                      └── OrderRepository.Create  ← real SQL to testcontainer PostgreSQL
```

The response comes back as a real HTTP response — status code, headers, JSON body — exactly as a frontend would receive it.

### Full lifecycle summary

```
go test ./tests/integration/...

1. TestMain starts
   ├── docker pull postgres:18-alpine  (first run only)
   ├── docker run postgres:18-alpine   → port 49821
   ├── wait for "database system is ready" × 2
   ├── DROP SCHEMA public CASCADE; CREATE SCHEMA public
   ├── golang-migrate: apply all migrations to port 49821
   ├── app.New(cfg)                    → full app boot (repos, services, handlers, routes)
   └── httptest.NewServer(echo)        → TCP listener on port 49822

2. m.Run() — for each Test* function:
   ├── cleanupDatabase(t)              → TRUNCATE + re-seed in ~10ms
   ├── registerTestUser(t)             → POST /api/v1/auth/register → JWT token
   ├── ... test-specific API calls ... → real HTTP → Echo → service → PostgreSQL
   └── assertions on HTTP responses and/or direct DB queries

3. TestMain cleanup
   ├── httptest.Server.Close()
   ├── app.Close()                     → disconnect Postgres pool
   └── postgres container.Terminate()
```

## Test Utilities

### `testutil/db.go`
- `NewTestDB(ctx)` — starts a `postgres:18-alpine` container via testcontainers, connects, and returns a `TestDB` with DSN
- `RunMigrations()` — drops and recreates the `public` schema, then applies all migrations using `golang-migrate`. Uses `findMigrationsDir()` to locate migrations by walking up to `go.mod`
- `TruncateAllTables()` — truncates all tables (including `stock_movements`, `stocks`) in FK-safe order with `RESTART IDENTITY CASCADE`, then re-seeds system data (roles, permissions, role_permissions)
- `TruncateTables(tables...)` — truncates specific tables
- `BeginTx(ctx)` — starts a transaction for test isolation
- `UniqueCounter()` — exported atomic counter for generating unique test data (e.g. emails, SKUs)

### `testutil/fixtures.go`
- `CreateBaseTestData(ctx)` — creates a complete set: company, branch, role, user, customer, category, product, variant (used for DB-level fixture creation; prefer API-based setup for integration tests)
- `CreateCompany/Branch/Role/User/Customer/ProductCategory/Product/ProductVariant` — individual fixture creators
- `CreateStock(ctx, variantID, branchID, quantity, minQuantity)` — seeds initial stock for a variant at a branch (upsert)

### `setup_test.go` helpers (integration package)
- `registerTestUser(t)` — registers a user via the API, returns `authContext` with `Token`, `RefreshToken`, `UserID`, `CompanyID`, `BranchID`
- `cleanupDatabase(t)` — truncates all tables and re-seeds system data between tests
- `createTestCategory(t, token)` — creates a product category via API, returns its ID
- `createTestProduct(t, token, categoryID, name, sku, price, cost)` — creates a product with one variant via API, returns `(productID, variantID)`
- `testutil.UniqueCounter()` — atomic counter for generating unique test data codes/names

### `testutil/http.go`
- `TestServer` — wraps `*httptest.Server` (real HTTP server with TCP listener) for full-stack HTTP testing
- `Close()` — shuts down the test server
- `GET/POST/PUT/DELETE(path, body/token)` — low-level methods returning `(*Response, error)`; integration tests wrap these via `doPost/doGet/doPut/doDelete` in `helpers_test.go` which assert no transport error and return `*Response` directly
- `Response.ParseResponse()` — parses the standard API envelope into `APIResponse`
- `Response.ParseData(v)` — parses `APIResponse.Data` into a typed struct

## Common Pitfalls

| Pitfall | Solution |
|---------|----------|
| **401 on all protected routes** | Pass `auth.Token` from `registerTestUser(t)` to all API calls |
| **422 "code is required"** | `CreateCustomerRequest` and `CreateProductCategoryRequest` require a `code` field |
| **422 "variants is required"** | `UpdateProductRequest` requires `variants` even for name-only updates |
| **422 "Stok tidak cukup"** | Seed stock in the database before creating orders |
| **"Only confirmed orders can be completed"** | Counter orders auto-complete on full payment; use `"delivery"` fulfillment to test explicit complete |
| **"refresh token expired"** | JWT config must set `AccessTokenExpiry` and `RefreshTokenExpiry` duration fields (not just int fields) |

## Mock Structure

### Repository mocks (`internal/repository/mocks/`)
- Use EXPECT() builder pattern with auto-cleanup
- Created with `NewMockXxxRepository(t)` constructor
