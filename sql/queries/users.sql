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

-- name: UpdateDetails :one
UPDATE users
SET email = $2, hashed_password = $3, updated_at = NOW()
WHERE id = $1 
RETURNING *;

-- name: UpgradeMembership :one
UPDATE users 
SET is_chirpy_red = TRUE 
WHERE id = $1
RETURNING *;