-- name: GetSubscriptionByID :one
SELECT * FROM subscriptions
WHERE id = $1
LIMIT 1;

-- name: GetSubscriptionByOrganization :one
SELECT * FROM subscriptions
WHERE organization_id = $1 AND status = 'active'
LIMIT 1;

-- name: GetActiveSubscriptionByOrganization :one
SELECT s.* FROM subscriptions s
JOIN organizations o ON o.id = s.organization_id
WHERE o.id = $1 AND s.status = 'active'
LIMIT 1;

-- name: GetSubscriptionByStripeCustomer :one
SELECT * FROM subscriptions
WHERE stripe_customer_id = $1
LIMIT 1;

-- name: GetSubscriptionByStripeSubscription :one
SELECT * FROM subscriptions
WHERE stripe_subscription_id = $1
LIMIT 1;

-- name: CreateSubscription :one
INSERT INTO subscriptions (
    organization_id,
    plan_id,
    stripe_customer_id,
    stripe_subscription_id,
    status,
    billing_cycle_start,
    billing_cycle_end
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateSubscription :one
UPDATE subscriptions SET
    plan_id = COALESCE($2, plan_id),
    stripe_customer_id = COALESCE($3, stripe_customer_id),
    stripe_subscription_id = COALESCE($4, stripe_subscription_id),
    status = COALESCE($5, status),
    billing_cycle_end = COALESCE($6, billing_cycle_end),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CancelSubscription :exec
UPDATE subscriptions SET
    status = 'canceled',
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateSubscriptionStatus :exec
UPDATE subscriptions SET
    status = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateSubscriptionPlan :exec
UPDATE subscriptions SET
    plan_id = $2,
    billing_cycle_start = NOW(),
    billing_cycle_end = NOW() + INTERVAL '1 month',
    updated_at = NOW()
WHERE id = $1;

-- name: ListSubscriptionsByOrganization :many
SELECT * FROM subscriptions
WHERE organization_id = $1
ORDER BY created_at DESC;

-- name: CountActiveSubscriptions :one
SELECT COUNT(*) FROM subscriptions
WHERE status = 'active';
