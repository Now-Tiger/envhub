package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Now-Tiger/envhub/internal/service"
	"github.com/Now-Tiger/envhub/pkg/types"
)

// EnvironmentHandler handles environment HTTP requests
type EnvironmentHandler struct {
	svc *service.EnvironmentService
}

// NewEnvironmentHandler creates a new EnvironmentHandler
func NewEnvironmentHandler(svc *service.EnvironmentService) *EnvironmentHandler {
	return &EnvironmentHandler{svc: svc}
}

// Create handles POST /projects/:id/environments
func (h *EnvironmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	projectUUID, err := parseUUID(projectID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "")
		return
	}

	var req types.CreateEnvironmentRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	resp, err := h.svc.Create(r.Context(), projectUUID, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create environment", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

// List handles GET /projects/:id/environments
func (h *EnvironmentHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	projectUUID, err := parseUUID(projectID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "")
		return
	}

	resp, err := h.svc.List(r.Context(), projectUUID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list environments", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Get handles GET /projects/:id/environments/:envId
func (h *EnvironmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	envID := chi.URLParam(r, "envId")
	envUUID, err := parseUUID(envID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid environment ID", "")
		return
	}

	resp, err := h.svc.GetByID(r.Context(), envUUID)
	if err != nil {
		if err == service.ErrEnvNotFound {
			respondError(w, http.StatusNotFound, "environment not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get environment", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Update handles PATCH /projects/:id/environments/:envId
func (h *EnvironmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	envID := chi.URLParam(r, "envId")
	envUUID, err := parseUUID(envID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid environment ID", "")
		return
	}

	var req types.UpdateEnvironmentRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	resp, err := h.svc.Update(r.Context(), envUUID, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update environment", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /projects/:id/environments/:envId
func (h *EnvironmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	envID := chi.URLParam(r, "envId")
	envUUID, err := parseUUID(envID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid environment ID", "")
		return
	}

	if err := h.svc.Delete(r.Context(), envUUID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete environment", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
