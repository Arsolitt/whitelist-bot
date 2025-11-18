-- User Queries
--
-- name: UserByTelegramID :one
SELECT * FROM users
WHERE telegram_id = :telegram_id;

-- name: CreateUser :one
INSERT INTO users (id, telegram_id, first_name, last_name, username, created_at, updated_at)
VALUES (:id, :telegram_id, :first_name, :last_name, :username, :created_at, :updated_at)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET telegram_id = :telegram_id, first_name = :first_name, last_name = :last_name, username = :username, updated_at = :updated_at
WHERE id = :id
RETURNING *;

-- name: UserByID :one
SELECT * FROM users
WHERE id = :id;

-- name: AllUsers :many
SELECT * FROM users;
