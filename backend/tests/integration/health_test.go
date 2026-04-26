package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth_Check(t *testing.T) {
	resp, err := testEnv.Server.GET("/health", "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &healthResp)
	require.NoError(t, err)

	assert.True(t, healthResp["success"].(bool))
}

func TestHealth_Liveness(t *testing.T) {
	resp, err := testEnv.Server.GET("/health/live", "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHealth_Readiness(t *testing.T) {
	resp, err := testEnv.Server.GET("/health/ready", "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHealth_RootEndpoint(t *testing.T) {
	// Root endpoint is not registered; expect 404 or 405
	resp, err := testEnv.Server.GET("/", "")
	require.NoError(t, err)
	assert.True(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed)
}
