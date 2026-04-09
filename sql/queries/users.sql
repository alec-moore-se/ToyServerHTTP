-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(), NOW(), NOW(), $1, $2
)
RETURNING id, created_at, updated_at, email, is_chirpy_red;

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

-- name: GetChirpByUser :one
SELECT * FROM chirps WHERE user_id = $1 ORDER BY created_at ASC;

-- name: GetUserwEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserIDwToken :one
SELECT user_id FROM refresh_tokens WHERE token = $1;

-- name: UpdateUser :exec 
UPDATE users SET hashed_password = $2 WHERE id = $1;

-- name: UpdateUserEmailaPass :exec 
UPDATE users SET email = $2, hashed_password = $3 WHERE id = $1;

-- name: DeleteChirp :exec 
DELETE FROM chirps WHERE id = $1;

-- name: PutUserChirpyRed :exec 
UPDATE users SET is_chirpy_red = TRUE WHERE id = $1;

-- name: PutUserChirpyRedRemove :exec 
UPDATE users SET is_chirpy_red = FALSE WHERE id = $1;
