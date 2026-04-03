package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/Now-Tiger/envhub/pkg/types"
)

// ErrForbidden represents a forbidden access error
var ErrForbidden = errors.New("forbidden")

// roleHierarchy defines the hierarchy of roles (higher index = more permissions)
var roleHierarchy = map[types.Role]int{
	types.RoleViewer: 0,
	types.RoleMember: 1,
	types.RoleAdmin:  2,
	types.RoleOwner:  3,
}

// RequireRole creates middleware that checks if user has one of the allowed roles
func RequireRole(allowedRoles ...types.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user role from context
			roleStr, ok := r.Context().Value(RoleKey).(string)
			if !ok {
				http.Error(w, "Role not found in context", http.StatusForbidden)
				return
			}

			userRole := types.Role(roleStr)

			// Check if user's role is in allowed roles
			for _, allowed := range allowedRoles {
				if userRole == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "Insufficient permissions", http.StatusForbidden)
		})
	}
}

// RequireProjectAccess creates middleware that checks project membership and role
func RequireProjectAccess(minRole types.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get project ID from URL params
			projectID := r.Context().Value(ProjectIDKey)
			if projectID == nil {
				http.Error(w, "Project ID not found in context", http.StatusForbidden)
				return
			}

			// Get user role from context
			roleStr, ok := r.Context().Value(RoleKey).(string)
			if !ok {
				http.Error(w, "Role not found in context", http.StatusForbidden)
				return
			}

			userRole := types.Role(roleStr)

			// Get required role level
			minLevel, ok := roleHierarchy[minRole]
			if !ok {
				http.Error(w, "Invalid required role", http.StatusInternalServerError)
				return
			}

			// Get user's role level
			userLevel, ok := roleHierarchy[userRole]
			if !ok {
				http.Error(w, "Invalid user role", http.StatusForbidden)
				return
			}

			// Check if user has sufficient permissions
			if userLevel < minLevel {
				http.Error(w, "Insufficient project access", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// HasSufficientRole checks if user role meets the minimum required role
func HasSufficientRole(userRole, minRole types.Role) bool {
	userLevel, ok := roleHierarchy[userRole]
	if !ok {
		return false
	}

	minLevel, ok := roleHierarchy[minRole]
	if !ok {
		return false
	}

	return userLevel >= minLevel
}

// GetUserRoleFromContext extracts user role from context
func GetUserRoleFromContext(ctx context.Context) (types.Role, bool) {
	role, ok := ctx.Value(RoleKey).(string)
	if !ok {
		return "", false
	}
	return types.Role(role), true
}
