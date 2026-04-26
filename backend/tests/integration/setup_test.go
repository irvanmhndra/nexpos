package integration

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/irvanmhndra/nexpos-api/tests/testutil"
)

var testEnv *testutil.TestEnv

func TestMain(m *testing.M) {
	ctx := context.Background()

	env, err := testutil.SetupTestEnv(ctx)
	if err != nil {
		log.Fatalf("Failed to setup tests: %v", err)
	}
	testEnv = env

	code := m.Run()

	env.Cleanup()
	os.Exit(code)
}

// cleanupDatabase truncates all tables before each test
func cleanupDatabase(t *testing.T) {
	t.Helper()
	if err := testEnv.TestDB.TruncateAllTables(); err != nil {
		t.Fatalf("Failed to cleanup database: %v", err)
	}
}

// authContext holds the authenticated user's context for integration tests.
type authContext struct {
	Token        string
	RefreshToken string
	UserID       int64
	CompanyID    int64
	BranchID     int64
}

// registerTestUser registers a new user via the API and returns the auth context.
func registerTestUser(t *testing.T) *authContext {
	t.Helper()

	registerBody := map[string]interface{}{
		"name":     "Integration Test User",
		"email":    "integration@example.com",
		"password": "password123",
	}

	resp, err := testEnv.Server.POST("/api/v1/auth/register", registerBody, "")
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to register test user: status %d, body: %s", resp.StatusCode, string(resp.Body))
	}

	var registerResp map[string]interface{}
	if err := json.Unmarshal(resp.Body, &registerResp); err != nil {
		t.Fatalf("Failed to parse register response: %v", err)
	}

	data := registerResp["data"].(map[string]interface{})
	user := data["user"].(map[string]interface{})
	company := user["company"].(map[string]interface{})

	// Look up branch ID (register creates an "HQ" branch)
	var branchID int64
	companyID := int64(company["id"].(float64))
	err = testEnv.DB.QueryRow(
		"SELECT id FROM branches WHERE company_id = $1 AND code = 'HQ'",
		companyID,
	).Scan(&branchID)
	if err != nil {
		t.Fatalf("Failed to look up branch ID: %v", err)
	}

	return &authContext{
		Token:        data["access_token"].(string),
		RefreshToken: data["refresh_token"].(string),
		UserID:       int64(user["id"].(float64)),
		CompanyID:    companyID,
		BranchID:     branchID,
	}
}
