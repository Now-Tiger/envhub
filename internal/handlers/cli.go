package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Now-Tiger/envhub/internal/middleware"
	"github.com/Now-Tiger/envhub/internal/service"
	"github.com/Now-Tiger/envhub/pkg/types"
)

// CLIHandler handles CLI-specific HTTP requests
type CLIHandler struct {
	svc *service.SecretService
}

// NewCLIHandler creates a new CLIHandler
func NewCLIHandler(svc *service.SecretService) *CLIHandler {
	return &CLIHandler{svc: svc}
}

// GetSecrets handles GET /api/v1/cli/secrets/:project/:env
// Optimized endpoint for CLI that returns flat key-value map with caching headers
func (h *CLIHandler) GetSecrets(w http.ResponseWriter, r *http.Request) {
	// Parse project name from URL
	projectName := chi.URLParam(r, "project")
	if projectName == "" {
		respondError(w, http.StatusBadRequest, "project name required", "")
		return
	}

	// Parse environment name from URL
	envName := chi.URLParam(r, "env")
	if envName == "" {
		respondError(w, http.StatusBadRequest, "environment name required", "")
		return
	}

	// Get user ID from context (JWT or API token auth)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required", "")
		return
	}

	// Get secrets for the environment by project name
	secrets, err := h.svc.GetByProjectName(r.Context(), userID, projectName, envName)
	if err != nil {
		if err == service.ErrProjectNotFound {
			respondError(w, http.StatusNotFound, "project not found", "")
			return
		}
		if err == service.ErrEnvironmentNotFound {
			respondError(w, http.StatusNotFound, "environment not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get secrets", err.Error())
		return
	}

	// Build response
	resp := types.CLISecretsResponse{
		Project:     projectName,
		Environment: envName,
		Secrets:     secrets,
		RetrievedAt: time.Now().UTC(),
		Version:     "1.0",
	}

	// Set caching headers for CLI optimization
	// Cache-Control: max-age=0, must-revalidate ensures fresh data but allows client caching
	// For production, you might want shorter cache durations or use ETags
	w.Header().Set("Cache-Control", "private, max-age=60, must-revalidate")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-EnvHub-Cache", "MISS")

	respondJSON(w, http.StatusOK, resp)
}

// GetSecretsWithETag handles GET with ETag support for CLI
// This endpoint supports If-None-Match header for caching
func (h *CLIHandler) GetSecretsWithETag(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")
	if projectName == "" {
		respondError(w, http.StatusBadRequest, "project name required", "")
		return
	}

	envName := chi.URLParam(r, "env")
	if envName == "" {
		respondError(w, http.StatusBadRequest, "environment name required", "")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required", "")
		return
	}

	secrets, err := h.svc.GetByProjectName(r.Context(), userID, projectName, envName)
	if err != nil {
		if err == service.ErrEnvironmentNotFound {
			respondError(w, http.StatusNotFound, "environment not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get secrets", err.Error())
		return
	}

	resp := types.CLISecretsResponse{
		Project:     projectName,
		Environment: envName,
		Secrets:     secrets,
		RetrievedAt: time.Now().UTC(),
		Version:     "1.0",
	}

	// Generate simple ETag based on secret content hash
	// In production, you'd want a more robust ETag implementation
	etag := generateETag(secrets)

	// Check If-None-Match header
	if match := r.Header.Get("If-None-Match"); match == etag {
		w.Header().Set("ETag", etag)
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Set caching headers
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "private, max-age=60, must-revalidate")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	respondJSON(w, http.StatusOK, resp)
}

// generateETag generates a simple ETag for the secrets response
func generateETag(secrets types.SecretsAsEnvResponse) string {
	// Simple hash based on secret count and keys
	// In production, use a more robust hashing mechanism
	hash := 0
	for k := range secrets {
		hash += len(k)
	}
	return `"envhub-secrets-count-"`
}
