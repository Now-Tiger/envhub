-- name: GetSecretByID :one
SELECT * FROM secrets
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetSecretByKey :one
SELECT * FROM secrets
WHERE environment_id = $1 AND key = $2 AND deleted_at IS NULL AND is_active = true
LIMIT 1;

-- name: CreateSecret :one
INSERT INTO secrets (
    environment_id,
    key,
    encrypted_value,
    description,
    is_active,
    version,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: UpdateSecret :one
UPDATE secrets
SET 
    encrypted_value = $2,
    description = COALESCE($3, description),
    version = version + 1,
    updated_by = $4,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ListSecretsByEnvironment :many
SELECT * FROM secrets
WHERE environment_id = $1 AND deleted_at IS NULL AND is_active = true
ORDER BY key ASC;

-- name: SoftDeleteSecret :exec
UPDATE secrets
SET 
    deleted_at = NOW(),
    is_active = false,
    updated_by = $2
WHERE id = $1;

-- name: CountSecretsByEnvironment :one
SELECT COUNT(*) FROM secrets
WHERE environment_id = $1 AND deleted_at IS NULL AND is_active = true;

-- name: DeactivateSecret :exec
UPDATE secrets
SET is_active = false, updated_by = $2
WHERE id = $1;
