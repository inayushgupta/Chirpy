-- name: UpdateRefreshToken :one
UPDATE refresh_tokens
SET updated_at = $1, revoked_at = $2
WHERE token = $3
RETURNING *;