package testutil

import (
	"context"
	"time"

	"github.com/irvanmhndra/nexpos-api/config"
	"github.com/irvanmhndra/nexpos-api/internal/app"
	"github.com/jmoiron/sqlx"
)

// TestEnv holds the shared test infrastructure for integration tests.
type TestEnv struct {
	App       *app.App
	DB        *sqlx.DB
	TestDB    *TestDB
	TestMongo *TestMongo
	Server    *TestServer
	Fixtures  *Fixtures
	Cleanup   func()
}

// Setup starts PostgreSQL + MongoDB testcontainers, runs migrations, and boots
// the full app. Call env.Cleanup() when done.
func Setup(ctx context.Context) (*TestEnv, error) {
	tdb, err := NewTestDB(ctx)
	if err != nil {
		return nil, err
	}

	if err := tdb.RunMigrations(); err != nil {
		_ = tdb.Close()
		return nil, err
	}

	tmongo, err := NewTestMongo(ctx)
	if err != nil {
		_ = tdb.Close()
		return nil, err
	}

	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: "8081",
			Env:  "test",
		},
		Postgres: config.PostgresConfig{
			DSNOverride: tdb.DSN,
		},
		Mongo: config.MongoConfig{
			URI:      tmongo.URI,
			Database: "nexpos_test",
		},
		JWT: config.JWTConfig{
			Secret:             "test-secret-key-for-integration-tests",
			AccessExpiresHours: 2,
			RefreshExpiresDays: 7,
			AccessTokenExpiry:  2 * time.Hour,
			RefreshTokenExpiry: 7 * 24 * time.Hour,
		},
	}

	a, err := app.New(cfg)
	if err != nil {
		_ = tdb.Close()
		_ = tmongo.Close()
		return nil, err
	}

	return &TestEnv{
		App:       a,
		DB:        tdb.DB,
		TestDB:    tdb,
		TestMongo: tmongo,
		Fixtures:  NewFixtures(tdb.DB),
		Cleanup: func() {
			a.Close()
			_ = tdb.Close()
			_ = tmongo.Close()
		},
	}, nil
}

// SetupTestEnv creates a full test environment including an HTTP test server.
func SetupTestEnv(ctx context.Context) (*TestEnv, error) {
	env, err := Setup(ctx)
	if err != nil {
		return nil, err
	}

	server := NewTestServer(env.App.Echo())
	env.Server = server

	origCleanup := env.Cleanup
	env.Cleanup = func() {
		server.Close()
		origCleanup()
	}

	return env, nil
}
