-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    $1, 
    NOW(), 
    NOW(), 
    $2, 
    $3
)
RETURNING *;

-- name: ResetUsers :exec
TRUNCATE TABLE users, chirps CASCADE;

-- name: GetUser :one
SELECT * FROM users WHERE email = $1;