package integration_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthFlow_RegisterAndLogin(t *testing.T) {
	cleanupDatabase(t)

	// Step 1: Register a new user
	resp := doPost(t, "/api/v1/auth/register", map[string]interface{}{
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "password123",
	}, "")
	require.Equal(t, http.StatusCreated, resp.StatusCode, "register failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	require.True(t, r.Success)
	assert.Equal(t, "Registration successful", r.Message)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	assert.NotEmpty(t, data["access_token"])
	assert.NotEmpty(t, data["refresh_token"])
	assert.NotNil(t, data["user"])

	user := data["user"].(map[string]interface{})
	assert.Equal(t, "John Doe", user["name"])
	assert.Equal(t, "john@example.com", user["email"])

	// Step 2: Login with registered credentials
	resp = doPost(t, "/api/v1/auth/login", map[string]interface{}{
		"email":    "john@example.com",
		"password": "password123",
	}, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r = decodeResponse(t, resp)
	assert.True(t, r.Success)
	assert.Equal(t, "Login successful", r.Message)

	var loginData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &loginData))
	assert.NotEmpty(t, loginData["access_token"])
	assert.NotEmpty(t, loginData["refresh_token"])
}

func TestAuthFlow_RefreshToken(t *testing.T) {
	cleanupDatabase(t)

	// Register a user first
	resp := doPost(t, "/api/v1/auth/register", map[string]interface{}{
		"name":     "Jane Doe",
		"email":    "jane@example.com",
		"password": "password123",
	}, "")
	require.Equal(t, http.StatusCreated, resp.StatusCode, "register failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	refreshToken := data["refresh_token"].(string)

	// Refresh the token
	resp = doPost(t, "/api/v1/auth/refresh", map[string]interface{}{
		"refresh_token": refreshToken,
	}, "")
	require.Equal(t, http.StatusOK, resp.StatusCode, "refresh failed: %s", string(resp.Body))

	r = decodeResponse(t, resp)
	require.True(t, r.Success)
	assert.Equal(t, "Token refreshed successfully", r.Message)

	var refreshData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &refreshData))
	assert.NotEmpty(t, refreshData["access_token"])
	assert.NotEmpty(t, refreshData["refresh_token"])
}

func TestAuthFlow_Logout(t *testing.T) {
	cleanupDatabase(t)

	// Register a user first
	resp := doPost(t, "/api/v1/auth/register", map[string]interface{}{
		"name":     "Bob Smith",
		"email":    "bob@example.com",
		"password": "password123",
	}, "")
	require.Equal(t, http.StatusCreated, resp.StatusCode, "register failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	accessToken := data["access_token"].(string)

	// Logout
	resp = doPost(t, "/api/v1/auth/logout", nil, accessToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r = decodeResponse(t, resp)
	assert.True(t, r.Success)
	assert.Equal(t, "Logged out successfully", r.Message)
}

func TestAuth_LoginWithInvalidCredentials(t *testing.T) {
	cleanupDatabase(t)

	// Register a user first
	resp := doPost(t, "/api/v1/auth/register", map[string]interface{}{
		"name":     "Test User",
		"email":    "test@example.com",
		"password": "password123",
	}, "")
	require.Equal(t, http.StatusCreated, resp.StatusCode, "register failed: %s", string(resp.Body))

	// Try to login with wrong password
	resp = doPost(t, "/api/v1/auth/login", map[string]interface{}{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	r := decodeResponse(t, resp)
	assert.False(t, r.Success)
}

func TestAuth_RegisterDuplicateEmail(t *testing.T) {
	cleanupDatabase(t)

	// Register first user
	resp := doPost(t, "/api/v1/auth/register", map[string]interface{}{
		"name":     "First User",
		"email":    "duplicate@example.com",
		"password": "password123",
	}, "")
	require.Equal(t, http.StatusCreated, resp.StatusCode, "register failed: %s", string(resp.Body))

	// Try to register with same email
	resp = doPost(t, "/api/v1/auth/register", map[string]interface{}{
		"name":     "Second User",
		"email":    "duplicate@example.com",
		"password": "password456",
	}, "")
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	r := decodeResponse(t, resp)
	assert.False(t, r.Success)
}

func TestAuth_RegisterValidation(t *testing.T) {
	cleanupDatabase(t)

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
	}{
		{
			name: "empty name",
			body: map[string]interface{}{
				"name":     "",
				"email":    "test@example.com",
				"password": "password123",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid email",
			body: map[string]interface{}{
				"name":     "Test User",
				"email":    "invalid-email",
				"password": "password123",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "short password",
			body: map[string]interface{}{
				"name":     "Test User",
				"email":    "test@example.com",
				"password": "short",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "missing email",
			body: map[string]interface{}{
				"name":     "Test User",
				"password": "password123",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := doPost(t, "/api/v1/auth/register", tt.body, "")
			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			r := decodeResponse(t, resp)
			assert.False(t, r.Success)
		})
	}
}

func TestAuth_LoginValidation(t *testing.T) {
	cleanupDatabase(t)

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
	}{
		{
			name: "empty email",
			body: map[string]interface{}{
				"email":    "",
				"password": "password123",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "empty password",
			body: map[string]interface{}{
				"email":    "test@example.com",
				"password": "",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid email format",
			body: map[string]interface{}{
				"email":    "not-an-email",
				"password": "password123",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := doPost(t, "/api/v1/auth/login", tt.body, "")
			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			r := decodeResponse(t, resp)
			assert.False(t, r.Success)
		})
	}
}

func TestAuth_RefreshTokenInvalid(t *testing.T) {
	cleanupDatabase(t)

	// Try to refresh with invalid token
	resp := doPost(t, "/api/v1/auth/refresh", map[string]interface{}{
		"refresh_token": "invalid-refresh-token",
	}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	r := decodeResponse(t, resp)
	assert.False(t, r.Success)
}
