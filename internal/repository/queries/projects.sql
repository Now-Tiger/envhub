-- name: GetProjectByID :one
SELECT * FROM projects
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: CreateProject :one
INSERT INTO projects (
    organization_id,
    name,
    description,
    encrypted_dek,
    dek_version,
    color,
    icon
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: UpdateProject :one
UPDATE projects
SET 
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    color = COALESCE($4, color),
    icon = COALESCE($5, icon),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ListProjectsByOrganization :many
SELECT * FROM projects
WHERE organization_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: RotateProjectDEK :one
UPDATE projects
SET 
    encrypted_dek = $2,
    dek_version = dek_version + 1,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteProject :exec
UPDATE projects
SET deleted_at = NOW()
WHERE id = $1;

-- name: CountProjectsByOrganization :one
SELECT COUNT(*) FROM projects
WHERE organization_id = $1 AND deleted_at IS NULL;
