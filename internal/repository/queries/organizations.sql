-- name: GetOrganizationByID :one
SELECT * FROM organizations
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetOrganizationBySlug :one
SELECT * FROM organizations
WHERE slug = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: CreateOrganization :one
INSERT INTO organizations (
    name,
    slug,
    plan_type,
    max_projects,
    max_secrets_per_project,
    owner_id
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: UpdateOrganization :one
UPDATE organizations
SET 
    name = COALESCE($2, name),
    plan_type = COALESCE($3, plan_type),
    max_projects = COALESCE($4, max_projects),
    max_secrets_per_project = COALESCE($5, max_secrets_per_project),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ListUserOrganizations :many
SELECT o.* FROM organizations o
JOIN organization_members om ON om.organization_id = o.id
WHERE om.user_id = $1 AND o.deleted_at IS NULL
ORDER BY o.created_at DESC;

-- name: SoftDeleteOrganization :exec
UPDATE organizations
SET deleted_at = NOW()
WHERE id = $1;
