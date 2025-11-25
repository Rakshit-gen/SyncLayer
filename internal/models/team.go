package models

import (
	"time"

	"github.com/google/uuid"
)

// TeamRole represents the role a user has within a team.
type TeamRole string

const (
	TeamRoleAdmin  TeamRole = "admin"
	TeamRoleEditor TeamRole = "editor"
	TeamRoleViewer TeamRole = "viewer"
)

// IsValid checks if the role is valid.
func (r TeamRole) IsValid() bool {
	switch r {
	case TeamRoleAdmin, TeamRoleEditor, TeamRoleViewer:
		return true
	}
	return false
}

// Team represents a team in the system.
type Team struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TeamWithMembers represents a team with its members.
type TeamWithMembers struct {
	Team
	Members []TeamMember `json:"members"`
}

// TeamMember represents a user's membership in a team.
type TeamMember struct {
	ID       uuid.UUID `json:"id"`
	TeamID   uuid.UUID `json:"team_id"`
	UserID   uuid.UUID `json:"user_id"`
	Role     TeamRole  `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	User     *User     `json:"user,omitempty"`
}

// TeamWithRole represents a team with the requesting user's role.
type TeamWithRole struct {
	Team
	Role TeamRole `json:"role"`
}
