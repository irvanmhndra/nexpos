CREATE TABLE IF NOT EXISTS oauth_clients (
    id BIGSERIAL PRIMARY KEY,
    client_id VARCHAR(100) NOT NULL,
    client_secret VARCHAR(255),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'public', -- public | confidential
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_oauth_clients_client_id ON oauth_clients(client_id);

-- Insert default public client for web/mobile apps
INSERT INTO oauth_clients (client_id, name, type)
VALUES ('aplikabizz-web', 'Aplikabizz Web Dashboard', 'public');
