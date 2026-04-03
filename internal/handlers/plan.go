package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Now-Tiger/envhub/internal/middleware"
	"github.com/Now-Tiger/envhub/internal/service"
)

// PlanHandler handles plan-related HTTP requests
type PlanHandler struct {
	svc *service.PlanService
}

// NewPlanHandler creates a new PlanHandler
func NewPlanHandler(svc *service.PlanService) *PlanHandler {
	return &PlanHandler{svc: svc}
}

// ListPlans handles GET /api/v1/plans
func (h *PlanHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.svc.ListPlans(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list plans", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"plans": plans,
	})
}

// HandleStripeWebhook handles POST /api/v1/webhooks/stripe
// This is a mock handler for local testing without real Stripe
func (h *PlanHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EventType string `json:"event_type"`
		OrgID     string `json:"org_id"`
		PlanName  string `json:"plan_name"`
	}

	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// Validate required fields
	if req.OrgID == "" || req.PlanName == "" {
		respondError(w, http.StatusBadRequest, "org_id and plan_name are required", "")
		return
	}

	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid organization ID", "")
		return
	}

	// Update the organization's plan
	_, err = h.svc.UpdateOrgPlan(r.Context(), orgID, req.PlanName)
	if err != nil {
		if err == service.ErrPlanNotFound {
			respondError(w, http.StatusNotFound, "plan not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update plan", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status":       "success",
		"message":      "subscription updated",
		"organization": req.OrgID,
		"plan":         req.PlanName,
	})
}

// AdminUpdateSubscription handles POST /api/v1/admin/subscription
// Admin endpoint to change organization plan (for local testing)
func (h *PlanHandler) AdminUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrganizationID string `json:"organization_id" validate:"required,uuid"`
		PlanName       string `json:"plan_name" validate:"required"`
	}

	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid organization ID", "")
		return
	}

	// Update the organization's plan
	plan, err := h.svc.UpdateOrgPlan(r.Context(), orgID, req.PlanName)
	if err != nil {
		if err == service.ErrPlanNotFound {
			respondError(w, http.StatusNotFound, "plan not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update plan", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":       "success",
		"message":      "subscription updated",
		"organization": req.OrganizationID,
		"plan":         plan,
	})
}

// GetOrgSubscription handles GET /api/v1/organizations/:id/subscription
func (h *PlanHandler) GetOrgSubscription(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid organization ID", "")
		return
	}

	// Get the organization's current plan
	plan, err := h.svc.GetOrgPlan(r.Context(), orgID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get plan", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"organization_id": orgID,
		"plan":            plan,
	})
}

// RequireFeature returns middleware that checks for a specific feature
func RequireFeature(feature string) func(http.Handler) http.Handler {
	return middleware.RequireFeature(feature)
}

// EnforceProjectLimit returns middleware that checks project limits
func EnforceProjectLimit() func(http.Handler) http.Handler {
	return middleware.EnforceProjectLimit()
}

// EnforceEnvironmentLimit returns middleware that checks environment limits
func EnforceEnvironmentLimit() func(http.Handler) http.Handler {
	return middleware.EnforceEnvironmentLimit()
}
