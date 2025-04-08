-- name: CreateChirp :one
INSERT INTO chirp(id, created_at, updated_at, body, user_id)
VALUES (
           $1,
           $2,
           $3,
           $4,
           $5
       )
RETURNING *;

-- name: GetAllChirps :many
SELECT * FROM chirp
ORDER BY created_at;

-- name: GetChirpById :one
SELECT * FROM chirp
WHERE id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirp
WHERE id = $1;

-- name: GetChirpsByAuthor :many
SELECT * FROM chirp
WHERE user_id = $1
ORDER BY created_at;