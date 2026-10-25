-- +goose Up
CREATE TABLE review_comments (
    id INTEGER PRIMARY KEY CHECK (id > 0),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id INTEGER REFERENCES review_comments(id) ON DELETE CASCADE,
    body TEXT NOT NULL CHECK (length(body) BETWEEN 1 AND 2000 AND length(trim(body)) > 0),
    rating INTEGER,
    created_at INTEGER NOT NULL,
    deleted_at INTEGER,
    CHECK (
        (parent_id IS NULL AND rating IS NOT NULL AND rating BETWEEN 1 AND 5)
        OR (parent_id IS NOT NULL AND rating IS NULL)
    ),
    CHECK (parent_id IS NULL OR parent_id <> id)
);

CREATE UNIQUE INDEX idx_review_comments_active_user
    ON review_comments(user_id)
    WHERE parent_id IS NULL AND deleted_at IS NULL;

CREATE INDEX idx_review_comments_roots
    ON review_comments(created_at DESC, id DESC)
    WHERE parent_id IS NULL AND deleted_at IS NULL;

CREATE INDEX idx_review_comments_children
    ON review_comments(parent_id, created_at, id);

-- +goose Down
DROP TABLE review_comments;
