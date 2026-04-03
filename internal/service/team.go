package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Now-Tiger/envhub/internal/repository"
	"github.com/Now-Tiger/envhub/pkg/types"
)

// ErrMemberNotFound represents a member not found error
var ErrMemberNotFound = errors.New("member not found")

// ErrUserNotFound represents a user not found error
var ErrUserNotFound = errors.New("user not found")

// ErrAlreadyMember represents already a member error
var ErrAlreadyMember = errors.New("user is already a member")

// TeamService handles team/organization member business logic
type TeamService struct {
	repo repository.Querier
}

// NewTeamService creates a new TeamService
func NewTeamService(repo repository.Querier) *TeamService {
	return &TeamService{
		repo: repo,
	}
}

// AddMember adds a member to an organization
func (s *TeamService) AddMember(ctx context.Context, orgID, projectID uuid.UUID, req types.AddMemberRequest) (*types.MemberResponse, error) {
	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// Check if user is already a member
	_, err = s.repo.GetOrganizationMember(ctx, repository.GetOrganizationMemberParams{
		OrganizationID: orgID,
		UserID:         userUUID,
	})
	if err == nil {
		return nil, ErrAlreadyMember
	}

	// Add user to organization
	member, err := s.repo.CreateOrganizationMember(ctx, repository.CreateOrganizationMemberParams{
		OrganizationID: orgID,
		UserID:         userUUID,
		Role:           string(req.Role),
		JoinedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return nil, errors.New("failed to add member")
	}

	// Get user details
	user, err := s.repo.GetUserByID(ctx, userUUID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &types.MemberResponse{
		UserID:    member.UserID,
		Email:     user.Email,
		FullName:  getStringPtr(user.FullName),
		AvatarURL: getStringPtr(user.AvatarUrl),
		Role:      types.Role(member.Role),
	}, nil
}

// RemoveMember removes a member from an organization
func (s *TeamService) RemoveMember(ctx context.Context, orgID, projectID, userID uuid.UUID) error {
	err := s.repo.DeleteOrganizationMember(ctx, repository.DeleteOrganizationMemberParams{
		OrganizationID: orgID,
		UserID:         userID,
	})
	if err != nil {
		return ErrMemberNotFound
	}
	return nil
}

// UpdateMemberRole updates a member's role
func (s *TeamService) UpdateMemberRole(ctx context.Context, orgID, projectID, userID uuid.UUID, req types.UpdateMemberRoleRequest) (*types.MemberResponse, error) {
	member, err := s.repo.UpdateOrganizationMemberRole(ctx, repository.UpdateOrganizationMemberRoleParams{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           string(req.Role),
	})
	if err != nil {
		return nil, ErrMemberNotFound
	}

	// Get user details
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &types.MemberResponse{
		UserID:    member.UserID,
		Email:     user.Email,
		FullName:  getStringPtr(user.FullName),
		AvatarURL: getStringPtr(user.AvatarUrl),
		Role:      types.Role(member.Role),
	}, nil
}

// ListMembers lists all members of an organization
func (s *TeamService) ListMembers(ctx context.Context, projectID uuid.UUID) (*types.ListMembersResponse, error) {
	// Get project to find orgID
	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	members, err := s.repo.ListOrganizationMembers(ctx, project.OrganizationID)
	if err != nil {
		return nil, errors.New("failed to list members")
	}

	resp := &types.ListMembersResponse{
		Members: make([]types.MemberResponse, len(members)),
	}

	for i, m := range members {
		resp.Members[i] = types.MemberResponse{
			UserID:    m.UserID,
			Email:     m.Email,
			FullName:  m.FullName.String,
			AvatarURL: m.AvatarUrl.String,
			Role:      types.Role(m.Role),
		}
	}

	return resp, nil
}

// GetUserByID retrieves a user by ID
func (s *TeamService) GetUserByID(ctx context.Context, userID uuid.UUID) (*types.MemberResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &types.MemberResponse{
		UserID:    user.ID,
		Email:     user.Email,
		FullName:  getStringPtr(user.FullName),
		AvatarURL: getStringPtr(user.AvatarUrl),
	}, nil
}

// getStringPtr returns pointer to string or nil
func getStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
