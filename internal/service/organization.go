package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/Now-Tiger/envhub/internal/repository"
)

// OrganizationService handles organization business logic
type OrganizationService struct {
	repo repository.Querier
}

// NewOrganizationService creates a new OrganizationService
func NewOrganizationService(repo repository.Querier) *OrganizationService {
	return &OrganizationService{repo: repo}
}

// ListUserOrganizations retrieves all organizations a user belongs to
func (s *OrganizationService) ListUserOrganizations(ctx context.Context, userID uuid.UUID) ([]repository.Organization, error) {
	return s.repo.ListUserOrganizations(ctx, userID)
}

// ListOrganizationsByOwner retrieves all organizations owned by a user
func (s *OrganizationService) ListOrganizationsByOwner(ctx context.Context, userID uuid.UUID) ([]repository.Organization, error) {
	return s.repo.ListOrganizationsByOwner(ctx, userID)
}
