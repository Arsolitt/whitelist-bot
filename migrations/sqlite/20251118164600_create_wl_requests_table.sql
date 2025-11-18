-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS whitelist_requests (
    id TEXT PRIMARY KEY NOT NULL,
    requester_id TEXT NOT NULL,
    nickname TEXT NOT NULL,
    status TEXT NOT NULL,
    decline_reason TEXT,
    arbiter_id TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_whitelist_requests_requester_id ON whitelist_requests(requester_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS whitelist_requests;
-- +goose StatementEnd
