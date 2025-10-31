-- name: Create :one
INSERT INTO users (email, password_hash, name) 
VALUES ($1, $2, $3) 
RETURNING *;

-- name: GetByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: Delete :exec
DELETE FROM users WHERE id = $1;
