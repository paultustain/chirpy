-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
    $1, 
    NOW(), 
    NOW(), 
    $2
)
RETURNING *;

-- name: ResetUsers :exec
TRUNCATE TABLE users, chirps CASCADE;