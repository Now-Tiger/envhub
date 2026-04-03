package handlers

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Now-Tiger/envhub/internal/middleware"
	"github.com/Now-Tiger/envhub/internal/service"
	"github.com/Now-Tiger/envhub/pkg/types"
)

// SecretHandler handles secret HTTP requests
type SecretHandler struct {
	svc *service.SecretService
}

// NewSecretHandler creates a new SecretHandler
func NewSecretHandler(svc *service.SecretService) *SecretHandler {
	return &SecretHandler{svc: svc}
}

// Create handles POST /projects/:id/secrets
func (h *SecretHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req types.CreateSecretRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	// Get environment ID from query params
	envIDStr := r.URL.Query().Get("environment_id")
	envID, err := uuid.Parse(envIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "environment_id required", "")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "")
		return
	}

	resp, err := h.svc.Create(r.Context(), userID, envID, req)
	if err != nil {
		if err == service.ErrEnvironmentNotFound {
			respondError(w, http.StatusNotFound, "environment not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create secret", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

// List handles GET /projects/:id/secrets
func (h *SecretHandler) List(w http.ResponseWriter, r *http.Request) {
	envIDStr := r.URL.Query().Get("environment_id")
	envID, err := uuid.Parse(envIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "environment_id required", "")
		return
	}

	params := types.PaginationParams{
		Page:     getIntQuery(r, "page", 1),
		PageSize: getIntQuery(r, "page_size", 20),
	}

	resp, err := h.svc.List(r.Context(), envID, params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list secrets", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// GetByEnvironment handles GET /projects/:id/secrets/:env
func (h *SecretHandler) GetByEnvironment(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "")
		return
	}

	envName := chi.URLParam(r, "envName")
	log.Printf("DEBUG: GetByEnvironment called with envName=%s\n", envName)
	if envName == "" {
		respondError(w, http.StatusBadRequest, "environment name required", "")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "")
		return
	}

	resp, err := h.svc.GetByEnvironment(r.Context(), userID, projectID, envName)
	if err != nil {
		if err == service.ErrEnvironmentNotFound {
			respondError(w, http.StatusNotFound, "environment not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get secrets", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// GetByID handles getting a single secret by ID
func (h *SecretHandler) GetByID(w http.ResponseWriter, r *http.Request, projectID, secretID string) {
	parsedSecretID, err := uuid.Parse(secretID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid secret ID", "")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "")
		return
	}

	resp, err := h.svc.GetByID(r.Context(), userID, parsedSecretID)
	if err != nil {
		if err == service.ErrSecretNotFound {
			respondError(w, http.StatusNotFound, "secret not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get secret", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Get handles GET /projects/:id/secrets/:id
func (h *SecretHandler) Get(w http.ResponseWriter, r *http.Request) {
	secretID, err := uuid.Parse(chi.URLParam(r, "secretId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid secret ID", "")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "")
		return
	}

	resp, err := h.svc.GetByID(r.Context(), userID, secretID)
	if err != nil {
		if err == service.ErrSecretNotFound {
			respondError(w, http.StatusNotFound, "secret not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get secret", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Update handles PATCH /projects/:id/secrets/:id
func (h *SecretHandler) Update(w http.ResponseWriter, r *http.Request) {
	secretID, err := uuid.Parse(chi.URLParam(r, "secretId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid secret ID", "")
		return
	}

	var req types.UpdateSecretRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "")
		return
	}

	resp, err := h.svc.Update(r.Context(), userID, secretID, req)
	if err != nil {
		if err == service.ErrSecretNotFound {
			respondError(w, http.StatusNotFound, "secret not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update secret", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /projects/:id/secrets/:id
func (h *SecretHandler) Delete(w http.ResponseWriter, r *http.Request) {
	secretID, err := uuid.Parse(chi.URLParam(r, "secretId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid secret ID", "")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, secretID); err != nil {
		if err == service.ErrSecretNotFound {
			respondError(w, http.StatusNotFound, "secret not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete secret", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
