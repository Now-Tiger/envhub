-- name: GetEnvironmentByID :one
SELECT * FROM environments
WHERE id = $1
LIMIT 1;

-- name: GetEnvironmentByName :one
SELECT * FROM environments
WHERE project_id = $1 AND name = $2
LIMIT 1;

-- name: CreateEnvironment :one
INSERT INTO environments (
    project_id,
    name,
    description,
    is_protected,
    color
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: UpdateEnvironment :one
UPDATE environments
SET 
    description = COALESCE($2, description),
    is_protected = COALESCE($3, is_protected),
    color = COALESCE($4, color),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListEnvironmentsByProject :many
SELECT * FROM environments
WHERE project_id = $1
ORDER BY created_at ASC;

-- name: DeleteEnvironment :exec
DELETE FROM environments
WHERE id = $1;
