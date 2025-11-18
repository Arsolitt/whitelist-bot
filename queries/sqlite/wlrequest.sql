-- WLRequest Queries
--
-- name: WLRequestByRequesterID :one
SELECT * FROM wl_requests
WHERE requester_id = :requester_id;

-- name: CreateWLRequest :one
INSERT INTO wl_requests (id, requester_id, nickname, status, decline_reason, arbiter_id, created_at, updated_at)
VALUES (:id, :requester_id, :nickname, :status, :decline_reason, :arbiter_id, :created_at, :updated_at)
RETURNING *;

-- name: UpdateWLRequest :one
UPDATE wl_requests
SET requester_id = :requester_id, nickname = :nickname, status = :status, decline_reason = :decline_reason, arbiter_id = :arbiter_id, updated_at = :updated_at
WHERE id = :id
RETURNING *;

-- name: WLRequestByID :one
SELECT * FROM wl_requests
WHERE id = :id;

-- name: PendingWLRequests :many
SELECT * FROM wl_requests
WHERE status = 'pending';

-- name: PendingWLRequest :one
SELECT * FROM wl_requests
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT 1;
