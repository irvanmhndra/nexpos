package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/irvanmhndra/pos-core-api/pkg/validator"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

// =======================
// Test Helpers
// =======================

func createTestContext(method, path, body string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// =======================
// Constructor Tests
// =======================

func TestNewAuthHandler(t *testing.T) {
	v := validator.New()
	h := NewAuthHandler(nil, v)

	assert.NotNil(t, h)
	assert.NotNil(t, h.validator)
}

// =======================
// Login Validation Tests
// =======================

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	// Arrange
	v := validator.New()
	h := NewAuthHandler(nil, v)

	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", "{invalid json}")

	// Act
	err := h.Login(c)

	// Assert - Echo handlers return nil even on errors (handled internally)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Contains(t, response["message"], "invalid request body")
}

func TestAuthHandler_Login_MissingEmail(t *testing.T) {
	// Arrange
	v := validator.New()
	h := NewAuthHandler(nil, v)

	requestBody := `{
		"email": "",
		"password": "password123"
	}`

	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", requestBody)

	// Act
	err := h.Login(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Validation failed", response["message"])
	assert.NotNil(t, response["errors"])
}

func TestAuthHandler_Login_InvalidEmailFormat(t *testing.T) {
	// Arrange
	v := validator.New()
	h := NewAuthHandler(nil, v)

	requestBody := `{
		"email": "not-an-email",
		"password": "password123"
	}`

	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", requestBody)

	// Act
	err := h.Login(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Validation failed", response["message"])
}

// =======================
// Table-Driven Validation Tests
// =======================

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
		{
			name:        "invalid email format",
			requestBody: `{"email": "invalid", "password": "pass123"}`,
			wantStatus:  http.StatusUnprocessableEntity,
			wantMessage: "Validation failed",
		},
		{
			name:        "both fields empty",
			requestBody: `{"email": "", "password": ""}`,
			wantStatus:  http.StatusUnprocessableEntity,
			wantMessage: "Validation failed",
		},
		{
			name:        "malformed json",
			requestBody: `{email: test}`,
			wantStatus:  http.StatusBadRequest, // JSON parsing error returns 400, not 422
			wantMessage: "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			v := validator.New()
			h := NewAuthHandler(nil, v)
			c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", tt.requestBody)

			// Act
			err := h.Login(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)
			assert.Contains(t, response["message"], tt.wantMessage)
		})
	}
}

// =======================
// Register Validation Tests
// =======================

func TestAuthHandler_Register_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		wantStatus  int
	}{
		{
			name:        "missing company name",
			requestBody: `{"company_name": "", "owner_name": "John", "email": "john@example.com", "password": "pass123"}`,
			wantStatus:  http.StatusUnprocessableEntity,
		},
		{
			name:        "missing owner name",
			requestBody: `{"company_name": "ACME", "owner_name": "", "email": "john@example.com", "password": "pass123"}`,
			wantStatus:  http.StatusUnprocessableEntity,
		},
		{
			name:        "invalid email",
			requestBody: `{"company_name": "ACME", "owner_name": "John", "email": "invalid", "password": "pass123"}`,
			wantStatus:  http.StatusUnprocessableEntity,
		},
		{
			name:        "short password",
			requestBody: `{"company_name": "ACME", "owner_name": "John", "email": "john@example.com", "password": "123"}`,
			wantStatus:  http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			h := NewAuthHandler(nil, v)
			c, rec := createTestContext(http.MethodPost, "/api/v1/auth/register", tt.requestBody)

			err := h.Register(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

// =======================
// RefreshToken Validation Tests
// =======================

func TestAuthHandler_RefreshToken_EmptyToken(t *testing.T) {
	v := validator.New()
	h := NewAuthHandler(nil, v)

	requestBody := `{"refresh_token": ""}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/refresh", requestBody)

	err := h.RefreshToken(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// =======================
// Logout Tests
// =======================

func TestAuthHandler_Logout_WithoutToken(t *testing.T) {
	v := validator.New()
	h := NewAuthHandler(nil, v)

	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/logout", "")

	err := h.Logout(c)

	// Should succeed even without token (graceful handling)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Logged out successfully", response["message"])
}

// TestAuthHandler_Logout_WithToken is commented out because it requires a mock service
// In a full implementation with mocks/interfaces, this test would verify that
// the service.Logout method is called with the correct token
//
// func TestAuthHandler_Logout_WithToken(t *testing.T) {
// 	mockSvc := new(MockAuthService)
// 	v := validator.New()
// 	h := NewAuthHandler(mockSvc, v)
//
// 	mockSvc.On("Logout", mock.Anything, "some_token").Return(nil)
//
// 	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/logout", "")
// 	(*c).Request().Header.Set("Authorization", "Bearer some_token")
//
// 	err := h.Logout(c)
//
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusOK, rec.Code)
// 	mockSvc.AssertExpectations(t)
// }
