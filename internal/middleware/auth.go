package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/Now-Tiger/envhub/config"
	"github.com/Now-Tiger/envhub/internal/repository"
)

// Context keys for request context
type contextKey string

const (
	UserIDKey         contextKey = "user_id"
	EmailKey          contextKey = "email"
	RoleKey           contextKey = "role"
	ProjectIDKey      contextKey = "project_id"
	OrganizationIDKey contextKey = "organization_id"
)

// JWT claims structure
type JWTClaims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// ErrInvalidToken represents an invalid token error
var ErrInvalidToken = errors.New("invalid token")

// ErrExpiredToken represents an expired token error
var ErrExpiredToken = errors.New("token has expired")

// ErrUnauthorized represents an unauthorized error
var ErrUnauthorized = errors.New("unauthorized")

// JWTAuthMiddleware creates JWT authentication middleware
func JWTAuthMiddleware(cfg config.AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Check Bearer token format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Parse and validate JWT
			token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, ErrInvalidToken
				}
				return []byte(cfg.JWTSecret), nil
			})

			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired) {
					http.Error(w, "Token has expired", http.StatusUnauthorized)
					return
				}
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(*JWTClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Validate user ID
			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// APITokenMiddleware creates API token authentication middleware
func APITokenMiddleware(querier repository.Querier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// No token, pass to next handler (JWT might be used instead)
				next.ServeHTTP(w, r)
				return
			}

			// Check Bearer token format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				next.ServeHTTP(w, r)
				return
			}

			tokenString := parts[1]

			// Hash the token using SHA256
			hash := sha256.Sum256([]byte(tokenString))
			tokenHash := hex.EncodeToString(hash[:])

			// Lookup token in database
			token, err := querier.GetAPITokenByHash(r.Context(), tokenHash)
			if err != nil {
				// Token not found, try JWT instead
				next.ServeHTTP(w, r)
				return
			}

			// Check if token is expired
			if token.ExpiresAt.Time.Before(time.Now()) {
				http.Error(w, "API token has expired", http.StatusUnauthorized)
				return
			}

			// Check if token is revoked
			if token.RevokedAt.Valid {
				http.Error(w, "API token has been revoked", http.StatusUnauthorized)
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), UserIDKey, token.UserID)

			// Update token usage
			_ = querier.UpdateTokenUsage(r.Context(), token.ID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// GetEmail extracts email from context
func GetEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(EmailKey).(string)
	return email, ok
}

// GetRole extracts role from context
func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(RoleKey).(string)
	return role, ok
}
