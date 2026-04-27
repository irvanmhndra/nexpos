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

// cleanupDatabase truncates all PostgreSQL tables and drops all MongoDB
// collections before each test, so every test starts with a clean state.
func cleanupDatabase(t *testing.T) {
	t.Helper()
	if err := testEnv.TestDB.TruncateAllTables(); err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}
	if err := testEnv.TestMongo.DropCollections("nexpos_test"); err != nil {
		t.Fatalf("Failed to drop mongo collections: %v", err)
	}
}
