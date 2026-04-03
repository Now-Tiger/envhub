-- name: GetPlanByID :one
SELECT * FROM plans
WHERE id = $1 AND is_active = true
LIMIT 1;

-- name: GetPlanByName :one
SELECT * FROM plans
WHERE name = $1 AND is_active = true
LIMIT 1;

-- name: ListPlans :many
SELECT * FROM plans
WHERE is_active = true
ORDER BY price_monthly ASC;

-- name: ListAllPlans :many
SELECT * FROM plans
ORDER BY price_monthly ASC;

-- name: CreatePlan :one
INSERT INTO plans (
    name,
    display_name,
    description,
    price_monthly,
    max_projects,
    max_environments_per_project,
    max_secrets_per_project,
    max_team_members,
    features
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdatePlan :one
UPDATE plans SET
    display_name = COALESCE($2, display_name),
    description = COALESCE($3, description),
    price_monthly = COALESCE($4, price_monthly),
    max_projects = COALESCE($5, max_projects),
    max_environments_per_project = COALESCE($6, max_environments_per_project),
    max_secrets_per_project = COALESCE($7, max_secrets_per_project),
    max_team_members = COALESCE($8, max_team_members),
    features = COALESCE($9, features),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeactivatePlan :exec
UPDATE plans SET
    is_active = false,
    updated_at = NOW()
WHERE id = $1;
