package types

import (
	"time"
)

// CLISecretsRequest represents the request to get secrets for CLI
type CLISecretsRequest struct {
	Project     string `json:"project" validate:"required"`
	Environment string `json:"environment" validate:"required"`
}

// CLISecretsResponse represents the response for CLI secrets endpoint
type CLISecretsResponse struct {
	Project     string               `json:"project"`
	Environment string               `json:"environment"`
	Secrets     SecretsAsEnvResponse `json:"secrets"`
	RetrievedAt time.Time            `json:"retrieved_at"`
	Version     string               `json:"version"`
}
