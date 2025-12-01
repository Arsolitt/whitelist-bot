-- WLRequest Queries
--
-- name: WLRequestByRequesterID :one
SELECT * FROM wl_requests
WHERE requester_id = $1;

-- name: CreateWLRequest :one
INSERT INTO wl_requests (id, requester_id, nickname, status, decline_reason, arbiter_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateWLRequest :one
UPDATE wl_requests
SET requester_id = $1, nickname = $2, status = $3, decline_reason = $4, arbiter_id = $5, updated_at = $6
WHERE id = $7
RETURNING *;

-- name: WLRequestByID :one
SELECT * FROM wl_requests
WHERE id = $1;

-- name: PendingWLRequests :many
SELECT * FROM wl_requests
WHERE status = 'pending'
LIMIT sqlc.arg('limit')::bigint;

-- name: PendingWLRequestsWithRequester :many
SELECT sqlc.embed(wl_requests), sqlc.embed(users) FROM wl_requests
JOIN users ON wl_requests.requester_id = users.id
WHERE status = 'pending'
LIMIT sqlc.arg('limit')::bigint;
