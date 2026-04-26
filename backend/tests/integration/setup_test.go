package integration_test

import (
	"context"
	"log"
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
