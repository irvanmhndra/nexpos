-- The platform panel: internal staff who look after every company, so nobody
-- needs direct database access for support work.
--
-- Everything lives in new tables. companies and user_sessions get no new
-- columns, because the Go API reads them with SELECT * into fixed structs.
-- Suspending a company uses the existing companies.is_active.

CREATE TABLE platform_users (
    id            BIGSERIAL PRIMARY KEY,
    email         VARCHAR(255) NOT NULL,
    name          VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(20)  NOT NULL,
    -- false: only the companies listed in platform_user_companies
    all_companies BOOLEAN      NOT NULL DEFAULT false,
    is_active     BOOLEAN      NOT NULL DEFAULT true,
    failed_logins INT          NOT NULL DEFAULT 0,
    locked_until  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_platform_users_role CHECK (role IN ('admin', 'ops', 'support'))
);
CREATE UNIQUE INDEX idx_platform_users_email ON platform_users (LOWER(email));

CREATE TABLE platform_user_companies (
    platform_user_id BIGINT NOT NULL REFERENCES platform_users(id) ON DELETE CASCADE,
    company_id       BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    PRIMARY KEY (platform_user_id, company_id)
);

-- Tokens are stored as hex SHA-256 digests, like user_sessions.
CREATE TABLE platform_sessions (
    id               BIGSERIAL PRIMARY KEY,
    platform_user_id BIGINT       NOT NULL REFERENCES platform_users(id) ON DELETE CASCADE,
    token_hash       VARCHAR(64)  NOT NULL,
    expires_at       TIMESTAMPTZ  NOT NULL,
    is_revoked       BOOLEAN      NOT NULL DEFAULT false,
    ip_address       VARCHAR(64),
    user_agent       TEXT,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    last_used_at     TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_platform_sessions_token_hash ON platform_sessions (token_hash);

-- Marks a tenant session as opened by staff. Such a session is read-only.
CREATE TABLE platform_impersonations (
    session_id       BIGINT      PRIMARY KEY REFERENCES user_sessions(id) ON DELETE CASCADE,
    platform_user_id BIGINT      NOT NULL REFERENCES platform_users(id),
    company_id       BIGINT      NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    reason           TEXT        NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Append-only. It keeps the actor's email and plain ids rather than foreign
-- keys, so a log line outlives the staff account or company it is about.
CREATE TABLE platform_audit_logs (
    id               BIGSERIAL PRIMARY KEY,
    platform_user_id BIGINT       NOT NULL,
    actor_email      VARCHAR(255) NOT NULL,
    action           VARCHAR(50)  NOT NULL,
    company_id       BIGINT,
    target_type      VARCHAR(50)  NOT NULL,
    target_id        BIGINT,
    reason           TEXT         NOT NULL,
    before           JSONB,
    after            JSONB,
    ip_address       VARCHAR(64),
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_platform_audit_logs_created_at ON platform_audit_logs (created_at DESC);
CREATE INDEX idx_platform_audit_logs_company ON platform_audit_logs (company_id, created_at DESC);
