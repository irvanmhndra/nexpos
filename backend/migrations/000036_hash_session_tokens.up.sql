-- Store only SHA-256 digests of session tokens, so a leaked user_sessions
-- table (backup, replica, SQL injection) cannot be replayed as live sessions.
-- Existing rows are hashed in place and the API hashes each presented token
-- the same way (hex SHA-256 of the token text), so nobody is signed out.
ALTER TABLE user_sessions RENAME COLUMN access_token TO access_token_hash;
ALTER TABLE user_sessions RENAME COLUMN refresh_token TO refresh_token_hash;

UPDATE user_sessions SET
    access_token_hash  = encode(sha256(convert_to(access_token_hash, 'UTF8')), 'hex'),
    refresh_token_hash = encode(sha256(convert_to(refresh_token_hash, 'UTF8')), 'hex');

ALTER INDEX idx_user_sessions_access_token RENAME TO idx_user_sessions_access_token_hash;
ALTER INDEX idx_user_sessions_refresh_token RENAME TO idx_user_sessions_refresh_token_hash;

COMMENT ON TABLE user_sessions IS 'Session tokens are stored as hex SHA-256 digests; raw tokens exist only on the client.';
