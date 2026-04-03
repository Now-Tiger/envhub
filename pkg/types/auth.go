package types

import (
	"time"

	"github.com/google/uuid"
)

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents the response after login/register
type AuthResponse struct {
	Token     string              `json:"token"`
	User      CurrentUserResponse `json:"user"`
	ExpiresAt time.Time           `json:"expires_at"`
}

// CLILoginRequest represents the request to generate a CLI token
type CLILoginRequest struct {
	ProjectID string `json:"project_id" validate:"required,uuid"`
}

// CLILoginResponse represents the response for CLI login
type CLILoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CurrentUserResponse represents the current authenticated user
type CurrentUserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
}
