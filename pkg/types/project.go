package types

import (
	"time"

	"github.com/google/uuid"
)

// CreateProjectRequest represents the request to create a new project
type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description" validate:"max=500"`
	Color       string `json:"color" validate:"omitempty,hexcolor"`
	Icon        string `json:"icon" validate:"max=50"`
	OrgID       string `json:"org_id" validate:"omitempty,uuid"`
}

// UpdateProjectRequest represents the request to update a project
type UpdateProjectRequest struct {
	Name        string `json:"name" validate:"max=100"`
	Description string `json:"description" validate:"max=500"`
	Color       string `json:"color" validate:"omitempty,hexcolor"`
	Icon        string `json:"icon" validate:"max=50"`
}

// ProjectResponse represents a project in API responses
type ProjectResponse struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description,omitempty"`
	Color            string    `json:"color,omitempty"`
	Icon             string    `json:"icon,omitempty"`
	DEKVersion       int       `json:"dek_version"`
	MemberCount      int       `json:"member_count"`
	EnvironmentCount int       `json:"environment_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ListProjectsResponse represents a list of projects with pagination
type ListProjectsResponse struct {
	Projects   []ProjectResponse  `json:"projects"`
	Pagination PaginationResponse `json:"pagination"`
}
