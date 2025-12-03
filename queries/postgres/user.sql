-- User Queries
--
-- name: UserByTelegramID :one
SELECT * FROM users
WHERE telegram_id = $1;

-- name: CreateUser :one
INSERT INTO users (id, telegram_id, chat_id, first_name, last_name, username, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET telegram_id = $1, chat_id = $2, first_name = $3, last_name = $4, username = $5, updated_at = $6
WHERE id = $7
RETURNING *;

-- name: UserByID :one
SELECT * FROM users
WHERE id = $1;
