-- +goose Up
CREATE TABLE auth_sessions (
    id INTEGER PRIMARY KEY CHECK (id > 0),
    token_hash BLOB NOT NULL UNIQUE CHECK (length(token_hash) = 32),
    user_id INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL CHECK (expires_at > created_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_auth_sessions_user_id ON auth_sessions(user_id);
CREATE INDEX idx_auth_sessions_expires_at ON auth_sessions(expires_at);

-- +goose Down
DROP TABLE auth_sessions;
