package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Now-Tiger/envhub/internal/middleware"
	"github.com/Now-Tiger/envhub/internal/service"
	"github.com/Now-Tiger/envhub/pkg/types"
)

// ProjectHandler handles project HTTP requests
type ProjectHandler struct {
	svc             *service.ProjectService
	planSvc         *service.PlanService
	organizationSvc *service.OrganizationService
}

// NewProjectHandler creates a new ProjectHandler
func NewProjectHandler(svc *service.ProjectService, planSvc *service.PlanService, organizationSvc *service.OrganizationService) *ProjectHandler {
	return &ProjectHandler{svc: svc, planSvc: planSvc, organizationSvc: organizationSvc}
}

// Create handles POST /projects
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req types.CreateProjectRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// Validate request (only name is required)
	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "")
		return
	}

	var orgID uuid.UUID

	// Hybrid approach: org_id is optional
	if req.OrgID != "" {
		// If org_id provided, parse and validate user has access
		var err error
		orgID, err = uuid.Parse(req.OrgID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid organization ID", "")
			return
		}

		// Verify user has access to this organization
		orgs, err := h.organizationSvc.ListUserOrganizations(r.Context(), userID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to verify organization access", err.Error())
			return
		}

		hasAccess := false
		for _, org := range orgs {
			if org.ID == orgID {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			respondError(w, http.StatusForbidden, "access denied to organization", "")
			return
		}
	} else {
		// If org_id NOT provided, auto-assign to user's first/default organization
		// First try organizations where user is a member
		orgs, err := h.organizationSvc.ListUserOrganizations(r.Context(), userID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list organizations", err.Error())
			return
		}

		// If no member organizations, check organizations owned by user
		if len(orgs) == 0 {
			orgs, err = h.organizationSvc.ListOrganizationsByOwner(r.Context(), userID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "failed to list organizations", err.Error())
				return
			}
		}

		if len(orgs) == 0 {
			respondError(w, http.StatusBadRequest, "no organization available", "You must be a member of an organization to create a project")
			return
		}

		// Use the first organization (default)
		orgID = orgs[0].ID
	}

	// Check project limit before creating
	canCreate, _, err := h.planSvc.CanCreateProject(r.Context(), orgID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to check project limit", err.Error())
		return
	}

	if !canCreate {
		respondError(w, http.StatusForbidden, "plan limit exceeded",
			"You have reached the maximum number of projects for your plan. Upgrade to create more projects.")
		return
	}

	resp, err := h.svc.Create(r.Context(), userID, orgID, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create project", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

// List handles GET /projects
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	params := types.PaginationParams{
		Page:     getIntQuery(r, "page", 1),
		PageSize: getIntQuery(r, "page_size", 20),
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "")
		return
	}

	orgIDStr := r.URL.Query().Get("org_id")
	var orgID uuid.UUID

	if orgIDStr != "" {
		// If org_id provided, parse and validate access
		var err error
		orgID, err = uuid.Parse(orgIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid organization ID", "")
			return
		}
		// Verify user has access
		orgs, err := h.organizationSvc.ListUserOrganizations(r.Context(), userID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to verify organization access", err.Error())
			return
		}
		hasAccess := false
		for _, o := range orgs {
			if o.ID == orgID {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			// Also check owned orgs
			ownedOrgs, _ := h.organizationSvc.ListOrganizationsByOwner(r.Context(), userID)
			for _, o := range ownedOrgs {
				if o.ID == orgID {
					hasAccess = true
					break
				}
			}
		}
		if !hasAccess {
			respondError(w, http.StatusForbidden, "access denied to organization", "")
			return
		}
	} else {
		// If org_id NOT provided, auto-assign to user's first/default organization
		orgs, err := h.organizationSvc.ListUserOrganizations(r.Context(), userID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list organizations", err.Error())
			return
		}
		// If no member organizations, check owned orgs
		if len(orgs) == 0 {
			orgs, err = h.organizationSvc.ListOrganizationsByOwner(r.Context(), userID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "failed to list organizations", err.Error())
				return
			}
		}
		if len(orgs) == 0 {
			respondError(w, http.StatusBadRequest, "no organization available", "You must be a member of an organization to list projects")
			return
		}
		orgID = orgs[0].ID
	}

	resp, err := h.svc.List(r.Context(), orgID, params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list projects", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Get handles GET /projects/:id
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "")
		return
	}

	resp, err := h.svc.GetByID(r.Context(), userID, projectID)
	if err != nil {
		if err == service.ErrProjectNotFound {
			respondError(w, http.StatusNotFound, "project not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get project", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Update handles PATCH /projects/:id
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "")
		return
	}

	var req types.UpdateProjectRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "")
		return
	}

	resp, err := h.svc.Update(r.Context(), userID, projectID, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update project", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /projects/:id
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, projectID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete project", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RotateDEK handles POST /projects/:id/rotate-dek
func (h *ProjectHandler) RotateDEK(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "")
		return
	}

	if err := h.svc.RotateDEK(r.Context(), userID, projectID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to rotate DEK", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "DEK rotated successfully"})
}
