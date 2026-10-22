CREATE TABLE IF NOT EXISTS user_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_id BIGINT NOT NULL REFERENCES companies(id),
    branch_id BIGINT REFERENCES branches(id),

    -- Raw tokens (EARLY PHASE - refactor before production)
    access_token TEXT NOT NULL,
    access_token_expires_at TIMESTAMPTZ NOT NULL,

    refresh_token TEXT NOT NULL,
    refresh_token_expires_at TIMESTAMPTZ NOT NULL,

    is_revoked BOOLEAN NOT NULL DEFAULT false,
    device_info TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_access_token ON user_sessions(access_token) WHERE NOT is_revoked;
CREATE INDEX idx_user_sessions_refresh_token ON user_sessions(refresh_token) WHERE NOT is_revoked;
CREATE INDEX idx_user_sessions_company_id ON user_sessions(company_id);

COMMENT ON TABLE user_sessions IS 'WARNING: Raw tokens stored for early-phase debugging. MUST refactor to hashed tokens before production.';
