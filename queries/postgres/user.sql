-- User Queries
--
-- name: UserByTelegramID :one
SELECT * FROM users
WHERE telegram_id = $1;

-- name: CreateUser :one
INSERT INTO users (id, telegram_id, first_name, last_name, username, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET telegram_id = $1, first_name = $2, last_name = $3, username = $4, updated_at = $5
WHERE id = $6
RETURNING *;

-- name: UserByID :one
SELECT * FROM users
WHERE id = $1;
