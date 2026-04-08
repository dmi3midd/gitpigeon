-- +goose Up
CREATE TABLE IF NOT EXISTS repositories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    last_seen_tag VARCHAR(255),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(owner, name)
);

CREATE INDEX idx_repositories_owner_name ON repositories(owner, name);

-- +goose Down
DROP TABLE IF EXISTS repositories;
