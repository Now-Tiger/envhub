package types

import (
	"time"

	"github.com/google/uuid"
)

// CreateSecretRequest represents the request to create a new secret
type CreateSecretRequest struct {
	Key         string `json:"key" validate:"required,max=255"`
	Value       string `json:"value" validate:"required"`
	Description string `json:"description" validate:"max=500"`
}

// UpdateSecretRequest represents the request to update a secret
type UpdateSecretRequest struct {
	Value       string `json:"value" validate:"required"`
	Description string `json:"description" validate:"max=500"`
}

// SecretResponse represents a secret in API responses (value is never exposed)
type SecretResponse struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Description string    `json:"description,omitempty"`
	Version     int       `json:"version"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListSecretsResponse represents a list of secrets with pagination
type ListSecretsResponse struct {
	Secrets    []SecretResponse   `json:"secrets"`
	Pagination PaginationResponse `json:"pagination"`
}

// SecretsAsEnvResponse represents secrets as environment variables for CLI
type SecretsAsEnvResponse map[string]string
