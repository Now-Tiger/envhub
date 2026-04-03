package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Now-Tiger/envhub/internal/repository"
	"github.com/Now-Tiger/envhub/internal/service"
	"github.com/Now-Tiger/envhub/pkg/types"
)

// TeamHandler handles team member HTTP requests
type TeamHandler struct {
	svc  *service.TeamService
	repo repository.Querier
}

// NewTeamHandler creates a new TeamHandler
func NewTeamHandler(svc *service.TeamService, repo repository.Querier) *TeamHandler {
	return &TeamHandler{svc: svc, repo: repo}
}

// AddMember handles POST /projects/:id/members
func (h *TeamHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	projectUUID, err := parseUUID(projectID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "")
		return
	}

	var req types.AddMemberRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	// Get organization ID from project
	project, err := h.repo.GetProjectByID(r.Context(), projectUUID)
	if err != nil {
		respondError(w, http.StatusNotFound, "project not found", "")
		return
	}
	orgID := project.OrganizationID

	resp, err := h.svc.AddMember(r.Context(), orgID, projectUUID, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to add member", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

// ListMembers handles GET /projects/:id/members
func (h *TeamHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	projectUUID, err := parseUUID(projectID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "")
		return
	}

	resp, err := h.svc.ListMembers(r.Context(), projectUUID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list members", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// RemoveMember handles DELETE /projects/:id/members/:userId
func (h *TeamHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	projectUUID, err := parseUUID(projectID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "")
		return
	}

	userIDStr := chi.URLParam(r, "userId")
	userID, err := parseUUID(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID", "")
		return
	}

	// Get organization ID from project
	project, err := h.repo.GetProjectByID(r.Context(), projectUUID)
	if err != nil {
		respondError(w, http.StatusNotFound, "project not found", "")
		return
	}
	orgID := project.OrganizationID

	if err := h.svc.RemoveMember(r.Context(), orgID, projectUUID, userID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to remove member", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateMemberRole handles PATCH /projects/:id/members/:userId
func (h *TeamHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	projectUUID, err := parseUUID(projectID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "")
		return
	}

	userIDStr := chi.URLParam(r, "userId")
	userID, err := parseUUID(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID", "")
		return
	}

	var req types.UpdateMemberRoleRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// Get organization ID from project
	project, err := h.repo.GetProjectByID(r.Context(), projectUUID)
	if err != nil {
		respondError(w, http.StatusNotFound, "project not found", "")
		return
	}
	orgID := project.OrganizationID

	resp, err := h.svc.UpdateMemberRole(r.Context(), orgID, projectUUID, userID, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update member role", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resp)
}
