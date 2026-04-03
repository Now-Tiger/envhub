package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/Now-Tiger/envhub/internal/repository"
	"github.com/Now-Tiger/envhub/pkg/types"
)

// ErrEnvNotFound represents an environment not found error
var ErrEnvNotFound = errors.New("environment not found")

// EnvironmentService handles environment business logic
type EnvironmentService struct {
	repo repository.Querier
}

// NewEnvironmentService creates a new EnvironmentService
func NewEnvironmentService(repo repository.Querier) *EnvironmentService {
	return &EnvironmentService{
		repo: repo,
	}
}

// Create creates a new environment
func (s *EnvironmentService) Create(ctx context.Context, projectID uuid.UUID, req types.CreateEnvironmentRequest) (*types.EnvironmentResponse, error) {
	env, err := s.repo.CreateEnvironment(ctx, repository.CreateEnvironmentParams{
		ProjectID:   projectID,
		Name:        req.Name,
		Description: toStrPtr(req.Description),
		Color:       toStrPtr(req.Color),
		IsProtected: toBoolPtr(req.IsProtected),
	})
	if err != nil {
		return nil, err
	}

	return s.toResponse(env), nil
}

// GetByID retrieves an environment by ID
func (s *EnvironmentService) GetByID(ctx context.Context, envID uuid.UUID) (*types.EnvironmentResponse, error) {
	env, err := s.repo.GetEnvironmentByID(ctx, envID)
	if err != nil {
		return nil, ErrEnvNotFound
	}

	return s.toResponse(env), nil
}

// List retrieves environments for a project
func (s *EnvironmentService) List(ctx context.Context, projectID uuid.UUID) (*types.ListEnvironmentsResponse, error) {
	envs, err := s.repo.ListEnvironmentsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	items := make([]types.EnvironmentResponse, len(envs))
	for i, e := range envs {
		items[i] = *s.toResponse(e)
	}

	return &types.ListEnvironmentsResponse{
		Environments: items,
	}, nil
}

// Update updates an environment
func (s *EnvironmentService) Update(ctx context.Context, envID uuid.UUID, req types.UpdateEnvironmentRequest) (*types.EnvironmentResponse, error) {
	env, err := s.repo.UpdateEnvironment(ctx, repository.UpdateEnvironmentParams{
		ID:          envID,
		Description: toStrPtr(req.Description),
		Color:       toStrPtr(req.Color),
		IsProtected: toBoolPtr(req.IsProtected),
	})
	if err != nil {
		return nil, err
	}

	return s.toResponse(env), nil
}

// Delete deletes an environment
func (s *EnvironmentService) Delete(ctx context.Context, envID uuid.UUID) error {
	return s.repo.DeleteEnvironment(ctx, envID)
}

// toResponse converts a repository Environment to a response
func (s *EnvironmentService) toResponse(env repository.Environment) *types.EnvironmentResponse {
	resp := &types.EnvironmentResponse{
		ID:        env.ID,
		Name:      env.Name,
		CreatedAt: env.CreatedAt,
		UpdatedAt: env.UpdatedAt,
	}

	if env.Description != nil {
		resp.Description = *env.Description
	}
	if env.Color != nil {
		resp.Color = *env.Color
	}
	if env.IsProtected != nil {
		resp.IsProtected = *env.IsProtected
	}

	return resp
}
