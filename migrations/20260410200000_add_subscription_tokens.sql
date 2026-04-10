-- +goose Up
ALTER TABLE subscriptions ADD COLUMN confirmed BOOLEAN NOT NULL DEFAULT 0;
ALTER TABLE subscriptions ADD COLUMN confirm_token VARCHAR(64) NOT NULL DEFAULT '';
ALTER TABLE subscriptions ADD COLUMN unsubscribe_token VARCHAR(64) NOT NULL DEFAULT '';

CREATE UNIQUE INDEX idx_subscriptions_confirm_token ON subscriptions(confirm_token);
CREATE UNIQUE INDEX idx_subscriptions_unsubscribe_token ON subscriptions(unsubscribe_token);

-- +goose Down
DROP INDEX IF EXISTS idx_subscriptions_unsubscribe_token;
DROP INDEX IF EXISTS idx_subscriptions_confirm_token;

CREATE TABLE subscriptions_backup AS SELECT id, repository_id, email, created_at FROM subscriptions;
DROP TABLE subscriptions;
CREATE TABLE subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repository_id INTEGER NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    UNIQUE(repository_id, email)
);
INSERT INTO subscriptions SELECT * FROM subscriptions_backup;
DROP TABLE subscriptions_backup;
