-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, expires_at, revoked_at, user_id)
VALUES (
    $1, NOW() AT TIME ZONE 'UTC', NOW() AT TIME ZONE 'UTC', $2, $3, $4
)
RETURNING *;


-- name: GetRefreshToken :one
SELECT *
FROM refresh_tokens
WHERE token = $1;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET updated_at = NOW() AT TIME ZONE 'UTC',
  revoked_at = NOW() AT TIME ZONE 'UTC'
WHERE token = $1
RETURNING *;

