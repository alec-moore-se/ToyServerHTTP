-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
    gen_random_uuid(), NOW(), NOW(), $1
)
RETURNING *;

-- name: ResetUsers :exec
TRUNCATE TABLE users CASCADE;

-- name: CreateChirp :one 
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(), NOW(), NOW(), $1, $2
)
RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps ORDER BY created_at ASC;

-- name: GetChirp :one
SELECT * FROM chirps WHERE id = $1;

-- name: GetUserwEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :exec 
UPDATE users SET hashed_password = $2 WHERE id = $1;
