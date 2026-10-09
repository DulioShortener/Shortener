-- +goose Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY CHECK (id > 0),
    username TEXT NOT NULL COLLATE NOCASE UNIQUE
        CHECK (length(username) BETWEEN 2 AND 32)
        CHECK (username = lower(username))
        CHECK (username NOT GLOB '*[^a-z0-9._]*')
        CHECK (instr(username, '..') = 0),
    display_name TEXT
        CHECK (
            display_name IS NULL OR
            (length(trim(display_name)) BETWEEN 2 AND 32)
        ),
    password_hash TEXT NOT NULL CHECK (length(password_hash) > 0),
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- +goose Down
DROP TABLE users;
