package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/Now-Tiger/envhub/internal/repository"
	"github.com/Now-Tiger/envhub/pkg/crypto"
	"github.com/Now-Tiger/envhub/pkg/types"
)

// ErrProjectNotFound represents a project not found error
var ErrProjectNotFound = errors.New("project not found")

// ProjectService handles project business logic
type ProjectService struct {
	repo      repository.Querier
	masterKey *crypto.MasterKey
}

// NewProjectService creates a new ProjectService
func NewProjectService(repo repository.Querier, masterKey *crypto.MasterKey) *ProjectService {
	return &ProjectService{
		repo:      repo,
		masterKey: masterKey,
	}
}

// Create creates a new project with a new DEK
func (s *ProjectService) Create(ctx context.Context, userID, orgID uuid.UUID, req types.CreateProjectRequest) (*types.ProjectResponse, error) {
	// Generate a new Data Encryption Key (DEK)
	dek, err := crypto.GenerateDataKey()
	if err != nil {
		return nil, err
	}

	// Encrypt the DEK with the master key
	encryptedDEK, err := crypto.EncryptDEK(dek, s.masterKey)
	if err != nil {
		return nil, err
	}

	// Create project in database
	project, err := s.repo.CreateProject(ctx, repository.CreateProjectParams{
		OrganizationID: orgID,
		Name:           req.Name,
		Description:    toPtr(req.Description),
		EncryptedDek:   encryptedDEK,
		DekVersion:     1,
		Color:          toPtr(req.Color),
		Icon:           toPtr(req.Icon),
	})
	if err != nil {
		return nil, err
	}

	return s.toResponse(project), nil
}

// GetByID retrieves a project by ID
func (s *ProjectService) GetByID(ctx context.Context, userID, projectID uuid.UUID) (*types.ProjectResponse, error) {
	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	return s.toResponse(project), nil
}

// List retrieves projects for an organization with pagination
func (s *ProjectService) List(ctx context.Context, orgID uuid.UUID, params types.PaginationParams) (*types.ListProjectsResponse, error) {
	// Get total count
	totalItems, err := s.repo.CountProjectsByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Calculate offset
	offset := (params.Page - 1) * params.PageSize

	// Get projects
	projects, err := s.repo.ListProjectsByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Apply pagination
	if offset > len(projects) {
		projects = []repository.Project{}
	} else {
		projects = projects[offset : offset+params.PageSize]
	}

	// Convert to responses - handle nil projects gracefully
	if len(projects) == 0 {
		return &types.ListProjectsResponse{
			Projects: []types.ProjectResponse{},
			Pagination: types.PaginationResponse{
				Page:       params.Page,
				PageSize:   params.PageSize,
				TotalItems: 0,
				TotalPages: 0,
			},
		}, nil
	}

	items := make([]types.ProjectResponse, 0, len(projects))
	for _, p := range projects {
		resp := s.toResponse(p)
		if resp != nil {
			items = append(items, *resp)
		}
	}

	return &types.ListProjectsResponse{
		Projects: items,
		Pagination: types.PaginationResponse{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: totalItems,
			TotalPages: int((totalItems + int64(params.PageSize) - 1) / int64(params.PageSize)),
		},
	}, nil
}

// Update updates a project
func (s *ProjectService) Update(ctx context.Context, userID, projectID uuid.UUID, req types.UpdateProjectRequest) (*types.ProjectResponse, error) {
	project, err := s.repo.UpdateProject(ctx, repository.UpdateProjectParams{
		ID:          projectID,
		Name:        req.Name,
		Description: toPtr(req.Description),
		Color:       toPtr(req.Color),
		Icon:        toPtr(req.Icon),
	})
	if err != nil {
		return nil, err
	}

	return s.toResponse(project), nil
}

// Delete soft-deletes a project
func (s *ProjectService) Delete(ctx context.Context, userID, projectID uuid.UUID) error {
	return s.repo.SoftDeleteProject(ctx, projectID)
}

// RotateDEK rotates the Data Encryption Key for a project
func (s *ProjectService) RotateDEK(ctx context.Context, userID, projectID uuid.UUID) error {
	// Get current project to get encrypted DEK
	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return ErrProjectNotFound
	}

	// Decrypt current DEK (for future re-encryption of secrets)
	_, err = crypto.DecryptDEK(project.EncryptedDek, s.masterKey)
	if err != nil {
		return err
	}

	// Generate new DEK
	newDEK, err := crypto.GenerateDataKey()
	if err != nil {
		return err
	}

	// Encrypt new DEK with master key
	encryptedNewDEK, err := crypto.EncryptDEK(newDEK, s.masterKey)
	if err != nil {
		return err
	}

	// Update project with new encrypted DEK
	_, err = s.repo.RotateProjectDEK(ctx, repository.RotateProjectDEKParams{
		ID:           projectID,
		EncryptedDek: encryptedNewDEK,
	})
	return err
}

// toResponse converts a repository Project to a response
func (s *ProjectService) toResponse(project repository.Project) *types.ProjectResponse {
	resp := &types.ProjectResponse{
		ID:         project.ID,
		Name:       project.Name,
		DEKVersion: int(project.DekVersion),
		CreatedAt:  project.CreatedAt,
		UpdatedAt:  project.UpdatedAt,
	}

	if project.Description != nil {
		resp.Description = *project.Description
	}
	if project.Color != nil {
		resp.Color = *project.Color
	}
	if project.Icon != nil {
		resp.Icon = *project.Icon
	}

	return resp
}

// toPtr returns pointer to string
func toPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
