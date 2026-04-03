package types

import (
	"time"

	"github.com/google/uuid"
)

// PlanResponse represents a subscription plan
type PlanResponse struct {
	ID                        uuid.UUID `json:"id"`
	Name                      string    `json:"name"`
	DisplayName               string    `json:"display_name"`
	Description               string    `json:"description"`
	PriceMonthly              float64   `json:"price_monthly"`
	MaxProjects               int32     `json:"max_projects"`
	MaxEnvironmentsPerProject int32     `json:"max_environments_per_project"`
	MaxSecretsPerProject      int32     `json:"max_secrets_per_project"`
	MaxTeamMembers            int32     `json:"max_team_members"`
	Features                  []string  `json:"features"`
	IsActive                  bool      `json:"is_active"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

// PlanFeature constants for feature flags
const (
	FeatureBasicEncryption         = "basic_encryption"
	FeatureAdvancedEncryption      = "advanced_encryption"
	FeatureEmailSupport            = "email_support"
	FeaturePrioritySupport         = "priority_support"
	Feature24SevenSupport          = "24_7_support"
	FeatureDedicatedSupport        = "dedicated_support"
	FeatureUnlimitedProjects       = "unlimited_projects"
	FeatureUnlimitedEnvironments   = "unlimited_environments"
	FeatureAuditLogs               = "audit_logs"
	FeatureAPIAccess               = "api_access"
	FeatureRBAC                    = "rbac"
	FeatureSSO                     = "sso"
	FeatureCustomIntegrations      = "custom_integrations"
	FeatureDedicatedAccountManager = "dedicated_account_manager"
)

// SubscriptionStatus represents the subscription status
type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusPastDue  SubscriptionStatus = "past_due"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
	SubscriptionStatusTrialing SubscriptionStatus = "trialing"
)

// SubscriptionResponse represents an organization's subscription
type SubscriptionResponse struct {
	ID                   uuid.UUID          `json:"id"`
	OrganizationID       uuid.UUID          `json:"organization_id"`
	PlanID               uuid.UUID          `json:"plan_id"`
	Plan                 *PlanResponse      `json:"plan,omitempty"`
	StripeCustomerID     *string            `json:"stripe_customer_id"`
	StripeSubscriptionID *string            `json:"stripe_subscription_id"`
	Status               SubscriptionStatus `json:"status"`
	BillingCycleStart    time.Time          `json:"billing_cycle_start"`
	BillingCycleEnd      *time.Time         `json:"billing_cycle_end"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}

// CreateSubscriptionRequest represents a request to create/update subscription
type CreateSubscriptionRequest struct {
	OrganizationID string `json:"organization_id" validate:"required,uuid"`
	PlanName       string `json:"plan_name" validate:"required"`
}

// ListPlansResponse represents response for listing plans
type ListPlansResponse struct {
	Plans []PlanResponse `json:"plans"`
}

// OrganizationWithPlan represents organization with its plan details
type OrganizationWithPlan struct {
	ID             uuid.UUID             `json:"id"`
	Name           string                `json:"name"`
	Slug           string                `json:"slug"`
	PlanID         *uuid.UUID            `json:"plan_id"`
	SubscriptionID *uuid.UUID            `json:"subscription_id"`
	OwnerID        uuid.UUID             `json:"owner_id"`
	CurrentPlan    *PlanResponse         `json:"current_plan,omitempty"`
	Subscription   *SubscriptionResponse `json:"subscription,omitempty"`
}
