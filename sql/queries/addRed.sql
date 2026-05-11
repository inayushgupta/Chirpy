-- name: UpdateRedToUser :one
update users
set is_chirpy_red = true
where id = $1
RETURNING *;