package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/service/mocks"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/irvanmhndra/nexpos-api/pkg/validator"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func createTestContextWithAuth(method, path, body, token string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// =======================
// Constructor Tests
// =======================

func TestNewAuthHandler(t *testing.T) {
	v := validator.New()
	mockSvc := new(mocks.MockAuthService)
	h := NewAuthHandler(mockSvc, v)

	assert.NotNil(t, h)
	assert.NotNil(t, h.validator)
	assert.NotNil(t, h.authSvc)
}

// =======================
// Login Tests
// =======================

func TestAuthHandler_Login_Success(t *testing.T) {
	// Arrange
	mockSvc := new(mocks.MockAuthService)
	v := validator.New()
	h := NewAuthHandler(mockSvc, v)

	loginReq := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedResponse := &dto.LoginResponse{
		User: &dto.AuthUserResponse{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
			Role:  "owner",
			Company: &dto.CompanyInfo{
				ID:   1,
				Code: "TEST-1234",
				Name: "Test Company",
			},
		},
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		ExpiresIn:    3600,
	}

	mockSvc.On("Login", mock.Anything, loginReq, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Return(expectedResponse, nil)

	requestBody := `{"email": "test@example.com", "password": "password123"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", requestBody)

	// Act
	err := h.Login(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "Login successful", response["message"])
	assert.NotNil(t, response["data"])

	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	// Arrange
	mockSvc := new(mocks.MockAuthService)
	v := validator.New()
	h := NewAuthHandler(mockSvc, v)

	loginReq := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockSvc.On("Login", mock.Anything, loginReq, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Return(nil, apperror.InvalidCredentials())

	requestBody := `{"email": "test@example.com", "password": "wrongpassword"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", requestBody)

	// Act
	err := h.Login(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	// Arrange
	v := validator.New()
	mockSvc := new(mocks.MockAuthService)
	h := NewAuthHandler(mockSvc, v)

	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", "{invalid json}")

	// Act
	err := h.Login(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Contains(t, response["message"], "invalid request body")
}

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
			wantStatus:  http.StatusBadRequest,
			wantMessage: "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			v := validator.New()
			mockSvc := new(mocks.MockAuthService)
			h := NewAuthHandler(mockSvc, v)
			c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", tt.requestBody)

			// Act
			err := h.Login(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			var response map[string]interface{}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
			assert.Contains(t, response["message"], tt.wantMessage)
		})
	}
}

// =======================
// Register Tests
// =======================

func TestAuthHandler_Register_Success(t *testing.T) {
	// Arrange
	mockSvc := new(mocks.MockAuthService)
	v := validator.New()
	h := NewAuthHandler(mockSvc, v)

	registerReq := dto.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	expectedResponse := &dto.RegisterResponse{
		User: &dto.AuthUserResponse{
			ID:    1,
			Name:  "John Doe",
			Email: "john@example.com",
			Role:  "owner",
			Company: &dto.CompanyInfo{
				ID:   1,
				Code: "JOHN-1234",
				Name: "John Doe's Business",
			},
		},
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		ExpiresIn:    3600,
	}

	mockSvc.On("Register", mock.Anything, registerReq, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Return(expectedResponse, nil)

	requestBody := `{"name": "John Doe", "email": "john@example.com", "password": "password123"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/register", requestBody)

	// Act
	err := h.Register(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "Registration successful", response["message"])
	assert.NotNil(t, response["data"])

	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Register_EmailAlreadyExists(t *testing.T) {
	// Arrange
	mockSvc := new(mocks.MockAuthService)
	v := validator.New()
	h := NewAuthHandler(mockSvc, v)

	registerReq := dto.RegisterRequest{
		Name:     "John Doe",
		Email:    "existing@example.com",
		Password: "password123",
	}

	mockSvc.On("Register", mock.Anything, registerReq, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Return(nil, apperror.EmailAlreadyExists())

	requestBody := `{"name": "John Doe", "email": "existing@example.com", "password": "password123"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/register", requestBody)

	// Act
	err := h.Register(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Register_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		wantStatus  int
	}{
		{
			name:        "missing name",
			requestBody: `{"name": "", "email": "john@example.com", "password": "pass123"}`,
			wantStatus:  http.StatusUnprocessableEntity,
		},
		{
			name:        "invalid email",
			requestBody: `{"name": "John", "email": "invalid", "password": "pass123"}`,
			wantStatus:  http.StatusUnprocessableEntity,
		},
		{
			name:        "short password",
			requestBody: `{"name": "John", "email": "john@example.com", "password": "123"}`,
			wantStatus:  http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			mockSvc := new(mocks.MockAuthService)
			h := NewAuthHandler(mockSvc, v)
			c, rec := createTestContext(http.MethodPost, "/api/v1/auth/register", tt.requestBody)

			err := h.Register(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

// =======================
// RefreshToken Tests
// =======================

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	// Arrange
	mockSvc := new(mocks.MockAuthService)
	v := validator.New()
	h := NewAuthHandler(mockSvc, v)

	expectedResponse := &dto.RefreshTokenResponse{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		ExpiresIn:    3600,
	}

	mockSvc.On("RefreshToken", mock.Anything, "valid_refresh_token").
		Return(expectedResponse, nil)

	requestBody := `{"refresh_token": "valid_refresh_token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/refresh", requestBody)

	// Act
	err := h.RefreshToken(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "Token refreshed successfully", response["message"])

	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	// Arrange
	mockSvc := new(mocks.MockAuthService)
	v := validator.New()
	h := NewAuthHandler(mockSvc, v)

	mockSvc.On("RefreshToken", mock.Anything, "invalid_token").
		Return(nil, apperror.Unauthorized("invalid refresh token"))

	requestBody := `{"refresh_token": "invalid_token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/refresh", requestBody)

	// Act
	err := h.RefreshToken(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_ExpiredToken(t *testing.T) {
	// Arrange
	mockSvc := new(mocks.MockAuthService)
	v := validator.New()
	h := NewAuthHandler(mockSvc, v)

	mockSvc.On("RefreshToken", mock.Anything, "expired_token").
		Return(nil, apperror.Unauthorized("refresh token expired"))

	requestBody := `{"refresh_token": "expired_token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/refresh", requestBody)

	// Act
	err := h.RefreshToken(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_EmptyToken(t *testing.T) {
	v := validator.New()
	mockSvc := new(mocks.MockAuthService)
	h := NewAuthHandler(mockSvc, v)

	requestBody := `{"refresh_token": ""}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/refresh", requestBody)

	err := h.RefreshToken(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// =======================
// Logout Tests
// =======================

func TestAuthHandler_Logout_WithToken(t *testing.T) {
	// Arrange
	mockSvc := new(mocks.MockAuthService)
	v := validator.New()
	h := NewAuthHandler(mockSvc, v)

	mockSvc.On("Logout", mock.Anything, "valid_access_token").
		Return(nil)

	c, rec := createTestContextWithAuth(http.MethodPost, "/api/v1/auth/logout", "", "valid_access_token")

	// Act
	err := h.Logout(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "Logged out successfully", response["message"])

	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Logout_WithoutToken(t *testing.T) {
	v := validator.New()
	mockSvc := new(mocks.MockAuthService)
	h := NewAuthHandler(mockSvc, v)

	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/logout", "")

	err := h.Logout(c)

	// Should succeed even without token (graceful handling)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "Logged out successfully", response["message"])
}

func TestAuthHandler_Logout_ServiceError(t *testing.T) {
	// Arrange
	mockSvc := new(mocks.MockAuthService)
	v := validator.New()
	h := NewAuthHandler(mockSvc, v)

	// Even if logout fails, we still return success (graceful handling)
	mockSvc.On("Logout", mock.Anything, "some_token").
		Return(apperror.InternalError(nil))

	c, rec := createTestContextWithAuth(http.MethodPost, "/api/v1/auth/logout", "", "some_token")

	// Act
	err := h.Logout(c)

	// Assert - logout always succeeds from handler perspective
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}
