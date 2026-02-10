-- name: CreateAccessLog :one
INSERT INTO access_logs (
    user_id,
    api_token_id,
    resource_type,
    resource_id,
    action,
    ip_address,
    user_agent,
    success,
    error_message
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: ListAccessLogsByUser :many
SELECT * FROM access_logs
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAccessLogsByResource :many
SELECT * FROM access_logs
WHERE resource_type = $1 AND resource_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListFailedAccessLogs :many
SELECT * FROM access_logs
WHERE success = false
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
