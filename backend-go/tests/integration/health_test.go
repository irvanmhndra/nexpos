package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealth_Check(t *testing.T) {
	resp := doGet(t, "/health", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r := decodeResponse(t, resp)
	assert.True(t, r.Success)
}

func TestHealth_Liveness(t *testing.T) {
	resp := doGet(t, "/health/live", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHealth_Readiness(t *testing.T) {
	resp := doGet(t, "/health/ready", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHealth_RootEndpoint(t *testing.T) {
	// Root endpoint is not registered; expect 404 or 405
	resp := doGet(t, "/", "")
	assert.True(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed)
}
