# Testing Guide

## File Naming Convention

Kami menggunakan **standard Go convention**:

```
handler/
  ├── auth.go              # Implementation
  ├── auth_test.go         # Tests
  ├── customer.go
  ├── customer_test.go
  ├── product.go
  ├── product_test.go
```

### ❌ JANGAN gunakan:
```
handler/
  ├── auth_handler.go
  ├── auth_handler_test.go   # Terlalu verbose
```

## Test Package Strategy

### Option 1: Same Package (White Box Testing)
```go
package handler

func TestAuthHandler_Login(t *testing.T) {
    // Dapat access private methods/fields
    // Gunakan untuk test internal logic
}
```

**Pros:**
- Access ke private methods
- Detail testing

**Cons:**
- Tight coupling

### Option 2: Separate Package (Black Box Testing) ⭐ **RECOMMENDED**
```go
package handler_test

import "github.com/irvanmhndra/pos-core-api/internal/handler"

func TestAuthHandler_Login(t *testing.T) {
    // Test hanya public API
}
```

**Pros:**
- Test public interface only
- Better design
- Easier refactoring
- Mencegah test implementation details

## Test Structure

### 1. Test Organization
```go
// =======================
// Test Helpers
// =======================

func createTestContext(method, path, body string) (*echo.Context, *httptest.ResponseRecorder) {
    // ... helper code
}

// =======================
// Constructor Tests
// =======================

func TestNewAuthHandler(t *testing.T) {
    // Test constructor
}

// =======================
// Login Tests
// =======================

func TestAuthHandler_Login_Success(t *testing.T) {
    // ... test code
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
    // ... test code
}
```

### 2. Test Naming Convention
Format: `Test<Handler>_<Method>_<Scenario>`

**Examples:**
```go
func TestAuthHandler_Login_Success(t *testing.T)
func TestAuthHandler_Login_InvalidEmail(t *testing.T)
func TestAuthHandler_Login_EmptyPassword(t *testing.T)
func TestProductHandler_Create_ValidationError(t *testing.T)
```

### 3. Arrange-Act-Assert Pattern
```go
func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
    // Arrange - Setup test data & dependencies
    v := validator.New()
    h := NewAuthHandler(nil, v)
    c, rec := createTestContext(http.MethodPost, "/login", "{invalid json}")

    // Act - Execute the function being tested
    err := h.Login(c)

    // Assert - Verify the results
    assert.NoError(t, err)
    assert.Equal(t, http.StatusBadRequest, rec.Code)
}
```

## Testing Patterns

### 1. Table-Driven Tests ⭐ **RECOMMENDED for multiple scenarios**

```go
func TestAuthHandler_Login_ValidationErrors(t *testing.T) {
    tests := []struct {
        name        string
        requestBody string
        wantStatus  int
        wantMessage string
    }{
        {
            name:        "empty email",
            requestBody: `{"email": "", "password": "pass123"}`,
            wantStatus:  http.StatusUnprocessableEntity,
            wantMessage: "Validation failed",
        },
        {
            name:        "empty password",
            requestBody: `{"email": "test@example.com", "password": ""}`,
            wantStatus:  http.StatusUnprocessableEntity,
            wantMessage: "Validation failed",
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            v := validator.New()
            h := NewAuthHandler(nil, v)
            c, rec := createTestContext(http.MethodPost, "/login", tt.requestBody)

            // Act
            err := h.Login(c)

            // Assert
            assert.NoError(t, err)
            assert.Equal(t, tt.wantStatus, rec.Code)
        })
    }
}
```

**Benefits:**
- DRY (Don't Repeat Yourself)
- Easy to add new test cases
- Clear test documentation
- Parallel test execution support

### 2. Individual Tests (for complex scenarios)

```go
func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
    v := validator.New()
    h := NewAuthHandler(nil, v)
    c, rec := createTestContext(http.MethodPost, "/login", "{invalid json}")

    err := h.Login(c)

    assert.NoError(t, err)
    assert.Equal(t, http.StatusBadRequest, rec.Code)

    var response map[string]interface{}
    json.Unmarshal(rec.Body.Bytes(), &response)
    assert.Contains(t, response["message"], "invalid request body")
}
```

## What to Test in Handlers

### ✅ DO Test:
1. **Input Validation**
   - Missing required fields
   - Invalid field formats
   - Malformed JSON
   - Field length constraints

2. **HTTP Response**
   - Status codes
   - Response structure
   - Error messages

3. **Edge Cases**
   - Empty strings
   - Null values
   - Boundary values

### ❌ DON'T Test (without mocks):
1. **Business Logic** - Test di service layer
2. **Database Operations** - Test di repository layer
3. **External API Calls** - Requires mocks

## Running Tests

### Run all tests
```bash
go test ./...
```

### Run specific package
```bash
go test ./internal/handler
```

### Run with verbose output
```bash
go test ./internal/handler -v
```

### Run specific test
```bash
go test ./internal/handler -run TestAuthHandler_Login
```

### Run with coverage
```bash
go test ./internal/handler -cover
```

### Generate coverage report
```bash
go test ./internal/handler -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Run tests in parallel
```bash
go test ./... -parallel=4
```

## Test Coverage Goals

- **Handlers:** 80%+ (focus on validation, input parsing)
- **Services:** 90%+ (business logic critical)
- **Repositories:** 70%+ (mostly integration tests)
- **Utils/Helpers:** 100% (small, pure functions)

## Advanced Topics (Future Implementation)

### 1. Mocking with Interfaces

Create service interfaces for better testability:

```go
// internal/service/interface.go
type AuthServiceInterface interface {
    Login(ctx context.Context, req dto.LoginRequest, ip, ua string) (*dto.LoginResponse, error)
    Register(ctx context.Context, req dto.RegisterRequest, ip, ua string) (*dto.LoginResponse, error)
    // ... other methods
}

// Handler depends on interface, not concrete type
type AuthHandler struct {
    authSvc   AuthServiceInterface  // Interface, not *AuthService
    validator *validator.CustomValidator
}
```

Then use testify/mock:

```go
type MockAuthService struct {
    mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, req dto.LoginRequest, ip, ua string) (*dto.LoginResponse, error) {
    args := m.Called(ctx, req, ip, ua)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*dto.LoginResponse), args.Error(1)
}

// In test
mockSvc := new(MockAuthService)
mockSvc.On("Login", mock.Anything, mock.AnythingOfType("dto.LoginRequest"), mock.Anything, mock.Anything).
    Return(&dto.LoginResponse{AccessToken: "token"}, nil)

h := handler.NewAuthHandler(mockSvc, v)
// ... test with mock
mockSvc.AssertExpectations(t)
```

### 2. Integration Tests

```go
// tests/integration/auth_test.go
func TestAuthFlow_EndToEnd(t *testing.T) {
    // Setup test database
    // Create real dependencies
    // Test full flow: Register -> Login -> API Call -> Logout
}
```

### 3. Test Fixtures

```go
// internal/handler/fixtures_test.go
func validLoginRequest() string {
    return `{"email": "test@example.com", "password": "password123"}`
}

func invalidEmailRequest() string {
    return `{"email": "invalid", "password": "password123"}`
}
```

## Best Practices

1. **Test one thing per test**
   - Each test should verify one specific behavior

2. **Use descriptive test names**
   - Name should describe what is being tested and expected outcome

3. **Keep tests independent**
   - Tests should not depend on each other
   - Use setup/teardown if needed

4. **Test behavior, not implementation**
   - Don't test private methods directly
   - Focus on public API

5. **Use table-driven tests for similar scenarios**
   - Reduces code duplication
   - Makes adding test cases easier

6. **Mock external dependencies**
   - Database calls
   - External APIs
   - Time-dependent code

7. **Clean up test resources**
   - Use t.Cleanup() or defer

8. **Run tests in CI/CD**
   - Automated testing on every commit

## Common Pitfalls to Avoid

❌ **Testing implementation details**
```go
// Bad - testing internal field
assert.NotNil(t, handler.authSvc)
```

❌ **Not testing error cases**
```go
// Bad - only testing happy path
func TestCreate_Success(t *testing.T) { ... }
// Missing: TestCreate_ValidationError, TestCreate_DBError, etc.
```

❌ **Hardcoding values without context**
```go
// Bad
assert.Equal(t, 200, rec.Code)

// Good
assert.Equal(t, http.StatusOK, rec.Code)
```

❌ **Not using test helpers**
```go
// Bad - repeating setup code in every test
e := echo.New()
req := httptest.NewRequest(...)
// ...

// Good - use helper function
c, rec := createTestContext(method, path, body)
```

## HTTP Status Codes Reference

- `200 OK` - Success
- `201 Created` - Resource created
- `400 Bad Request` - Malformed request (JSON parsing error)
- `401 Unauthorized` - Authentication required/failed
- `404 Not Found` - Resource not found
- `422 Unprocessable Entity` - Validation error (correct data format, invalid values)
- `500 Internal Server Error` - Server error

## Example: Complete Test File Structure

```go
package handler

import (
    "testing"
    // ... imports
)

// =======================
// Test Helpers
// =======================

func createTestContext(...) { }

// =======================
// Constructor Tests
// =======================

func TestNewHandler(t *testing.T) { }

// =======================
// Create Tests
// =======================

func TestHandler_Create_Success(t *testing.T) { }
func TestHandler_Create_ValidationError(t *testing.T) { }
func TestHandler_Create_Errors(t *testing.T) {
    // Table-driven tests for multiple error cases
}

// =======================
// List Tests
// =======================

func TestHandler_List_Success(t *testing.T) { }
func TestHandler_List_WithFilters(t *testing.T) { }

// =======================
// Update Tests
// =======================

func TestHandler_Update_Success(t *testing.T) { }
func TestHandler_Update_NotFound(t *testing.T) { }

// =======================
// Delete Tests
// =======================

func TestHandler_Delete_Success(t *testing.T) { }
```

## Resources

- [Go Testing](https://golang.org/pkg/testing/)
- [Testify](https://github.com/stretchr/testify)
- [Echo Testing](https://echo.labstack.com/docs/testing)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
