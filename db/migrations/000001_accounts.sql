CREATE TABLE IF NOT EXISTS users (
    id text PRIMARY KEY,
    username varchar(24) NOT NULL,
    username_key varchar(24) NOT NULL UNIQUE,
    password_hash text NOT NULL,
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_username_format CHECK (username ~ '^[A-Za-z0-9_]{3,24}$')
);

CREATE INDEX IF NOT EXISTS users_presence_order_idx
    ON users (last_seen_at DESC, username_key, id);

CREATE TABLE IF NOT EXISTS sessions (
    token_hash bytea PRIMARY KEY,
    user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS sessions_user_id_idx ON sessions (user_id);
CREATE INDEX IF NOT EXISTS sessions_expires_at_idx ON sessions (expires_at);
