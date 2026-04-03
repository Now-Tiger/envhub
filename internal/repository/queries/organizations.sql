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

-- name: ListOrganizationsByOwner :many
SELECT * FROM organizations
WHERE owner_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreateOrganizationMember :one
INSERT INTO organization_members (
    organization_id,
    user_id,
    role,
    joined_at
) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetOrganizationMember :one
SELECT * FROM organization_members
WHERE organization_id = $1 AND user_id = $2
LIMIT 1;

-- name: UpdateOrganizationMemberRole :one
UPDATE organization_members
SET role = $3, updated_at = NOW()
WHERE organization_id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteOrganizationMember :exec
DELETE FROM organization_members
WHERE organization_id = $1 AND user_id = $2;

-- name: ListOrganizationMembers :many
SELECT om.*, u.email, u.full_name, u.avatar_url 
FROM organization_members om
JOIN users u ON om.user_id = u.id
WHERE om.organization_id = $1
ORDER BY om.joined_at DESC;

-- name: SoftDeleteOrganization :exec
UPDATE organizations
SET deleted_at = NOW()
WHERE id = $1;
