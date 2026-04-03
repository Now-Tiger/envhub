package service

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Now-Tiger/envhub/internal/repository"
	"github.com/Now-Tiger/envhub/pkg/crypto"
	"github.com/Now-Tiger/envhub/pkg/types"
)

// ErrSecretNotFound represents a secret not found error
var ErrSecretNotFound = errors.New("secret not found")

// ErrEnvironmentNotFound represents an environment not found error
var ErrEnvironmentNotFound = errors.New("environment not found")

// SecretService handles secret business logic with encryption
type SecretService struct {
	repo      repository.Querier
	masterKey *crypto.MasterKey
}

// NewSecretService creates a new SecretService
func NewSecretService(repo repository.Querier, masterKey *crypto.MasterKey) *SecretService {
	return &SecretService{
		repo:      repo,
		masterKey: masterKey,
	}
}

// getDEK retrieves and decrypts the DEK for a project
func (s *SecretService) getDEK(ctx context.Context, projectID uuid.UUID) (*crypto.DataKey, error) {
	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	return crypto.DecryptDEK(project.EncryptedDek, s.masterKey)
}

// Create creates a new secret with encryption
func (s *SecretService) Create(ctx context.Context, userID, envID uuid.UUID, req types.CreateSecretRequest) (*types.SecretResponse, error) {
	// Get environment to find project
	env, err := s.repo.GetEnvironmentByID(ctx, envID)
	if err != nil {
		return nil, ErrEnvironmentNotFound
	}

	// Get and decrypt DEK
	dek, err := s.getDEK(ctx, env.ProjectID)
	if err != nil {
		return nil, err
	}

	// Encrypt the secret value (returns []byte, convert to base64 string)
	encryptedBytes, err := crypto.EncryptWithDEK([]byte(req.Value), dek)
	if err != nil {
		return nil, err
	}
	encryptedValue := base64.StdEncoding.EncodeToString(encryptedBytes)

	// Create secret in database
	secret, err := s.repo.CreateSecret(ctx, repository.CreateSecretParams{
		EnvironmentID:  envID,
		Key:            req.Key,
		EncryptedValue: encryptedValue,
		Description:    toStrPtr(req.Description),
		IsActive:       toBoolPtr(true),
		Version:        1,
		CreatedBy:      toPGUUID(userID),
	})
	if err != nil {
		return nil, err
	}

	return s.toResponse(secret), nil
}

// GetByID retrieves a single secret (without decrypted value)
func (s *SecretService) GetByID(ctx context.Context, userID, secretID uuid.UUID) (*types.SecretResponse, error) {
	secret, err := s.repo.GetSecretByID(ctx, secretID)
	if err != nil {
		return nil, ErrSecretNotFound
	}

	return s.toResponse(secret), nil
}

// GetByEnvironment retrieves all secrets for an environment as decrypted env vars
func (s *SecretService) GetByEnvironment(ctx context.Context, userID, projectID uuid.UUID, envName string) (types.SecretsAsEnvResponse, error) {
	// Get environment by name
	env, err := s.repo.GetEnvironmentByName(ctx, repository.GetEnvironmentByNameParams{
		ProjectID: projectID,
		Name:      envName,
	})
	if err != nil {
		return nil, ErrEnvironmentNotFound
	}

	// Get and decrypt DEK
	dek, err := s.getDEK(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Get all secrets for environment
	secrets, err := s.repo.ListSecretsByEnvironment(ctx, env.ID)
	if err != nil {
		return nil, err
	}

	// Decrypt and return as env vars
	result := make(types.SecretsAsEnvResponse)
	for _, sec := range secrets {
		if sec.IsActive != nil && *sec.IsActive {
			// Decode base64 first
			ciphertext, err := base64.StdEncoding.DecodeString(sec.EncryptedValue)
			if err != nil {
				continue
			}
			decrypted, err := crypto.DecryptWithDEK(ciphertext, dek)
			if err != nil {
				continue // Skip failed decryptions
			}
			result[sec.Key] = string(decrypted)
		}
	}

	return result, nil
}

// GetByProjectName retrieves all secrets for a project by project name
func (s *SecretService) GetByProjectName(ctx context.Context, userID uuid.UUID, projectName, envName string) (types.SecretsAsEnvResponse, error) {
	// Get project by name
	project, err := s.repo.GetProjectByName(ctx, projectName)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	return s.GetByEnvironment(ctx, userID, project.ID, envName)
}

// List retrieves secrets for an environment with pagination
func (s *SecretService) List(ctx context.Context, envID uuid.UUID, params types.PaginationParams) (*types.ListSecretsResponse, error) {
	// Get total count
	totalItems, err := s.repo.CountSecretsByEnvironment(ctx, envID)
	if err != nil {
		return nil, err
	}

	// Get secrets
	secrets, err := s.repo.ListSecretsByEnvironment(ctx, envID)
	if err != nil {
		return nil, err
	}

	// Apply pagination
	offset := (params.Page - 1) * params.PageSize
	if offset > len(secrets) {
		secrets = []repository.Secret{}
	} else {
		secrets = secrets[offset : offset+params.PageSize]
	}

	// Convert to responses
	items := make([]types.SecretResponse, len(secrets))
	for i, sec := range secrets {
		items[i] = *s.toResponse(sec)
	}

	return &types.ListSecretsResponse{
		Secrets: items,
		Pagination: types.PaginationResponse{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: totalItems,
			TotalPages: int((totalItems + int64(params.PageSize) - 1) / int64(params.PageSize)),
		},
	}, nil
}

// Update updates a secret with re-encryption
func (s *SecretService) Update(ctx context.Context, userID, secretID uuid.UUID, req types.UpdateSecretRequest) (*types.SecretResponse, error) {
	// Get secret to find environment/project
	sec, err := s.repo.GetSecretByID(ctx, secretID)
	if err != nil {
		return nil, ErrSecretNotFound
	}

	// Get environment to find project
	env, err := s.repo.GetEnvironmentByID(ctx, sec.EnvironmentID)
	if err != nil {
		return nil, ErrEnvironmentNotFound
	}

	// Get and decrypt DEK
	dek, err := s.getDEK(ctx, env.ProjectID)
	if err != nil {
		return nil, err
	}

	// Encrypt new value
	encryptedBytes, err := crypto.EncryptWithDEK([]byte(req.Value), dek)
	if err != nil {
		return nil, err
	}
	encryptedValue := base64.StdEncoding.EncodeToString(encryptedBytes)

	// Update secret in database
	updated, err := s.repo.UpdateSecret(ctx, repository.UpdateSecretParams{
		ID:             secretID,
		EncryptedValue: encryptedValue,
		Description:    toStrPtr(req.Description),
		UpdatedBy:      toPGUUID(userID),
	})
	if err != nil {
		return nil, err
	}

	return s.toResponse(updated), nil
}

// Delete soft-deletes a secret
func (s *SecretService) Delete(ctx context.Context, userID, secretID uuid.UUID) error {
	_, err := s.repo.GetSecretByID(ctx, secretID)
	if err != nil {
		return ErrSecretNotFound
	}

	return s.repo.SoftDeleteSecret(ctx, repository.SoftDeleteSecretParams{
		ID:        secretID,
		UpdatedBy: toPGUUID(userID),
	})
}

// toResponse converts a repository Secret to a response
func (s *SecretService) toResponse(sec repository.Secret) *types.SecretResponse {
	resp := &types.SecretResponse{
		ID:        sec.ID,
		Key:       sec.Key,
		Version:   int(sec.Version),
		CreatedAt: sec.CreatedAt,
		UpdatedAt: sec.UpdatedAt,
	}

	if sec.Description != nil {
		resp.Description = *sec.Description
	}
	// Note: CreatedBy is pgtype.UUID, skipping for now as it requires special handling

	return resp
}

// toStrPtr returns pointer to string
func toStrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// toBoolPtr returns pointer to bool
func toBoolPtr(b bool) *bool {
	return &b
}

// toPGUUID converts uuid.UUID to pgtype.UUID
func toPGUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}
