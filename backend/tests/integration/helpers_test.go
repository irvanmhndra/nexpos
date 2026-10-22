package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/irvanmhndra/nexpos-api/tests/testutil"
	"github.com/stretchr/testify/require"
)

func doPost(t *testing.T, path string, body any, token string) *testutil.Response {
	t.Helper()
	resp, err := testEnv.Server.POST(path, body, token)
	require.NoError(t, err)
	return resp
}

func doGet(t *testing.T, path string, token string) *testutil.Response {
	t.Helper()
	resp, err := testEnv.Server.GET(path, token)
	require.NoError(t, err)
	return resp
}

func doPut(t *testing.T, path string, body any, token string) *testutil.Response {
	t.Helper()
	resp, err := testEnv.Server.PUT(path, body, token)
	require.NoError(t, err)
	return resp
}

func doDelete(t *testing.T, path string, token string) *testutil.Response {
	t.Helper()
	resp, err := testEnv.Server.DELETE(path, token)
	require.NoError(t, err)
	return resp
}

func decodeResponse(t *testing.T, resp *testutil.Response) testutil.APIResponse {
	t.Helper()
	r, err := resp.ParseResponse()
	require.NoError(t, err)
	return *r
}

// registerTestUser creates a unique test user and returns the auth context.
func registerTestUser(t *testing.T) *testutil.AuthContext {
	t.Helper()
	n := testutil.UniqueCounter()
	return registerUser(t,
		fmt.Sprintf("Test User %d", n),
		fmt.Sprintf("test%d@example.com", n),
		"password123",
	)
}

// registerUser registers a user with specific credentials and returns the auth context.
func registerUser(t *testing.T, name, email, password string) *testutil.AuthContext {
	t.Helper()
	resp := doPost(t, "/api/v1/auth/register", map[string]any{
		"name":     name,
		"email":    email,
		"password": password,
	}, "")
	require.Equal(t, http.StatusCreated, resp.StatusCode, "register failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var data struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			ID      int64 `json:"id"`
			Company struct {
				ID int64 `json:"id"`
			} `json:"company"`
		} `json:"user"`
	}
	require.NoError(t, json.Unmarshal(r.Data, &data))

	// Look up branch ID (register creates an "HQ" branch)
	var branchID int64
	err := testEnv.DB.QueryRow(
		"SELECT id FROM branches WHERE company_id = $1 AND code = 'HQ'",
		data.User.Company.ID,
	).Scan(&branchID)
	require.NoError(t, err, "failed to look up branch ID")

	return &testutil.AuthContext{
		Token:        data.AccessToken,
		RefreshToken: data.RefreshToken,
		UserID:       data.User.ID,
		CompanyID:    data.User.Company.ID,
		BranchID:     branchID,
	}
}
