package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Now-Tiger/envhub/internal/repository"
	"github.com/Now-Tiger/envhub/pkg/types"
)

// ErrPlanNotFound is returned when plan is not found
var ErrPlanNotFound = errors.New("plan not found")

// ErrSubscriptionNotFound is returned when subscription is not found
var ErrSubscriptionNotFound = errors.New("subscription not found")

// ErrPlanLimitExceeded is returned when plan limit is exceeded
var ErrPlanLimitExceeded = errors.New("plan limit exceeded")

// ErrFeatureNotAllowed is returned when feature is not allowed
var ErrFeatureNotAllowed = errors.New("feature not allowed")

// PlanService handles plan-related business logic
type PlanService struct {
	querier *repository.Queries
	pool    *pgxpool.Pool
}

// planInfo holds static plan configuration
type planInfo struct {
	name                      string
	displayName               string
	description               string
	priceMonthly              float64
	maxProjects               int32
	maxEnvironmentsPerProject int32
	maxSecretsPerProject      int32
	maxTeamMembers            int32
	features                  []string
}

// plansConfig is the static plan configuration
var plansConfig = map[string]planInfo{
	"free": {
		name:                      "free",
		displayName:               "Free",
		description:               "Perfect for individuals and small projects",
		priceMonthly:              0,
		maxProjects:               3,
		maxEnvironmentsPerProject: 5,
		maxSecretsPerProject:      50,
		maxTeamMembers:            2,
		features:                  []string{"basic_encryption", "email_support", "3_projects", "5_environments_per_project"},
	},
	"team": {
		name:                      "team",
		displayName:               "Team",
		description:               "For growing teams needing more resources",
		priceMonthly:              29,
		maxProjects:               10,
		maxEnvironmentsPerProject: 10,
		maxSecretsPerProject:      200,
		maxTeamMembers:            10,
		features:                  []string{"basic_encryption", "email_support", "priority_support", "10_projects", "10_environments_per_project", "audit_logs", "api_access"},
	},
	"enterprise": {
		name:                      "enterprise",
		displayName:               "Enterprise",
		description:               "For large organizations with advanced needs",
		priceMonthly:              99,
		maxProjects:               50,
		maxEnvironmentsPerProject: 20,
		maxSecretsPerProject:      1000,
		maxTeamMembers:            50,
		features:                  []string{"advanced_encryption", "24_7_support", "dedicated_support", "unlimited_projects", "unlimited_environments", "audit_logs", "api_access", "rbac", "sso", "custom_integrations", "priority_support", "dedicated_account_manager"},
	},
}

// NewPlanService creates a new PlanService
func NewPlanService(querier *repository.Queries, pool *pgxpool.Pool) *PlanService {
	return &PlanService{
		querier: querier,
		pool:    pool,
	}
}

// GetOrgPlan returns the organization's current plan
func (s *PlanService) GetOrgPlan(ctx context.Context, orgID uuid.UUID) (*types.PlanResponse, error) {
	org, err := s.querier.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Get plan name from plan_type (default to free)
	planName := "free"
	if org.PlanType != nil && *org.PlanType != "" {
		switch *org.PlanType {
		case "enterprise":
			planName = "enterprise"
		case "pro", "team":
			planName = "team"
		}
	}

	return s.GetPlanByName(ctx, planName)
}

// GetPlanByName returns a plan by its name
func (s *PlanService) GetPlanByName(ctx context.Context, name string) (*types.PlanResponse, error) {
	plan, ok := plansConfig[name]
	if !ok {
		return nil, ErrPlanNotFound
	}

	return &types.PlanResponse{
		ID:                        uuid.Nil, // Static plans don't have UUIDs
		Name:                      plan.name,
		DisplayName:               plan.displayName,
		Description:               plan.description,
		PriceMonthly:              plan.priceMonthly,
		MaxProjects:               plan.maxProjects,
		MaxEnvironmentsPerProject: plan.maxEnvironmentsPerProject,
		MaxSecretsPerProject:      plan.maxSecretsPerProject,
		MaxTeamMembers:            plan.maxTeamMembers,
		Features:                  plan.features,
		IsActive:                  true,
	}, nil
}

// ListPlans returns all available plans
func (s *PlanService) ListPlans(ctx context.Context) ([]types.PlanResponse, error) {
	result := make([]types.PlanResponse, 0, len(plansConfig))
	for _, plan := range plansConfig {
		result = append(result, types.PlanResponse{
			ID:                        uuid.Nil,
			Name:                      plan.name,
			DisplayName:               plan.displayName,
			Description:               plan.description,
			PriceMonthly:              plan.priceMonthly,
			MaxProjects:               plan.maxProjects,
			MaxEnvironmentsPerProject: plan.maxEnvironmentsPerProject,
			MaxSecretsPerProject:      plan.maxSecretsPerProject,
			MaxTeamMembers:            plan.maxTeamMembers,
			Features:                  plan.features,
			IsActive:                  true,
		})
	}
	return result, nil
}

// CanCreateProject checks if the organization can create another project
func (s *PlanService) CanCreateProject(ctx context.Context, orgID uuid.UUID) (bool, *types.PlanResponse, error) {
	plan, err := s.GetOrgPlan(ctx, orgID)
	if err != nil {
		return false, nil, err
	}

	// Count current projects
	count, err := s.querier.CountProjectsByOrganization(ctx, orgID)
	if err != nil {
		return false, nil, err
	}

	canCreate := count < int64(plan.MaxProjects)
	return canCreate, plan, nil
}

// CanCreateEnvironment checks if the project can have another environment
func (s *PlanService) CanCreateEnvironment(ctx context.Context, orgID, projectID uuid.UUID) (bool, *types.PlanResponse, error) {
	plan, err := s.GetOrgPlan(ctx, orgID)
	if err != nil {
		return false, nil, err
	}

	// Check for unlimited environments feature
	for _, f := range plan.Features {
		if f == types.FeatureUnlimitedEnvironments {
			return true, plan, nil
		}
	}

	// Count current environments
	envs, err := s.querier.ListEnvironmentsByProject(ctx, projectID)
	if err != nil {
		return false, nil, err
	}

	canCreate := int32(len(envs)) < plan.MaxEnvironmentsPerProject
	return canCreate, plan, nil
}

// CanCreateSecret checks if the environment can have another secret
func (s *PlanService) CanCreateSecret(ctx context.Context, orgID, environmentID uuid.UUID) (bool, *types.PlanResponse, error) {
	plan, err := s.GetOrgPlan(ctx, orgID)
	if err != nil {
		return false, nil, err
	}

	// Count current secrets
	count, err := s.querier.CountSecretsByEnvironment(ctx, environmentID)
	if err != nil {
		return false, nil, err
	}

	canCreate := count < int64(plan.MaxSecretsPerProject)
	return canCreate, plan, nil
}

// HasFeature checks if the organization's plan has a specific feature
func (s *PlanService) HasFeature(ctx context.Context, orgID uuid.UUID, feature string) (bool, error) {
	plan, err := s.GetOrgPlan(ctx, orgID)
	if err != nil {
		return false, err
	}

	for _, f := range plan.Features {
		if f == feature {
			return true, nil
		}
	}

	return false, nil
}

// UpdateOrgPlan updates an organization's plan (for admin/testing)
func (s *PlanService) UpdateOrgPlan(ctx context.Context, orgID uuid.UUID, planName string) (*types.PlanResponse, error) {
	// Validate plan exists
	plan, ok := plansConfig[planName]
	if !ok {
		return nil, ErrPlanNotFound
	}

	// Update organization plan_type
	planType := planName
	_, err := s.querier.UpdateOrganization(ctx, repository.UpdateOrganizationParams{
		ID:       orgID,
		PlanType: &planType,
	})
	if err != nil {
		return nil, err
	}

	return &types.PlanResponse{
		ID:                        uuid.Nil,
		Name:                      plan.name,
		DisplayName:               plan.displayName,
		Description:               plan.description,
		PriceMonthly:              plan.priceMonthly,
		MaxProjects:               plan.maxProjects,
		MaxEnvironmentsPerProject: plan.maxEnvironmentsPerProject,
		MaxSecretsPerProject:      plan.maxSecretsPerProject,
		MaxTeamMembers:            plan.maxTeamMembers,
		Features:                  plan.features,
		IsActive:                  true,
	}, nil
}
