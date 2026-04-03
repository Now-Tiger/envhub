package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/Now-Tiger/envhub/config"
	"github.com/Now-Tiger/envhub/internal/middleware"
	"github.com/Now-Tiger/envhub/internal/repository"
	"github.com/Now-Tiger/envhub/pkg/types"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	repo repository.Querier
	cfg  config.AuthConfig
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(repo repository.Querier, cfg config.AuthConfig) *AuthHandler {
	return &AuthHandler{
		repo: repo,
		cfg:  cfg,
	}
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req types.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	// Check if user already exists
	_, err := h.repo.GetUserByEmail(r.Context(), req.Email)
	if err == nil {
		respondError(w, http.StatusConflict, "user already exists", "")
		return
	}

	// Hash password
	passwordHash := hashPassword(req.Password)

	// Create user
	user, err := h.repo.CreateUser(r.Context(), repository.CreateUserParams{
		Email:          req.Email,
		FullName:       &req.FullName,
		PasswordHash:   &passwordHash,
		IsActive:       toBoolPtr(true),
		EmailVerified:  toBoolPtr(false),
		AuthProviderID: nil,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create user", err.Error())
		return
	}

	// Create default organization for the user
	orgName := req.FullName + "'s Organization"
	orgSlug := strings.ToLower(strings.ReplaceAll(req.FullName, " ", "-") + "-" + user.ID.String()[:8])
	org, err := h.repo.CreateOrganization(r.Context(), repository.CreateOrganizationParams{
		Name:                 orgName,
		Slug:                 orgSlug,
		PlanType:             toStrPtr("free"),
		MaxProjects:          toInt32Ptr(3),
		MaxSecretsPerProject: toInt32Ptr(50),
		OwnerID:              user.ID,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create organization", err.Error())
		return
	}

	// Insert into organization_members for the owner
	_, err = h.repo.CreateOrganizationMember(r.Context(), repository.CreateOrganizationMemberParams{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           "owner",
		JoinedAt:       toTimestamptzPtr(time.Now()),
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to add owner to organization", err.Error())
		return
	}

	// Generate JWT
	token, err := h.GenerateJWT(user.ID, user.Email, "user")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token", err.Error())
		return
	}

	resp := types.AuthResponse{
		Token: token,
		User: types.CurrentUserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FullName:  getStringPtr(user.FullName),
			AvatarURL: getStringPtr(user.AvatarUrl),
		},
		ExpiresAt: time.Now().Add(h.cfg.JWTExpiry),
	}

	respondJSON(w, http.StatusCreated, resp)
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req types.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	// Get user by email
	user, err := h.repo.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials", "")
		return
	}

	// Verify password
	if user.PasswordHash == nil || !verifyPassword(req.Password, *user.PasswordHash) {
		respondError(w, http.StatusUnauthorized, "invalid credentials", "")
		return
	}

	// Check if user is active
	if user.IsActive != nil && !*user.IsActive {
		respondError(w, http.StatusForbidden, "account is disabled", "")
		return
	}

	// Update last login
	_ = h.repo.UpdateUserLastLogin(r.Context(), user.ID)

	// Generate JWT
	token, err := h.GenerateJWT(user.ID, user.Email, "user")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token", err.Error())
		return
	}

	resp := types.AuthResponse{
		Token: token,
		User: types.CurrentUserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FullName:  getStringPtr(user.FullName),
			AvatarURL: getStringPtr(user.AvatarUrl),
		},
		ExpiresAt: time.Now().Add(h.cfg.JWTExpiry),
	}

	respondJSON(w, http.StatusOK, resp)
}

// hashPassword creates a secure bcrypt hash of the password
func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash)
}

// verifyPassword compares password with bcrypt hash
func verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// toBoolPtr returns pointer to bool
func toBoolPtr(b bool) *bool {
	return &b
}

// toStrPtr returns pointer to string
func toStrPtr(s string) *string {
	return &s
}

// toInt32Ptr returns pointer to int32
func toInt32Ptr(i int32) *int32 {
	return &i
}

// CLILogin handles POST /auth/cli-login
func (h *AuthHandler) CLILogin(w http.ResponseWriter, r *http.Request) {
	var req types.CLILoginRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	// Parse project ID
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "")
		return
	}

	// Get project to verify it exists and get organization
	project, err := h.repo.GetProjectByID(r.Context(), projectID)
	if err != nil {
		respondError(w, http.StatusNotFound, "project not found", "")
		return
	}

	// Get user from context (this is for authenticated users creating CLI tokens)
	// For CLI login, we generate a token that can be used for API access
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		// For unauthenticated CLI login, we need a different flow
		// This would typically involve a device flow or similar
		respondError(w, http.StatusUnauthorized, "authentication required", "")
		return
	}

	// Generate a random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token", "")
		return
	}
	token := hex.EncodeToString(tokenBytes)

	// Hash the token for storage
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Set token expiry (default 30 days for CLI)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	// Create API token in database
	_, err = h.repo.CreateAPIToken(r.Context(), repository.CreateAPITokenParams{
		UserID:         userID,
		Name:           "CLI Token",
		TokenHash:      tokenHash,
		Scopes:         []string{"read", "write"},
		OrganizationID: toPGUUIDPtr(project.OrganizationID),
		ExpiresAt:      toTimestamptzPtr(expiresAt),
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create API token", err.Error())
		return
	}

	resp := types.CLILoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}

	respondJSON(w, http.StatusOK, resp)
}

// Me handles GET /auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "")
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found", "")
		return
	}

	resp := types.CurrentUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  getStringPtr(user.FullName),
		AvatarURL: getStringPtr(user.AvatarUrl),
	}

	respondJSON(w, http.StatusOK, resp)
}

// GenerateJWT generates a JWT token for a user
func (h *AuthHandler) GenerateJWT(userID uuid.UUID, email string, role string) (string, error) {
	claims := middleware.JWTClaims{
		UserID: userID.String(),
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.cfg.JWTExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}

// toPGUUIDPtr returns pointer to pgtype.UUID
func toPGUUIDPtr(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// toTimestamptzPtr returns pgtype.Timestamptz
func toTimestamptzPtr(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// ListTokens handles GET /auth/tokens - list user's API tokens
func (h *AuthHandler) ListTokens(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "")
		return
	}

	tokens, err := h.repo.ListUserAPITokens(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list tokens", err.Error())
		return
	}

	type tokenResponse struct {
		ID         string    `json:"id"`
		Name       string    `json:"name"`
		Scopes     []string  `json:"scopes"`
		ExpiresAt  time.Time `json:"expires_at,omitempty"`
		CreatedAt  time.Time `json:"created_at"`
		LastUsedAt time.Time `json:"last_used_at,omitempty"`
	}

	resp := make([]tokenResponse, len(tokens))
	for i, t := range tokens {
		resp[i] = tokenResponse{
			ID:         t.ID.String(),
			Name:       t.Name,
			Scopes:     t.Scopes,
			ExpiresAt:  t.ExpiresAt.Time,
			CreatedAt:  t.CreatedAt,
			LastUsedAt: t.LastUsedAt.Time,
		}
	}

	respondJSON(w, http.StatusOK, resp)
}

// RevokeToken handles DELETE /auth/tokens/:id - revoke an API token
func (h *AuthHandler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "")
		return
	}

	tokenID := chi.URLParam(r, "id")
	tokenUUID, err := uuid.Parse(tokenID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid token ID", "")
		return
	}

	// Verify token belongs to user before revoking
	token, err := h.repo.GetAPITokenByID(r.Context(), tokenUUID)
	if err != nil || token.UserID != userID {
		respondError(w, http.StatusNotFound, "token not found", "")
		return
	}

	err = h.repo.RevokeAPIToken(r.Context(), tokenUUID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to revoke token", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getStringPtr returns string from *string
func getStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
