-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByAuthProviderID :one
SELECT * FROM users
WHERE auth_provider_id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
    email,
    full_name,
    avatar_url,
    auth_provider_id,
    is_active,
    email_verified
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET 
    full_name = COALESCE($2, full_name),
    avatar_url = COALESCE($3, avatar_url),
    is_active = COALESCE($4, is_active),
    email_verified = COALESCE($5, email_verified),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW()
WHERE id = $1;

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = NOW()
WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
