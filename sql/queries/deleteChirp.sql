-- name: DeleteChirpById :one
delete from chirps
where id = $1
RETURNING *;