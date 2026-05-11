-- name: UpdateUser :one
update users
set 
hashed_password = $1, email = $2
where id = $3
RETURNING *;