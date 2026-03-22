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

Integration tests use the real HTTP stack and database:

```go
package integration

func TestCustomer_CreateAndGet(t *testing.T) {
    cleanupDatabase(t)
    ctx := context.Background()

    testData, err := testFixture.CreateBaseTestData(ctx)
    require.NoError(t, err)

    // Create
    body := map[string]interface{}{"name": "John", "phone": "+123"}
    resp, err := testServer.POST("/api/v1/customers", body, "")
    require.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    // Get
    resp, err = testServer.GET("/api/v1/customers/1", "")
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

### Integration test with DB assertions

Use direct database queries to verify side effects not visible in API responses:

```go
func TestOrderFlow_Complete_StockDeduction(t *testing.T) {
    cleanupDatabase(t)
    ctx := context.Background()

    testData, err := testFixture.CreateBaseTestData(ctx)
    require.NoError(t, err)

    // Seed initial stock
    err = testFixture.CreateStock(ctx, testData.ProductVariant.ID, testData.Branch.ID, 10, 0)
    require.NoError(t, err)

    // ... create, confirm, pay, complete order ...

    // DB assertion: verify stock was deducted
    var stockQty int
    err = testDB.DB.QueryRowContext(ctx,
        "SELECT quantity FROM stocks WHERE product_variant_id = $1 AND branch_id = $2",
        testData.ProductVariant.ID, testData.Branch.ID,
    ).Scan(&stockQty)
    require.NoError(t, err)
    assert.Equal(t, 9, stockQty) // started at 10, ordered 1

    // DB assertion: verify stock movement record
    var movementType string
    var movementQty int
    err = testDB.DB.QueryRowContext(ctx,
        `SELECT type, quantity FROM stock_movements
         WHERE reference_id = $1 AND reference_type = 'order'`,
        orderID,
    ).Scan(&movementType, &movementQty)
    require.NoError(t, err)
    assert.Equal(t, "OUT", movementType)
    assert.Equal(t, 1, movementQty)
}
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
- `RunMigrations()` — applies all migration files
- `TruncateAllTables()` — truncates all tables (including `stock_movements`, `stocks`) in FK-safe order
- `TruncateTables(tables...)` — truncates specific tables
- `BeginTx(ctx)` — starts a transaction for test isolation

### `testutil/fixtures.go`
- `CreateBaseTestData(ctx)` — creates a complete set: company, branch, role, user, customer, category, product, variant
- `CreateCompany/Branch/Role/User/Customer/ProductCategory/Product/ProductVariant` — individual fixture creators
- `CreateStock(ctx, variantID, branchID, quantity, minQuantity)` — seeds initial stock for a variant at a branch (upsert)

### `testutil/http.go`
- `TestServer` — wraps Echo for HTTP testing via `ServeHTTP`
- `GET/POST/PUT/DELETE(path, body/token)` — convenience methods
- `Response.ParseResponse()` — parses standard API response
- `Response.ParseData(v)` — parses response data into a struct

## Mock Structure

### Repository mocks (`internal/repository/mocks/`)
- Use EXPECT() builder pattern with auto-cleanup
- Created with `NewMockXxxRepository(t)` constructor

### Service mocks (`internal/service/mocks/`)
- Simple `testify/mock` with `.On().Return()` pattern
- Created with `new(mocks.MockXxxService)`
