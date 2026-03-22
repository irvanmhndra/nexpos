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
│  Handler Unit Tests (internal/handler/*_test.go) │
│  Mock services, test HTTP binding & validation   │
│  Tests: all handlers                             │
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
| **Handler unit tests** | HTTP binding, validation, status codes, response format | Service interfaces | **Medium** — handlers are thin, but catches HTTP-level bugs without a database |
| **Integration tests** | Full request → response → database flow, including side effects | Nothing (real stack) | **High** — verifies everything works together |

### Key design decisions

1. **Handlers use service interfaces** (`service.XxxServiceInterface`) — enables mocking for handler unit tests
2. **Services use repository interfaces** — enables mocking for service unit tests
3. **Integration tests live in `tests/integration/`** — a separate package is required because integration tests need to import `internal/app` to boot the real application, but `internal/app` already imports `internal/handler`. Placing integration tests inside `internal/handler/` would create a circular import (`handler_test → app → handler`). The separate package can import anything without this issue, giving it access to the real app, database, middleware, and route registration
4. **Integration tests share setup via `TestMain`** in `setup_test.go` — database, app, and test server are initialized once
5. **Integration tests use DB assertions for side effects** — when an operation produces database changes not visible in the API response (e.g. stock movements, payment status transitions), we query the database directly to verify

## Project Test Structure

```
internal/
  handler/
    auth.go
    auth_test.go          ← handler unit test (mocks service)
    customer.go
    customer_test.go
    ...
  service/
    branch.go
    branch_test.go        ← service unit test (mocks repository)
    customer.go
    customer_test.go
    ...
    mocks/                ← service mocks (for handler tests)
      auth.go
      customer.go
      ...
  repository/
    mocks/                ← repository mocks (for service tests)
      branch.go
      customer.go
      ...
tests/
  integration/
    setup_test.go         ← TestMain: DB + app setup/teardown
    auth_test.go          ← full-stack integration tests
    customer_test.go
    order_test.go         ← includes DB assertions for stock/payment
    ...
    testutil/
      db.go               ← test database utilities (truncate, tx)
      http.go             ← test HTTP client (GET, POST, PUT, DELETE)
      fixtures.go         ← test data factories (company, branch, user, product, stock, etc.)
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

### Integration tests (requires test database)
```bash
# Start test database
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test -v ./tests/integration/...

# Stop test database
docker-compose -f docker-compose.test.yml down
```

### All tests
```bash
go test ./...
```

### Specific package
```bash
go test -v ./internal/service/...
go test -v ./internal/handler/...
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

### Handler unit test pattern

Handler tests mock service dependencies:

```go
package handler

func setupCustomerHandler(t *testing.T) (*CustomerHandler, *mocks.MockCustomerService) {
    t.Helper()
    v := validator.New()
    mockSvc := new(mocks.MockCustomerService)
    h := NewCustomerHandler(mockSvc, v)
    return h, mockSvc
}

func createCustomerContext(method, path, body string) (*echo.Context, *httptest.ResponseRecorder) {
    e := echo.New()
    req := httptest.NewRequest(method, path, strings.NewReader(body))
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    c.Set("company_id", int64(1))
    return c, rec
}

func TestCustomerHandler_Create_Success(t *testing.T) {
    h, mockSvc := setupCustomerHandler(t)

    mockSvc.On("Create", mock.Anything, int64(1), mock.AnythingOfType("dto.CreateCustomerRequest")).
        Return(&dto.CustomerResponse{ID: 1, Name: "Test"}, nil)

    body := `{"code": "CUST-001", "name": "Test"}`
    c, rec := createCustomerContext(http.MethodPost, "/api/v1/customers", body)

    err := h.Create(c)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, rec.Code)
    mockSvc.AssertExpectations(t)
}
```

### Echo v5 path parameters

Use `SetPathValues` (not `SetPathParams`) for Echo v5:

```go
c, rec := createContext(http.MethodGet, "/api/v1/items/1", "")
c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})
```

### Integration test pattern

Integration tests use the real HTTP stack and database. All protected routes require authentication via `registerTestUser()`, which registers a user via the API and returns an `authContext` with tokens and IDs:

```go
package integration

func TestCustomer_CreateAndGet(t *testing.T) {
    cleanupDatabase(t)
    auth := registerTestUser(t)

    // Create (all API calls pass auth.Token)
    body := map[string]interface{}{
        "code": "CUST-001",
        "name": "John Customer",
        "phone": "+1234567890",
    }
    resp, err := testServer.POST("/api/v1/customers", body, auth.Token)
    require.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    // Parse ID from response
    var createResp map[string]interface{}
    json.Unmarshal(resp.Body, &createResp)
    data := createResp["data"].(map[string]interface{})
    customerID := int64(data["id"].(float64))

    // Get
    resp, err = testServer.GET(fmt.Sprintf("/api/v1/customers/%d", customerID), auth.Token)
    require.NoError(t, err)
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
    customerID := createTestCustomer(t, auth.Token)
    categoryID := createTestCategory(t, auth.Token)
    productID, variantID := createTestProduct(t, auth.Token, categoryID, "Test Product", "SKU-001", 100.00, 50.00)

    // Seed stock (required — order creation validates stock availability)
    _, err := testDB.DB.Exec(
        `INSERT INTO stocks (product_variant_id, branch_id, quantity, min_quantity)
         VALUES ($1, $2, 100, 0)
         ON CONFLICT (product_variant_id, branch_id) DO UPDATE SET quantity = 100`,
        variantID, auth.BranchID,
    )
    require.NoError(t, err)

    return &orderTestData{auth: auth, customerID: customerID, ...}
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
err = testDB.DB.QueryRowContext(ctx,
    "SELECT quantity FROM stocks WHERE product_variant_id = $1 AND branch_id = $2",
    variantID, auth.BranchID,
).Scan(&stockQty)
require.NoError(t, err)
assert.Equal(t, 9, stockQty) // started at 10, ordered 1
```
```

### When to use DB assertions

Most integration tests only need to verify via HTTP responses (POST then GET back). Use direct DB assertions when:

| Scenario | Why DB assertion is needed |
|----------|--------------------------|
| **Stock deduction on order complete** | Stock changes are best-effort side effects, not returned in the order response |
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

// Handler tests
func TestCustomerHandler_Create_Success(t *testing.T)
func TestCustomerHandler_Create_InvalidJSON(t *testing.T)
func TestCustomerHandler_Create_ValidationError(t *testing.T)

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

### Handler tests — focus on:
- JSON binding (valid/invalid)
- Validation errors (required fields, format constraints)
- Path/query parameter parsing
- HTTP status codes
- Response message format

### Integration tests — focus on:
- End-to-end flows (create → get → update → delete)
- Database constraints (unique, foreign keys)
- Pagination and search
- Authentication flows
- Side effects via DB assertions (stock movements, payment status)

## Test Utilities

### `testutil/db.go`
- `NewTestDB()` — creates a database connection to the test PostgreSQL instance
- `RunMigrations()` — drops and recreates the `public` schema, then applies all `.up.sql` migration files in order. The schema reset ensures migrations can be re-applied cleanly across test runs
- `TruncateAllTables()` — truncates all tables (including `stock_movements`, `stocks`) in FK-safe order, then re-seeds system data (roles, permissions, role_permissions)
- `TruncateTables(tables...)` — truncates specific tables
- `BeginTx(ctx)` — starts a transaction for test isolation

### `testutil/fixtures.go`
- `CreateBaseTestData(ctx)` — creates a complete set: company, branch, role, user, customer, category, product, variant (used for DB-level fixture creation; prefer API-based setup for integration tests)
- `CreateCompany/Branch/Role/User/Customer/ProductCategory/Product/ProductVariant` — individual fixture creators
- `CreateStock(ctx, variantID, branchID, quantity, minQuantity)` — seeds initial stock for a variant at a branch (upsert)

### `setup_test.go` helpers (integration package)
- `registerTestUser(t)` — registers a user via the API, returns `authContext` with `Token`, `RefreshToken`, `UserID`, `CompanyID`, `BranchID`
- `cleanupDatabase(t)` — truncates all tables and re-seeds system data between tests
- `createTestCategory(t, token)` — creates a product category via API, returns its ID
- `createTestProduct(t, token, categoryID, name, sku, price, cost)` — creates a product with one variant via API, returns `(productID, variantID)`
- `uniqueCounter()` — atomic counter for generating unique test data codes/names

### `testutil/http.go`
- `TestServer` — wraps Echo for HTTP testing via `ServeHTTP`
- `GET/POST/PUT/DELETE(path, body/token)` — convenience methods
- `Response.ParseResponse()` — parses standard API response
- `Response.ParseData(v)` — parses response data into a struct

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

### Service mocks (`internal/service/mocks/`)
- Simple `testify/mock` with `.On().Return()` pattern
- Created with `new(mocks.MockXxxService)`
