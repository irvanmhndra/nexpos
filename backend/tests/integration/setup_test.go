package integration

import (
	"log"
	"os"
	"testing"

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
