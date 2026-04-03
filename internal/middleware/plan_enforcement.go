package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Now-Tiger/envhub/internal/service"
	"github.com/Now-Tiger/envhub/internal/utils"
)

// Feature constants for plan features
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

// RequireFeature returns a middleware that requires a specific feature
func RequireFeature(feature string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get organization ID from URL or context
			orgIDStr := chi.URLParam(r, "orgId")
			if orgIDStr == "" {
				// Try to get from query params
				orgIDStr = r.URL.Query().Get("org_id")
			}

			if orgIDStr == "" {
				http.Error(w, "organization ID required", http.StatusBadRequest)
				return
			}

			orgID, err := uuid.Parse(orgIDStr)
			if err != nil {
				http.Error(w, "invalid organization ID", http.StatusBadRequest)
				return
			}

			// Get plan service from context
			planSvc, ok := r.Context().Value("plan_service").(*service.PlanService)
			if !ok {
				http.Error(w, "plan service not available", http.StatusInternalServerError)
				return
			}

			hasFeature, err := planSvc.HasFeature(r.Context(), orgID, feature)
			if err != nil {
				http.Error(w, "failed to check feature", http.StatusInternalServerError)
				return
			}

			if !hasFeature {
				respondFeatureNotAllowed(w, feature)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// EnforceProjectLimit returns a middleware that enforces project limits
func EnforceProjectLimit() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get organization ID
			var orgID uuid.UUID
			var err error

			// Try URL param first
			orgIDStr := chi.URLParam(r, "orgId")
			if orgIDStr == "" {
				orgIDStr = r.URL.Query().Get("org_id")
			}

			if orgIDStr != "" {
				orgID, err = uuid.Parse(orgIDStr)
				if err != nil {
					http.Error(w, "invalid organization ID", http.StatusBadRequest)
					return
				}
			} else {
				// If no org ID in URL, check if we're creating a project and get org from body
				if r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects" {
					// We'll handle this in the handler instead
					next.ServeHTTP(w, r)
					return
				}
				http.Error(w, "organization ID required", http.StatusBadRequest)
				return
			}

			// Get plan service from context
			planSvc, ok := r.Context().Value("plan_service").(*service.PlanService)
			if !ok {
				http.Error(w, "plan service not available", http.StatusInternalServerError)
				return
			}

			canCreate, plan, err := planSvc.CanCreateProject(r.Context(), orgID)
			if err != nil {
				http.Error(w, "failed to check project limit", http.StatusInternalServerError)
				return
			}

			if !canCreate {
				respondPlanLimitExceeded(w, "projects", plan.MaxProjects)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// EnforceEnvironmentLimit returns a middleware that enforces environment limits per project
func EnforceEnvironmentLimit() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This middleware expects project ID in URL
			projectIDStr := chi.URLParam(r, "id")
			if projectIDStr == "" {
				next.ServeHTTP(w, r)
				return
			}

			projectID, err := uuid.Parse(projectIDStr)
			if err != nil {
				http.Error(w, "invalid project ID", http.StatusBadRequest)
				return
			}

			// Get organization ID from query or context
			orgIDStr := r.URL.Query().Get("org_id")
			if orgIDStr == "" {
				http.Error(w, "organization ID required", http.StatusBadRequest)
				return
			}

			orgID, err := uuid.Parse(orgIDStr)
			if err != nil {
				http.Error(w, "invalid organization ID", http.StatusBadRequest)
				return
			}

			// Get plan service from context
			planSvc, ok := r.Context().Value("plan_service").(*service.PlanService)
			if !ok {
				http.Error(w, "plan service not available", http.StatusInternalServerError)
				return
			}

			canCreate, plan, err := planSvc.CanCreateEnvironment(r.Context(), orgID, projectID)
			if err != nil {
				http.Error(w, "failed to check environment limit", http.StatusInternalServerError)
				return
			}

			if !canCreate {
				respondPlanLimitExceeded(w, "environments per project", plan.MaxEnvironmentsPerProject)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// EnforceSecretLimit returns a middleware that enforces secret limits per environment
func EnforceSecretLimit() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be handled in the secret handler directly
			// as we need environment ID from the request path
			next.ServeHTTP(w, r)
		})
	}
}

// respondPlanLimitExceeded returns a 403 response with limit info
func respondPlanLimitExceeded(w http.ResponseWriter, resource string, limit int32) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(utils.ErrorResponse{
		Success:    false,
		StatusCode: 403,
		Message:    fmt.Sprintf("plan limit exceeded: %s limit is %d", resource, limit),
	})
}

// respondFeatureNotAllowed returns a 403 response for missing features
func respondFeatureNotAllowed(w http.ResponseWriter, feature string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(utils.ErrorResponse{
		Success:    false,
		StatusCode: 403,
		Message:    fmt.Sprintf("feature not allowed: plan does not include feature '%s'", feature),
	})
}
