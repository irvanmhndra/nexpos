# Comprehensive Testing Implementation Plan - POS Core API

> **Document Version:** 2.0
> **Date:** 2026-03-28
> **Goal:** Achieve 90%+ service coverage, 80%+ handler coverage, 70%+ repository coverage

---

## Table of Contents

1. [Current State Analysis](#current-state-analysis)
2. [Architecture Overview](#architecture-overview)
3. [Testing Strategy](#testing-strategy)
4. [Implementation Phases](#implementation-phases)
5. [Verification & Coverage Goals](#verification--coverage-goals)
6. [Timeline & Priorities](#timeline--priorities)
7. [Reference Files](#reference-files)

---

## Current State Analysis

### ✅ What We Have (as of 2026-03-28)

- **18 Repository Mocks** — all interfaces mocked in `internal/repository/mocks/`
- **11 Service Unit Test Files — ALL PASSING (198 tests):**
  - `internal/service/auth_test.go`
  - `internal/service/branch_test.go`
  - `internal/service/company_settings_test.go`
  - `internal/service/customer_test.go`
  - `internal/service/inventory_test.go` ✅ (added 2026-03-28)
  - `internal/service/order_test.go` ✅ (added 2026-03-28)
  - `internal/service/product_category_test.go`
  - `internal/service/product_test.go`
  - `internal/service/promotion_test.go`
  - `internal/service/report_test.go`
  - `internal/service/user_test.go`
- **5 Integration Tests** (require live DB via Docker):
  - `tests/integration/auth_test.go`
  - `tests/integration/customer_test.go`
  - `tests/integration/product_test.go`
  - `tests/integration/order_test.go`
  - `tests/integration/health_test.go`
- **Excellent Testing Infrastructure:**
  - testify/mock framework (v1.11.1)
  - Docker-based PostgreSQL for integration tests
  - Test utilities (`testutil/db.go`, `http.go`, `fixtures.go`)
  - Makefile with test targets
  - `TESTING.md` documentation with best practices

### ❌ What's Still Missing

- **Handler Unit Tests** — intentionally skipped (covered by integration tests per team decision)
- **Integration tests for newer features** — suppliers, purchase orders, shifts, expenses not yet covered
- **Service tests for new features** — `purchase_order_test.go`, `shift_test.go`, `expense_test.go` not yet written

---

## Architecture Overview

### Three-Tier Layered Architecture

```
┌─────────────────────────────────────────┐
│     Handler Layer (11 files)            │
│     - Echo v5 REST endpoints             │
│     - Request/Response handling          │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│     Service Layer (11 services)         │
│     - Business logic                     │
│     - Validation & orchestration         │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│   Repository Layer (16 interfaces)      │
│   - Data access via sqlx                │
│   - Database operations                  │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│        PostgreSQL Database              │
└─────────────────────────────────────────┘
```

### 16 Repository Interfaces

1. `UserRepository` - User account management
2. `CompanyRepository` - Company/tenant data
3. `BranchRepository` - Branch/location management
4. `UserSessionRepository` - Authentication sessions
5. `RoleRepository` - User roles
6. `PermissionRepository` - Access permissions
7. `RolePermissionRepository` - Role-permission mapping
8. `UserBranchRepository` - User-branch assignments
9. `CustomerRepository` - Customer data
10. `ProductCategoryRepository` - Product categories
11. `ProductRepository` - Product catalog
12. `ProductVariantRepository` - Product variants/SKUs
13. `OrderRepository` - Order transactions
14. `OrderItemRepository` - Order line items
15. `PaymentRepository` - Payment records
16. `CompanySettingsRepository` - Company configuration
17. `PromotionRepository` - Promotions/discounts
18. `ReportRepository` - Reporting queries

### 11 Services

1. `AuthService` - Login, registration, tokens
2. `UserService` - User CRUD, status management
3. `BranchService` - Branch CRUD operations
4. `CustomerService` - Customer management
5. `ProductService` - Product catalog management
6. `ProductCategoryService` - Category hierarchy
7. `OrderService` - Order processing (most complex)
8. `PromotionService` - Promotion management
9. `ReportService` - Business intelligence reports

---

## Testing Strategy

### Unit Testing Approach

**Service Layer Tests:**
- Mock all repository dependencies using testify/mock
- Test business logic in isolation
- Focus on validation, error handling, and edge cases
- Target: **90%+ coverage**

**Handler Layer Tests:**
- Mock service dependencies
- Test HTTP layer (binding, validation, status codes)
- Do NOT re-test business logic (that's in service tests)
- Target: **80%+ coverage**

### Integration Testing Approach

- Use real PostgreSQL database (Docker)
- Test end-to-end flows through all layers
- Verify actual database operations
- Clean database before each test
- Target: **70%+ repository coverage**

### Test Structure Pattern

Following `TESTING.md` best practices:

```go
package service

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
    repoMocks "github.com/irvanmhndra/pos-core-api/internal/repository/mocks"
)

// =======================
// Test Helpers
// =======================

func setupTestService(t *testing.T) (*Service, *repoMocks.MockRepository) {
    mockRepo := repoMocks.NewMockRepository(t)
    service := NewService(mockRepo)
    return service, mockRepo
}

// =======================
// Constructor Tests
// =======================

func TestNewService(t *testing.T) {
    // Test constructor
}

// =======================
// Method Tests (Arrange-Act-Assert)
// =======================

func TestService_Method_Success(t *testing.T) {
    // Arrange
    service, mockRepo := setupTestService(t)
    mockRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(&model.Entity{}, nil)

    // Act
    result, err := service.Method(context.Background(), 1)

    // Assert
    require.NoError(t, err)
    assert.NotNil(t, result)
    mockRepo.AssertExpectations(t)
}

func TestService_Method_ValidationError(t *testing.T) {
    // Test validation failures
}

func TestService_Method_RepositoryError(t *testing.T) {
    // Test repository error handling
}
```

---

## Implementation Phases

### Phase 1: Mock Generation Foundation ✅ COMPLETE

**Completed:** 2026-03-28
**Result:** 18 repository mocks generated in `internal/repository/mocks/`

#### 1.1 Update .mockery.yaml

**File:** `/Users/irvan/Projects/mybiz/pos-core-api/.mockery.yaml`

Add all 16 repository interfaces:

```yaml
with-expecter: true
packages:
  github.com/irvanmhndra/pos-core-api/internal/repository:
    interfaces:
      UserRepository:
      CompanyRepository:
      BranchRepository:
      UserSessionRepository:
      RoleRepository:
      PermissionRepository:
      RolePermissionRepository:
      UserBranchRepository:
      CustomerRepository:
      ProductCategoryRepository:
      ProductRepository:
      ProductVariantRepository:
      OrderRepository:
      OrderItemRepository:
      PaymentRepository:
      CompanySettingsRepository:
      PromotionRepository:
      ReportRepository:
    config:
      dir: internal/repository/mocks
      outpkg: mocks
```

#### 1.2 Generate Repository Mocks

```bash
cd /Users/irvan/Projects/mybiz/pos-core-api
make mocks
```

**Expected Output:**
- 16 mock files in `internal/repository/mocks/`
- Files: `UserRepository.go`, `CompanyRepository.go`, etc.

#### 1.3 Add Missing Service Interfaces

**File:** `internal/service/interface.go`

Add interfaces for services that don't have them yet:

```go
// BranchServiceInterface defines branch service operations
type BranchServiceInterface interface {
    Create(ctx context.Context, companyID int64, req dto.CreateBranchRequest) (*dto.BranchResponse, error)
    GetByID(ctx context.Context, companyID, id int64) (*dto.BranchResponse, error)
    List(ctx context.Context, companyID int64, req dto.ListBranchRequest) (*dto.BranchListResponse, error)
    Update(ctx context.Context, companyID, id int64, req dto.UpdateBranchRequest) (*dto.BranchResponse, error)
    Delete(ctx context.Context, companyID, id int64) error
}

// PromotionServiceInterface defines promotion service operations
type PromotionServiceInterface interface {
    Create(ctx context.Context, companyID int64, req dto.CreatePromotionRequest) (*dto.PromotionResponse, error)
    GetByID(ctx context.Context, companyID, id int64) (*dto.PromotionResponse, error)
    List(ctx context.Context, companyID int64, req dto.ListPromotionRequest) (*dto.PromotionListResponse, error)
    Update(ctx context.Context, companyID, id int64, req dto.UpdatePromotionRequest) (*dto.PromotionResponse, error)
    Delete(ctx context.Context, companyID, id int64) error
}

// ReportServiceInterface defines reporting service operations
type ReportServiceInterface interface {
    GetSummary(ctx context.Context, companyID int64, req dto.ReportSummaryRequest) (*dto.ReportSummaryResponse, error)
    GetSalesTrend(ctx context.Context, companyID int64, req dto.SalesTrendRequest) ([]*dto.SalesTrendResponse, error)
    GetTopProducts(ctx context.Context, companyID int64, req dto.TopProductsRequest) ([]*dto.TopProductResponse, error)
    GetCategoryRevenue(ctx context.Context, companyID int64, req dto.CategoryRevenueRequest) ([]*dto.CategoryRevenueResponse, error)
    GetPaymentMethods(ctx context.Context, companyID int64, req dto.PaymentMethodsRequest) ([]*dto.PaymentMethodResponse, error)
    GetHourlySales(ctx context.Context, companyID int64, req dto.HourlySalesRequest) ([]*dto.HourlySalesResponse, error)
}

// Compile-time interface assertions
var _ BranchServiceInterface = (*BranchService)(nil)
var _ PromotionServiceInterface = (*PromotionService)(nil)
var _ ReportServiceInterface = (*ReportService)(nil)
```

#### 1.4 Create Service Mocks Manually

**Directory:** `internal/service/mocks/`

Create mock files following the pattern in `auth.go`:

1. `user.go` - UserService mock
2. `branch.go` - BranchService mock
3. `customer.go` - CustomerService mock
4. `product_category.go` - ProductCategoryService mock
5. `product.go` - ProductService mock
6. `order.go` - OrderService mock
7. `promotion.go` - PromotionService mock
8. `report.go` - ReportService mock

**Pattern:**

```go
package mocks

import (
    "context"
    "github.com/stretchr/testify/mock"
    "github.com/irvanmhndra/pos-core-api/internal/dto"
)

type MockServiceName struct {
    mock.Mock
}

func (m *MockServiceName) MethodName(ctx context.Context, params...) (*dto.Response, error) {
    args := m.Called(ctx, params...)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*dto.Response), args.Error(1)
}
```

#### 1.5 Verification

```bash
# Verify mocks are generated
ls -la internal/repository/mocks/

# Verify code compiles
go build ./...

# Run existing tests
go test ./...
```

---

### Phase 2: Service Unit Tests ✅ COMPLETE

**Completed:** 2026-03-28
**Result:** 198 tests across 11 service files — all passing

Create comprehensive unit tests for all 9 services with mocked repositories.

#### Files to Create (9 files, ~2,790 lines total)

| File | Lines | Key Test Cases |
|------|-------|----------------|
| `internal/service/auth_test.go` | ~400 | Login (success, invalid credentials, inactive user, db error), Register (success, email exists, company creation fails), RefreshToken, Logout |
| `internal/service/branch_test.go` | ~270 | Create (success, code exists, repo error), GetByID (success, not found, wrong company), List (with filters, pagination), Update, Delete |
| `internal/service/customer_test.go` | ~250 | Standard CRUD patterns with validation |
| `internal/service/product_category_test.go` | ~280 | HasChildren validation, HasProducts validation, hierarchy validation |
| `internal/service/product_test.go` | ~320 | Variant creation/validation, SKU uniqueness, category association |
| `internal/service/order_test.go` | ~450 | **MOST COMPLEX:** Create with items/payments, ConfirmOrder state transition, AddPayment calculations, CompleteOrder/CancelOrder/VoidOrder validations, RefundPayment logic |
| `internal/service/user_test.go` | ~300 | Create, UpdateStatus, role assignments |
| `internal/service/promotion_test.go` | ~270 | Date range validation, promotion type validation (discount/bundle/conditional) |
| `internal/service/report_test.go` | ~250 | Date range parameter validation, complex query result formatting |

#### Test Coverage Checklist per Service

- [ ] Constructor tests (`TestNewService`)
- [ ] Success cases for all public methods
- [ ] Validation error cases
- [ ] Repository error handling
- [ ] Edge cases (nil values, empty strings, boundary conditions)
- [ ] Business logic validation
- [ ] State transitions (for stateful operations like orders)
- [ ] Table-driven tests for multiple validation scenarios
- [ ] Mock expectations verification

---

### Phase 3: Handler Unit Tests ⏭️ SKIPPED (by design)

**Decision:** Handler layer is covered by integration tests. No handler unit tests beyond `auth_test.go`.

Create unit tests for all handlers with mocked services.

#### Files to Create (9 files, ~2,780 lines total)

| File | Lines | Key Test Cases |
|------|-------|----------------|
| `internal/handler/user_test.go` | ~350 | HTTP binding, validation errors, service mock coordination |
| `internal/handler/branch_test.go` | ~300 | Path/query parameters, list filters |
| `internal/handler/customer_test.go` | ~300 | Standard CRUD HTTP tests |
| `internal/handler/product_category_test.go` | ~320 | Hierarchy endpoint tests |
| `internal/handler/product_test.go` | ~380 | Variant endpoints, SKU validation |
| `internal/handler/order_test.go` | ~500 | **MOST COMPLEX:** Multi-step order flow endpoints |
| `internal/handler/promotion_test.go` | ~300 | Promotion type filters |
| `internal/handler/report_test.go` | ~280 | Query parameter validation for all report types |
| `internal/handler/health_test.go` | ~50 | Simple health check |

#### Handler Test Pattern

Follow existing `internal/handler/auth_test.go` (500+ lines):

```go
package handler

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/labstack/echo/v5"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    serviceMocks "github.com/irvanmhndra/pos-core-api/internal/service/mocks"
)

func setupTestHandler() (*Handler, *serviceMocks.MockService, *echo.Echo) {
    mockService := new(serviceMocks.MockService)
    handler := NewHandler(mockService)
    e := echo.New()
    return handler, mockService, e
}

func TestHandler_Endpoint_Success(t *testing.T) {
    // Arrange
    handler, mockService, e := setupTestHandler()

    requestBody := map[string]interface{}{
        "field": "value",
    }
    body, _ := json.Marshal(requestBody)

    req := httptest.NewRequest(http.MethodPost, "/api/v1/endpoint", bytes.NewBuffer(body))
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    mockService.On("Method", mock.Anything, mock.Anything).Return(&dto.Response{}, nil)

    // Act
    err := handler.Endpoint(c)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, rec.Code)
    mockService.AssertExpectations(t)
}

func TestHandler_Endpoint_ValidationError(t *testing.T) {
    // Test validation failures with table-driven tests
}

func TestHandler_Endpoint_BindingError(t *testing.T) {
    // Test invalid JSON binding
}
```

#### Handler Test Focus

- ✅ Request binding (valid/invalid JSON)
- ✅ Validation errors (use table-driven tests)
- ✅ Path/query parameter parsing
- ✅ HTTP status codes (200, 201, 400, 404, 500)
- ✅ Response structure (success, error format)
- ✅ Service mock expectations
- ❌ NOT testing business logic (that's in service tests)

---

### Phase 4: Integration Tests (MEDIUM Priority)

**Estimated Time:** 1 week
**Complexity:** Medium
**Dependencies:** Phases 1-3 complete

Expand integration tests with real PostgreSQL database.

#### Existing Tests (Keep & Maintain)

- ✅ `tests/integration/auth_test.go`
- ✅ `tests/integration/customer_test.go`
- ✅ `tests/integration/product_test.go`
- ✅ `tests/integration/order_test.go`
- ✅ `tests/integration/health_test.go`

#### Files to Create (5 files, ~1,050 lines total)

| File | Lines | Key Test Cases |
|------|-------|----------------|
| `tests/integration/branch_test.go` | ~200 | CRUD flows, list filters, code uniqueness constraint |
| `tests/integration/user_test.go` | ~250 | CRUD, UpdateStatus, email uniqueness, role assignments |
| `tests/integration/product_category_test.go` | ~220 | CRUD, delete validation (HasChildren/HasProducts), ListAll |
| `tests/integration/promotion_test.go` | ~200 | CRUD, filter by type, active promotions |
| `tests/integration/report_test.go` | ~180 | All report endpoints, date range validation |

#### Integration Test Pattern

```go
func TestIntegration_Feature(t *testing.T) {
    // Clean database before test
    cleanupDatabase(t)
    ctx := context.Background()

    // Create base test data (user, company, branch, etc.)
    testData, err := testFixture.CreateBaseTestData(ctx)
    require.NoError(t, err)

    // Make API call via testServer
    requestBody := map[string]interface{}{
        "field": "value",
    }
    resp, err := testServer.POST("/api/v1/endpoint", requestBody, testData.Token)
    require.NoError(t, err)

    // Assert HTTP response
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    var result dto.Response
    err = resp.ParseResponse(&result)
    require.NoError(t, err)
    assert.True(t, result.Success)

    // Verify database state
    var dbRecord model.Entity
    err = testDB.Get(&dbRecord, "SELECT * FROM table WHERE id = $1", result.Data.ID)
    require.NoError(t, err)
    assert.Equal(t, "expected", dbRecord.Field)
}
```

#### Integration Test Checklist

- [ ] Always clean database before test
- [ ] Use real database transactions
- [ ] Create necessary test data via fixtures
- [ ] Test end-to-end API flows
- [ ] Verify database constraints (unique, foreign keys)
- [ ] Test error responses (404, 400, 409)
- [ ] Clean up resources after test

---

### Phase 5: Test Utilities Enhancement (LOW Priority)

**Estimated Time:** 2-3 days
**Complexity:** Low
**Dependencies:** Phase 4 (to identify common patterns)

#### Extend Test Fixtures

**File:** `tests/integration/testutil/fixtures.go`

Add helper functions for common test data creation:

```go
// CreatePromotion creates a test promotion
func (f *Fixture) CreatePromotion(ctx context.Context, companyID int64, opts ...PromotionOption) (*model.Promotion, error) {
    // Implementation
}

// CreatePayment creates a test payment record
func (f *Fixture) CreatePayment(ctx context.Context, orderID int64, opts ...PaymentOption) (*model.Payment, error) {
    // Implementation
}

// CreateOrderWithItems creates an order with line items
func (f *Fixture) CreateOrderWithItems(ctx context.Context, branchID int64, itemCount int) (*model.Order, []*model.OrderItem, error) {
    // Implementation
}

// CreateRoleWithPermissions creates a role with specific permissions
func (f *Fixture) CreateRoleWithPermissions(ctx context.Context, companyID int64, permissions []string) (*model.Role, error) {
    // Implementation
}
```

---

## Verification & Coverage Goals

### Step-by-Step Verification

#### Step 1: Verify Mock Generation

```bash
cd /Users/irvan/Projects/mybiz/pos-core-api

# Generate mocks
make mocks

# Verify mock files exist
ls -la internal/repository/mocks/
# Expected: 16 *Repository.go files

# Verify compilation
go build ./...
```

#### Step 2: Run Service Tests

```bash
# Run all service tests with verbose output
go test -v ./internal/service/...

# Generate coverage report
go test -coverprofile=service_coverage.out ./internal/service/...

# View coverage summary
go tool cover -func=service_coverage.out | grep total

# Target: (statements) 90.0%+
```

#### Step 3: Run Handler Tests

```bash
# Run all handler tests
go test -v ./internal/handler/...

# Generate coverage report
go test -coverprofile=handler_coverage.out ./internal/handler/...

# View coverage summary
go tool cover -func=handler_coverage.out | grep total

# Target: (statements) 80.0%+
```

#### Step 4: Run Integration Tests

```bash
# Start PostgreSQL container
make test-integration-setup

# Run integration tests
make test-integration

# Stop container
make test-integration-teardown
```

#### Step 5: Full Test Suite

```bash
# Run all tests (unit + integration)
make test

# Generate HTML coverage report
make test-coverage

# Open coverage report in browser
open coverage.html
```

#### Step 6: Coverage by Layer

```bash
# Generate overall coverage
go test -coverprofile=coverage.out ./...

# View coverage by package
go tool cover -func=coverage.out | grep -E "(service|handler|repository)"

# Expected output example:
# internal/service/auth.go         92.5%
# internal/service/branch.go       91.2%
# internal/handler/auth.go         85.3%
# internal/handler/branch.go       82.7%
```

### Coverage Goals Summary

| Layer | Target Coverage | Measurement |
|-------|----------------|-------------|
| **Services** | **90%+** | Line coverage via unit tests with mocked repositories |
| **Handlers** | **80%+** | Line coverage via unit tests with mocked services |
| **Repositories** | **70%+** | Indirectly via integration tests with real database |

---

## Timeline & Priorities

### Week 1: Foundation (CRITICAL)

**Priority:** Must complete before any other phase

- [x] Update `.mockery.yaml` with all 16 repository interfaces
- [ ] Run `make mocks` to generate repository mocks
- [ ] Add missing service interfaces (Branch, Promotion, Report)
- [ ] Create service mocks manually (8 files)
- [ ] Verify compilation: `go build ./...`

**Deliverable:** All mocks generated, code compiles

---

### Week 2-3: Service Tests (HIGH)

**Priority:** Core business logic testing

**Week 2:**
- [ ] `auth_test.go` - Authentication flows (~400 lines)
- [ ] `branch_test.go` - Branch CRUD (~270 lines)
- [ ] `customer_test.go` - Customer management (~250 lines)
- [ ] Verify: `go test ./internal/service/auth*` passing

**Week 3:**
- [ ] `product_category_test.go` - Category hierarchy (~280 lines)
- [ ] `product_test.go` - Product variants (~320 lines)
- [ ] `order_test.go` - Order processing **MOST COMPLEX** (~450 lines)
- [ ] `user_test.go` - User management (~300 lines)
- [ ] Verify: Service coverage > 90%

**Week 3 (continued):**
- [ ] `promotion_test.go` - Promotions (~270 lines)
- [ ] `report_test.go` - Reports (~250 lines)
- [ ] Verify: `go test ./internal/service/...` all passing

**Deliverable:** 9 service test files, 90%+ coverage

---

### Week 4: Handler Tests (MEDIUM)

**Priority:** HTTP layer validation

- [ ] `user_test.go` (~350 lines)
- [ ] `branch_test.go` (~300 lines)
- [ ] `customer_test.go` (~300 lines)
- [ ] `product_category_test.go` (~320 lines)
- [ ] `product_test.go` (~380 lines)
- [ ] `order_test.go` **MOST COMPLEX** (~500 lines)
- [ ] `promotion_test.go` (~300 lines)
- [ ] `report_test.go` (~280 lines)
- [ ] `health_test.go` (~50 lines)

**Pattern:** Follow `internal/handler/auth_test.go` (existing 500+ lines)

**Deliverable:** 9 handler test files, 80%+ coverage

---

### Week 5: Integration Tests (MEDIUM)

**Priority:** End-to-end validation

- [ ] `branch_test.go` (~200 lines)
- [ ] `user_test.go` (~250 lines)
- [ ] `product_category_test.go` (~220 lines)
- [ ] `promotion_test.go` (~200 lines)
- [ ] `report_test.go` (~180 lines)

**Deliverable:** 5 integration test files, 70%+ repository coverage

---

### Week 6: Polish & Documentation

**Priority:** Final cleanup

- [ ] Review coverage gaps
- [ ] Add missing edge case tests
- [ ] Update `TESTING.md` with new patterns
- [ ] Document complex test scenarios
- [ ] Create coverage baseline for CI/CD

**Deliverable:** Complete test suite, documentation updated

---

## Reference Files

### Critical Configuration Files

| File | Purpose |
|------|---------|
| `.mockery.yaml` | Mock generation configuration |
| `Makefile` | Test commands (`test`, `test-integration`, `mocks`) |
| `TESTING.md` | Testing best practices and patterns |

### Service Layer

| File | Purpose |
|------|---------|
| `internal/service/interface.go` | Service interface definitions |
| `internal/service/*_test.go` | Service unit tests (9 files to create) |
| `internal/service/mocks/*.go` | Service mocks (8 files to create) |

### Handler Layer

| File | Purpose |
|------|---------|
| `internal/handler/auth_test.go` | ✅ Reference pattern (existing 500+ lines) |
| `internal/handler/*_test.go` | Handler unit tests (9 files to create) |

### Integration Layer

| File | Purpose |
|------|---------|
| `tests/integration/*_test.go` | Integration tests (5 existing + 5 to create) |
| `tests/integration/testutil/fixtures.go` | Test data factories |
| `tests/integration/testutil/db.go` | Database utilities |
| `tests/integration/testutil/http.go` | HTTP test helpers |

### Repository Layer

| File | Purpose |
|------|---------|
| `internal/repository/mocks/*Repository.go` | Generated mocks (16 files) |

---

## Best Practices Summary

### ✅ DO

1. **Use Table-Driven Tests** for validation scenarios
2. **Always call `mock.AssertExpectations(t)`** to verify mock calls
3. **Follow Arrange-Act-Assert** pattern in all tests
4. **Clean database** before each integration test
5. **Test error cases** as thoroughly as success cases
6. **Use existing patterns** from `auth_test.go` and `TESTING.md`
7. **Keep tests focused** - one concept per test function
8. **Use meaningful test names** - `TestService_Method_Scenario`
9. **Verify coverage** after each phase

### ❌ DON'T

1. **Don't test implementation details** - test public interfaces only
2. **Don't share state** between test functions
3. **Don't skip error assertions** - always check errors
4. **Don't copy-paste** without understanding - reuse patterns thoughtfully
5. **Don't test business logic in handlers** - that belongs in service tests
6. **Don't skip `t.Helper()`** in test helper functions
7. **Don't forget cleanup** in integration tests
8. **Don't ignore flaky tests** - fix or remove them

---

## Summary Statistics

| Metric | Value |
|--------|-------|
| **New Test Files** | 27 files |
| **Total Lines of Code** | ~7,220 lines |
| **Test Functions** | ~260 functions |
| **Service Coverage Goal** | 90%+ |
| **Handler Coverage Goal** | 80%+ |
| **Repository Coverage Goal** | 70%+ |
| **Estimated Duration** | 6 weeks |

---

## Next Steps

1. **Review this plan** with the team
2. **Start with Phase 1** (mock generation)
3. **Implement one example** (e.g., `branch_test.go`) for review
4. **Get feedback** on test patterns and style
5. **Proceed systematically** through each phase
6. **Measure progress** with coverage reports
7. **Iterate and improve** based on learnings

---

**Document Owner:** Testing Implementation Team
**Last Updated:** 2026-03-28
**Status:** Phases 1 & 2 complete. Phase 3 skipped by design. Phases 4 & 5 pending.
