-- name: GetRefreshToken :one
SELECT * from refresh_tokens
where token = $1;