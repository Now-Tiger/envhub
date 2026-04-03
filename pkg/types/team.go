package types

import (
	"time"

	"github.com/google/uuid"
)

// Role represents organization/team member roles
type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
	RoleViewer Role = "viewer"
)

// AddMemberRequest represents the request to add a member to a project
type AddMemberRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
	Role   Role   `json:"role" validate:"required,oneof=owner admin member viewer"`
}

// UpdateMemberRoleRequest represents the request to update a member's role
type UpdateMemberRoleRequest struct {
	Role Role `json:"role" validate:"required,oneof=owner admin member viewer"`
}

// MemberResponse represents a team member in API responses
type MemberResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Role      Role      `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

// ListMembersResponse represents a list of team members
type ListMembersResponse struct {
	Members []MemberResponse `json:"members"`
}
