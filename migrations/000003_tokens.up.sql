CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY,
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT UNIQUE NOT NULL,          -- SHA-256 of the actual token
    app_id      UUID REFERENCES registered_apps(id),
    user_agent  TEXT,
    ip_address  INET,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked     BOOLEAN DEFAULT FALSE,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- index for fast lookup
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);