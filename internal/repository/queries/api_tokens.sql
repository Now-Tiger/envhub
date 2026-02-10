-- name: GetAPITokenByHash :one
SELECT * FROM api_tokens
WHERE token_hash = $1 AND revoked_at IS NULL
AND (expires_at IS NULL OR expires_at > NOW())
LIMIT 1;

-- name: CreateAPIToken :one
INSERT INTO api_tokens (
    user_id,
    name,
    token_hash,
    scopes,
    organization_id,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: UpdateTokenUsage :exec
UPDATE api_tokens
SET 
    last_used_at = NOW(),
    usage_count = usage_count + 1
WHERE id = $1;

-- name: RevokeAPIToken :exec
UPDATE api_tokens
SET revoked_at = NOW()
WHERE id = $1;

-- name: ListUserAPITokens :many
SELECT * FROM api_tokens
WHERE user_id = $1 AND revoked_at IS NULL
ORDER BY created_at DESC;
