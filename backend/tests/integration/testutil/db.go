package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestDB wraps a database connection for testing
type TestDB struct {
	DB *sqlx.DB
}

// NewTestDB creates a new test database connection
func NewTestDB() (*TestDB, error) {
	dsn := getTestDSN()

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &TestDB{DB: db}, nil
}

// getTestDSN returns the test database connection string
func getTestDSN() string {
	host := getEnv("TEST_POSTGRES_HOST", "localhost")
	port := getEnv("TEST_POSTGRES_PORT", "5433")
	user := getEnv("TEST_POSTGRES_USER", "pos_test_user")
	password := getEnv("TEST_POSTGRES_PASSWORD", "pos_test_password")
	dbName := getEnv("TEST_POSTGRES_DB", "pos_test_db")
	sslMode := getEnv("TEST_POSTGRES_SSLMODE", "disable")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbName, sslMode)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// RunMigrations runs all database migrations
func (t *TestDB) RunMigrations() error {
	migrationsPath := getMigrationsPath()

	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort up migrations
	var upMigrations []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".up.sql") {
			upMigrations = append(upMigrations, f.Name())
		}
	}
	sort.Strings(upMigrations)

	for _, migration := range upMigrations {
		content, err := os.ReadFile(filepath.Join(migrationsPath, migration))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", migration, err)
		}

		_, err = t.DB.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration, err)
		}
	}

	return nil
}

// getMigrationsPath returns the path to migrations directory
func getMigrationsPath() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "..", "..", "..", "migrations")
}

// TruncateTables truncates specified tables
func (t *TestDB) TruncateTables(tables ...string) error {
	for _, table := range tables {
		_, err := t.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}
	return nil
}

// TruncateAllTables truncates all application tables
func (t *TestDB) TruncateAllTables() error {
	tables := []string{
		"payments",
		"order_items",
		"orders",
		"product_variants",
		"products",
		"product_categories",
		"customers",
		"user_sessions",
		"user_branches",
		"users",
		"role_permissions",
		"permissions",
		"roles",
		"branches",
		"companies",
	}
	return t.TruncateTables(tables...)
}

// Close closes the database connection
func (t *TestDB) Close() error {
	return t.DB.Close()
}

// BeginTx starts a transaction for test isolation
func (t *TestDB) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return t.DB.BeginTxx(ctx, nil)
}
