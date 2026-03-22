package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthFlow_RegisterAndLogin(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Step 1: Register a new user
	registerBody := map[string]interface{}{
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "password123",
	}

	resp, err := testServer.POST("/api/v1/auth/register", registerBody, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "register failed: %s", string(resp.Body))

	var registerResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &registerResp)
	require.NoError(t, err)

	require.True(t, registerResp["success"].(bool))
	assert.Equal(t, "Registration successful", registerResp["message"])

	data := registerResp["data"].(map[string]interface{})
	assert.NotEmpty(t, data["access_token"])
	assert.NotEmpty(t, data["refresh_token"])
	assert.NotNil(t, data["user"])

	user := data["user"].(map[string]interface{})
	assert.Equal(t, "John Doe", user["name"])
	assert.Equal(t, "john@example.com", user["email"])

	// Step 2: Login with registered credentials
	loginBody := map[string]interface{}{
		"email":    "john@example.com",
		"password": "password123",
	}

	resp, err = testServer.POST("/api/v1/auth/login", loginBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &loginResp)
	require.NoError(t, err)

	assert.True(t, loginResp["success"].(bool))
	assert.Equal(t, "Login successful", loginResp["message"])

	loginData := loginResp["data"].(map[string]interface{})
	accessToken := loginData["access_token"].(string)
	refreshToken := loginData["refresh_token"].(string)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	_ = ctx // Context available for future use
}

func TestAuthFlow_RefreshToken(t *testing.T) {
	cleanupDatabase(t)

	// Register a user first
	registerBody := map[string]interface{}{
		"name":     "Jane Doe",
		"email":    "jane@example.com",
		"password": "password123",
	}

	resp, err := testServer.POST("/api/v1/auth/register", registerBody, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "register failed: %s", string(resp.Body))

	var registerResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &registerResp)
	require.NoError(t, err)

	data := registerResp["data"].(map[string]interface{})
	refreshToken := data["refresh_token"].(string)

	// Refresh the token
	refreshBody := map[string]interface{}{
		"refresh_token": refreshToken,
	}

	resp, err = testServer.POST("/api/v1/auth/refresh", refreshBody, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "refresh failed: %s", string(resp.Body))

	var refreshResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &refreshResp)
	require.NoError(t, err)

	require.True(t, refreshResp["success"].(bool))
	assert.Equal(t, "Token refreshed successfully", refreshResp["message"])

	refreshData := refreshResp["data"].(map[string]interface{})
	assert.NotEmpty(t, refreshData["access_token"])
	assert.NotEmpty(t, refreshData["refresh_token"])
}

func TestAuthFlow_Logout(t *testing.T) {
	cleanupDatabase(t)

	// Register a user first
	registerBody := map[string]interface{}{
		"name":     "Bob Smith",
		"email":    "bob@example.com",
		"password": "password123",
	}

	resp, err := testServer.POST("/api/v1/auth/register", registerBody, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "register failed: %s", string(resp.Body))

	var registerResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &registerResp)
	require.NoError(t, err)

	data := registerResp["data"].(map[string]interface{})
	accessToken := data["access_token"].(string)

	// Logout
	resp, err = testServer.POST("/api/v1/auth/logout", nil, accessToken)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var logoutResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &logoutResp)
	require.NoError(t, err)

	assert.True(t, logoutResp["success"].(bool))
	assert.Equal(t, "Logged out successfully", logoutResp["message"])
}

func TestAuth_LoginWithInvalidCredentials(t *testing.T) {
	cleanupDatabase(t)

	// Register a user first
	registerBody := map[string]interface{}{
		"name":     "Test User",
		"email":    "test@example.com",
		"password": "password123",
	}

	resp, err := testServer.POST("/api/v1/auth/register", registerBody, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "register failed: %s", string(resp.Body))

	// Try to login with wrong password
	loginBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}

	resp, err = testServer.POST("/api/v1/auth/login", loginBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var loginResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &loginResp)
	require.NoError(t, err)

	assert.False(t, loginResp["success"].(bool))
}

func TestAuth_RegisterDuplicateEmail(t *testing.T) {
	cleanupDatabase(t)

	// Register first user
	registerBody := map[string]interface{}{
		"name":     "First User",
		"email":    "duplicate@example.com",
		"password": "password123",
	}

	resp, err := testServer.POST("/api/v1/auth/register", registerBody, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "register failed: %s", string(resp.Body))

	// Try to register with same email
	registerBody2 := map[string]interface{}{
		"name":     "Second User",
		"email":    "duplicate@example.com",
		"password": "password456",
	}

	resp, err = testServer.POST("/api/v1/auth/register", registerBody2, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var registerResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &registerResp)
	require.NoError(t, err)

	assert.False(t, registerResp["success"].(bool))
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
			resp, err := testServer.POST("/api/v1/auth/register", tt.body, "")
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			var registerResp map[string]interface{}
			err = json.Unmarshal(resp.Body, &registerResp)
			require.NoError(t, err)
			assert.False(t, registerResp["success"].(bool))
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
			resp, err := testServer.POST("/api/v1/auth/login", tt.body, "")
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			var loginResp map[string]interface{}
			err = json.Unmarshal(resp.Body, &loginResp)
			require.NoError(t, err)
			assert.False(t, loginResp["success"].(bool))
		})
	}
}

func TestAuth_RefreshTokenInvalid(t *testing.T) {
	cleanupDatabase(t)

	// Try to refresh with invalid token
	refreshBody := map[string]interface{}{
		"refresh_token": "invalid-refresh-token",
	}

	resp, err := testServer.POST("/api/v1/auth/refresh", refreshBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var refreshResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &refreshResp)
	require.NoError(t, err)

	assert.False(t, refreshResp["success"].(bool))
}
