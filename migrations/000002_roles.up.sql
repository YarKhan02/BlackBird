CREATE TABLE global_roles (
    id   SMALLINT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL    -- super_admin, admin, user, banned
);

CREATE TABLE user_global_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id SMALLINT REFERENCES global_roles(id),
    PRIMARY KEY (user_id, role_id)
);

-- per-app roles stored by each app, but auth server knows the mapping
CREATE TABLE registered_apps (
    id            UUID PRIMARY KEY,
    client_id     TEXT UNIQUE NOT NULL,
    client_secret TEXT NOT NULL,               -- bcrypt hashed
    name          TEXT NOT NULL,
    redirect_uris TEXT[] NOT NULL,
    is_active     BOOLEAN DEFAULT TRUE,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE user_app_roles (
    user_id   UUID REFERENCES users(id) ON DELETE CASCADE,
    app_id    UUID REFERENCES registered_apps(id) ON DELETE CASCADE,
    role      TEXT NOT NULL,                   -- app defines what this means
    PRIMARY KEY (user_id, app_id, role)
);