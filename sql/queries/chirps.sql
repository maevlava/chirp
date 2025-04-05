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