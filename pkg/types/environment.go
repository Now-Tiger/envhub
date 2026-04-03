package types

import (
	"time"

	"github.com/google/uuid"
)

// CreateEnvironmentRequest represents the request to create a new environment
type CreateEnvironmentRequest struct {
	Name        string `json:"name" validate:"required,max=50"`
	Description string `json:"description" validate:"max=500"`
	Color       string `json:"color" validate:"omitempty,hexcolor"`
	IsProtected bool   `json:"is_protected"`
}

// UpdateEnvironmentRequest represents the request to update an environment
type UpdateEnvironmentRequest struct {
	Name        string `json:"name" validate:"max=50"`
	Description string `json:"description" validate:"max=500"`
	Color       string `json:"color" validate:"omitempty,hexcolor"`
	IsProtected bool   `json:"is_protected"`
}

// EnvironmentResponse represents an environment in API responses
type EnvironmentResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Color       string    `json:"color,omitempty"`
	IsProtected bool      `json:"is_protected"`
	SecretCount int       `json:"secret_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListEnvironmentsResponse represents a list of environments
type ListEnvironmentsResponse struct {
	Environments []EnvironmentResponse `json:"environments"`
}
