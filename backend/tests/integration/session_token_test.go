package integration_test

import (
	"net/http"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/irvanmhndra/nexpos-api/pkg/sessiontoken"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth_SessionsStoreOnlyTokenDigests(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	var digests, raw int
	require.NoError(t, testEnv.DB.Get(&digests,
		`SELECT count(*) FROM user_sessions WHERE access_token_hash = $1 AND refresh_token_hash = $2`,
		sessiontoken.Hash(auth.Token), sessiontoken.Hash(auth.RefreshToken)))
	require.NoError(t, testEnv.DB.Get(&raw,
		`SELECT count(*) FROM user_sessions WHERE access_token_hash IN ($1, $2) OR refresh_token_hash IN ($1, $2)`,
		auth.Token, auth.RefreshToken))
	assert.Equal(t, 1, digests, "session must be stored by digest")
	assert.Zero(t, raw, "raw tokens must never be stored")

	// A stored digest is not itself a credential
	resp := doGet(t, "/api/v1/branches", sessiontoken.Hash(auth.Token))
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp = doGet(t, "/api/v1/branches", auth.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuth_RefreshTokenIsSingleUse(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	resp := doPost(t, "/api/v1/auth/refresh", map[string]any{"refresh_token": auth.RefreshToken}, "")
	require.Equal(t, http.StatusOK, resp.StatusCode, "refresh failed: %s", string(resp.Body))

	// Replaying the exchanged token fails, and the old access token is revoked
	resp = doPost(t, "/api/v1/auth/refresh", map[string]any{"refresh_token": auth.RefreshToken}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp = doGet(t, "/api/v1/branches", auth.Token)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// Sessions issued before migration 000036 hold raw tokens; the migration
// hashes them in place so those users stay signed in after the upgrade.
func TestMigration036_KeepsExistingSessionsValid(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	m, err := migrate.New("file://../../migrations", testEnv.TestDB.DSN)
	require.NoError(t, err)
	defer func() { _, _ = m.Close() }()
	require.NoError(t, m.Steps(-1)) // back to raw-token columns

	const access, refresh = "pre-upgrade-access-token", "pre-upgrade-refresh-token"
	_, err = testEnv.DB.Exec(`
		INSERT INTO user_sessions (user_id, company_id, branch_id, access_token, access_token_expires_at,
			refresh_token, refresh_token_expires_at)
		VALUES ($1, $2, $3, $4, NOW() + INTERVAL '1 hour', $5, NOW() + INTERVAL '1 day')`,
		auth.UserID, auth.CompanyID, auth.BranchID, access, refresh)
	require.NoError(t, err)

	require.NoError(t, m.Steps(1)) // apply 000036

	var stored string
	require.NoError(t, testEnv.DB.Get(&stored,
		`SELECT access_token_hash FROM user_sessions WHERE refresh_token_hash = $1`, sessiontoken.Hash(refresh)))
	assert.Equal(t, sessiontoken.Hash(access), stored)

	resp := doGet(t, "/api/v1/branches", access)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "pre-upgrade session must still authenticate")
	resp = doPost(t, "/api/v1/auth/refresh", map[string]any{"refresh_token": refresh}, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "pre-upgrade refresh token must still work")
}
