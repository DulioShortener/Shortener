-- +goose Up
CREATE TABLE links (
    id INTEGER PRIMARY KEY CHECK (id > 0),
    user_id INTEGER NOT NULL,
    code TEXT NOT NULL COLLATE BINARY UNIQUE CHECK (length(code) = 8),
    target_url TEXT NOT NULL CHECK (length(target_url) BETWEEN 1 AND 4096),
    target_url_key TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (user_id, target_url_key)
);

CREATE INDEX idx_links_user_created_at ON links(user_id, created_at DESC, id DESC);

-- +goose StatementBegin
CREATE TRIGGER enforce_user_link_limit
BEFORE INSERT ON links
WHEN (SELECT COUNT(*) FROM links WHERE user_id = NEW.user_id) >= 50
BEGIN
    SELECT RAISE(ABORT, 'LINK_LIMIT_REACHED');
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER enforce_user_link_limit;
DROP TABLE links;
