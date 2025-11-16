-- +goose Up
-- +goose StatementBegin
-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY AUTOINCREMENT,
    telegram_id INTEGER UNIQUE NOT NULL,
    username TEXT,
    first_name TEXT,
    last_name TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Whitelist requests table
CREATE TABLE IF NOT EXISTS whitelist_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    nickname TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    admin_note TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Whitelist entries table
CREATE TABLE IF NOT EXISTS whitelist_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    nickname TEXT NOT NULL,
    added_by INTEGER NOT NULL,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    removed_at DATETIME,
    is_active BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (request_id) REFERENCES whitelist_requests(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (added_by) REFERENCES users(id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_requests_user_id ON whitelist_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_requests_status ON whitelist_requests(status);
CREATE INDEX IF NOT EXISTS idx_requests_created_at ON whitelist_requests(created_at);
CREATE INDEX IF NOT EXISTS idx_whitelist_user_id ON whitelist_entries(user_id);
CREATE INDEX IF NOT EXISTS idx_whitelist_is_active ON whitelist_entries(is_active);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_whitelist_is_active;
DROP INDEX IF EXISTS idx_whitelist_user_id;
DROP INDEX IF EXISTS idx_requests_created_at;
DROP INDEX IF EXISTS idx_requests_status;
DROP INDEX IF EXISTS idx_requests_user_id;
DROP TABLE IF EXISTS whitelist_entries;
DROP TABLE IF EXISTS whitelist_requests;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
