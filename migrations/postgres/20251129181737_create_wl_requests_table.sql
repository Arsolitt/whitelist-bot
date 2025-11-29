-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wl_requests (
    id UUID PRIMARY KEY NOT NULL,
    requester_id UUID NOT NULL,
    nickname TEXT NOT NULL,
    status TEXT NOT NULL,
    decline_reason TEXT NOT NULL,
    arbiter_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS wl_requests;
DROP INDEX IF EXISTS idx_wl_requests_requester_id;
-- +goose StatementEnd
