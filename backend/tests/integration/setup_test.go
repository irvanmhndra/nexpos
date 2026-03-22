package integration

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/irvanmhndra/nexpos-api/config"
	"github.com/irvanmhndra/nexpos-api/internal/app"
	"github.com/irvanmhndra/nexpos-api/tests/integration/testutil"
)

var (
	testDB      *testutil.TestDB
	testApp     *app.App
	testServer  *testutil.TestServer
	testFixture *testutil.Fixtures
)

func TestMain(m *testing.M) {
	// Setup
	if err := setup(); err != nil {
		log.Fatalf("Failed to setup tests: %v", err)
	}

	// Run tests
	code := m.Run()

	// Teardown
	teardown()

	os.Exit(code)
}

func setup() error {
	// Create test database connection
	db, err := testutil.NewTestDB()
	if err != nil {
		return err
	}
	testDB = db

	// Run migrations
	if err := testDB.RunMigrations(); err != nil {
		return err
	}

	// Create test config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: "8081",
			Env:  "test",
		},
		Postgres: config.PostgresConfig{
			Host:     getEnv("TEST_POSTGRES_HOST", "localhost"),
			Port:     getEnv("TEST_POSTGRES_PORT", "5433"),
			User:     getEnv("TEST_POSTGRES_USER", "pos_test_user"),
			Password: getEnv("TEST_POSTGRES_PASSWORD", "pos_test_password"),
			DB:       getEnv("TEST_POSTGRES_DB", "pos_test_db"),
			SSLMode:  "disable",
		},
		JWT: config.JWTConfig{
			Secret:             "test-secret-key-for-integration-tests",
			AccessExpiresHours: 2,
			RefreshExpiresDays: 7,
			AccessTokenExpiry:  2 * time.Hour,
			RefreshTokenExpiry: 7 * 24 * time.Hour,
		},
	}

	// Create app
	testApp, err = app.New(cfg)
	if err != nil {
		return err
	}

	// Create test server
	testServer = testutil.NewTestServer(testApp.Echo())

	// Create fixtures utility
	testFixture = testutil.NewFixtures(testDB.DB)

	return nil
}

func teardown() {
	if testApp != nil {
		testApp.Close()
	}
	if testDB != nil {
		_ = testDB.Close()
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// cleanupDatabase truncates all tables before each test
func cleanupDatabase(t *testing.T) {
	t.Helper()
	if err := testDB.TruncateAllTables(); err != nil {
		t.Fatalf("Failed to cleanup database: %v", err)
	}
}

var counter atomic.Int64

// uniqueCounter returns a monotonically increasing integer for unique test data.
func uniqueCounter() int64 {
	return counter.Add(1)
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
// This is used by integration tests that need to access protected routes.
// The register API creates a company, branch, user, and session automatically.
func registerTestUser(t *testing.T) *authContext {
	t.Helper()

	registerBody := map[string]interface{}{
		"name":     "Integration Test User",
		"email":    "integration@example.com",
		"password": "password123",
	}

	resp, err := testServer.POST("/api/v1/auth/register", registerBody, "")
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
	err = testDB.DB.QueryRow(
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
