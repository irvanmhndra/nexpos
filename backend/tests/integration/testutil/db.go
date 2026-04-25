package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDB wraps a database connection for testing
type TestDB struct {
	DB        *sqlx.DB
	DSN       string
	container *postgres.PostgresContainer
	ctx       context.Context
}

// NewTestDB starts a PostgreSQL testcontainer and returns a connected TestDB.
func NewTestDB(ctx context.Context) (*TestDB, error) {
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:18-alpine"),
		postgres.WithDatabase("pos_test_db"),
		postgres.WithUsername("pos_test_user"),
		postgres.WithPassword("pos_test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pgContainer.Terminate(ctx) // #nosec G104 -- best-effort cleanup on error path
		return nil, fmt.Errorf("get connection string: %w", err)
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		_ = pgContainer.Terminate(ctx) // #nosec G104 -- best-effort cleanup on error path
		return nil, fmt.Errorf("connect to test database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &TestDB{
		DB:        db,
		DSN:       dsn,
		container: pgContainer,
		ctx:       ctx,
	}, nil
}

// RunMigrations runs all database migrations.
// It first drops all existing objects to ensure a clean slate, then applies all up migrations.
func (t *TestDB) RunMigrations() error {
	// Drop all tables/types so migrations can be re-applied cleanly
	_, err := t.DB.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	if err != nil {
		return fmt.Errorf("failed to reset schema: %w", err)
	}

	migrationsDir, err := findMigrationsDir()
	if err != nil {
		return fmt.Errorf("find migrations dir: %w", err)
	}

	m, err := migrate.New("file://"+migrationsDir, t.DSN)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

// findMigrationsDir walks up from the working directory until it finds go.mod,
// then returns the migrations/ directory next to it.
func findMigrationsDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "migrations"), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", dir)
		}
		dir = parent
	}
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

// TruncateAllTables truncates all application tables and re-seeds system data.
// TRUNCATE CASCADE is used for speed and to reset ID sequences.
// System seed data (roles, permissions, role_permissions) is restored afterward.
func (t *TestDB) TruncateAllTables() error {
	_, err := t.DB.Exec(`
		TRUNCATE TABLE
			stock_movements, stocks, payments, order_items, orders,
			promotions, company_settings,
			product_variants, products, product_categories,
			customers, user_sessions, user_branches, users,
			branches, companies,
			role_permissions, roles, permissions
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	// Re-seed system roles
	_, err = t.DB.Exec(`
		INSERT INTO roles (company_id, code, name, description, is_system) VALUES
		(NULL, 'owner', 'Owner', 'Full access to all features', true),
		(NULL, 'admin', 'Administrator', 'Can manage users, products, inventory, and daily operations', true),
		(NULL, 'staff', 'Staff', 'Basic access for daily operations (cashier, inventory)', true)
	`)
	if err != nil {
		return fmt.Errorf("failed to re-seed roles: %w", err)
	}

	// Re-seed permissions
	_, err = t.DB.Exec(`
		INSERT INTO permissions (code, name, module, description) VALUES
		('users.view', 'View Users', 'users', 'Can view user list and details'),
		('users.create', 'Create Users', 'users', 'Can create new users'),
		('users.update', 'Update Users', 'users', 'Can update user information'),
		('users.delete', 'Delete Users', 'users', 'Can delete users'),
		('products.view', 'View Products', 'products', 'Can view product list and details'),
		('products.create', 'Create Products', 'products', 'Can create new products'),
		('products.update', 'Update Products', 'products', 'Can update product information'),
		('products.delete', 'Delete Products', 'products', 'Can delete products'),
		('inventory.view', 'View Inventory', 'inventory', 'Can view stock levels'),
		('inventory.adjust', 'Adjust Inventory', 'inventory', 'Can adjust stock quantities'),
		('orders.view', 'View Orders', 'orders', 'Can view order list and details'),
		('orders.create', 'Create Orders', 'orders', 'Can create new orders (cashier)'),
		('orders.void', 'Void Orders', 'orders', 'Can void/cancel orders'),
		('orders.refund', 'Refund Orders', 'orders', 'Can process refunds'),
		('reports.view', 'View Reports', 'reports', 'Can view reports and analytics'),
		('reports.export', 'Export Reports', 'reports', 'Can export reports'),
		('settings.view', 'View Settings', 'settings', 'Can view company settings'),
		('settings.update', 'Update Settings', 'settings', 'Can update company settings'),
		('branches.view', 'View Branches', 'branches', 'Can view branch list'),
		('branches.manage', 'Manage Branches', 'branches', 'Can create/update/delete branches')
	`)
	if err != nil {
		return fmt.Errorf("failed to re-seed permissions: %w", err)
	}

	// Re-seed role_permissions
	_, err = t.DB.Exec(`
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id FROM roles r, permissions p WHERE r.code = 'owner';

		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id FROM roles r, permissions p
		WHERE r.code = 'admin' AND p.code NOT IN ('settings.update', 'users.delete');

		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id FROM roles r, permissions p
		WHERE r.code = 'staff' AND p.code IN ('products.view', 'inventory.view', 'orders.view', 'orders.create')
	`)
	if err != nil {
		return fmt.Errorf("failed to re-seed role_permissions: %w", err)
	}

	return nil
}

// Close closes the database connection and terminates the container.
func (t *TestDB) Close() error {
	if err := t.DB.Close(); err != nil {
		return err
	}
	return t.container.Terminate(t.ctx)
}

// BeginTx starts a transaction for test isolation
func (t *TestDB) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return t.DB.BeginTxx(ctx, nil)
}

var counter atomic.Int64

// UniqueCounter returns a monotonically increasing integer for unique test data.
func UniqueCounter() int64 {
	return counter.Add(1)
}
