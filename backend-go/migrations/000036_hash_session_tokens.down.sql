-- Digests cannot be turned back into tokens: revoke every session so users
-- sign in again after the rollback.
UPDATE user_sessions SET is_revoked = true;

ALTER INDEX idx_user_sessions_access_token_hash RENAME TO idx_user_sessions_access_token;
ALTER INDEX idx_user_sessions_refresh_token_hash RENAME TO idx_user_sessions_refresh_token;

ALTER TABLE user_sessions RENAME COLUMN access_token_hash TO access_token;
ALTER TABLE user_sessions RENAME COLUMN refresh_token_hash TO refresh_token;

COMMENT ON TABLE user_sessions IS 'WARNING: Raw tokens stored for early-phase debugging. MUST refactor to hashed tokens before production.';
