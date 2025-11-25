// Package dto defines Data Transfer Objects for API requests and responses.
package dto

import (
	"time"

	"github.com/google/uuid"
	"taskboard-backend/internal/models"
)

// ===============================
// User DTOs
// ===============================

// CreateUserRequest represents a request to create a new user.
type CreateUserRequest struct {
	Email     string  `json:"email" validate:"required,email"`
	Name      string  `json:"name" validate:"required,min=1,max=255"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// UpdateUserRequest represents a request to update a user.
type UpdateUserRequest struct {
	Name      *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// ===============================
// Team DTOs
// ===============================

// CreateTeamRequest represents a request to create a new team.
type CreateTeamRequest struct {
	Name        string    `json:"name" validate:"required,min=1,max=255"`
	Description *string   `json:"description,omitempty"`
	CreatedBy   uuid.UUID `json:"created_by" validate:"required"`
}

// UpdateTeamRequest represents a request to update a team.
type UpdateTeamRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty"`
}

// AddTeamMemberRequest represents a request to add a team member.
type AddTeamMemberRequest struct {
	UserID uuid.UUID       `json:"user_id" validate:"required"`
	Role   models.TeamRole `json:"role" validate:"required,oneof=admin editor viewer"`
}

// UpdateTeamMemberRequest represents a request to update a team member's role.
type UpdateTeamMemberRequest struct {
	Role models.TeamRole `json:"role" validate:"required,oneof=admin editor viewer"`
}

// ===============================
// Board DTOs
// ===============================

// CreateBoardRequest represents a request to create a new board.
type CreateBoardRequest struct {
	TeamID      uuid.UUID `json:"team_id" validate:"required"`
	Name        string    `json:"name" validate:"required,min=1,max=255"`
	Description *string   `json:"description,omitempty"`
	CreatedBy   uuid.UUID `json:"created_by" validate:"required"`
}

// UpdateBoardRequest represents a request to update a board.
type UpdateBoardRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty"`
}

// ===============================
// Column DTOs
// ===============================

// CreateColumnRequest represents a request to create a new column.
type CreateColumnRequest struct {
	BoardID  uuid.UUID `json:"board_id" validate:"required"`
	Name     string    `json:"name" validate:"required,min=1,max=255"`
	Position int       `json:"position" validate:"gte=0"`
	Color    *string   `json:"color,omitempty"`
}

// UpdateColumnRequest represents a request to update a column.
type UpdateColumnRequest struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Color *string `json:"color,omitempty"`
}

// MoveColumnRequest represents a request to move a column.
type MoveColumnRequest struct {
	Position int `json:"position" validate:"gte=0"`
}

// ===============================
// Task DTOs
// ===============================

// CreateTaskRequest represents a request to create a new task.
type CreateTaskRequest struct {
	ColumnID    uuid.UUID           `json:"column_id" validate:"required"`
	Title       string              `json:"title" validate:"required,min=1,max=500"`
	Description *string             `json:"description,omitempty"`
	Priority    *models.TaskPriority `json:"priority,omitempty"`
	DueDate     *time.Time          `json:"due_date,omitempty"`
	AssignedTo  *uuid.UUID          `json:"assigned_to,omitempty"`
	CreatedBy   uuid.UUID           `json:"created_by" validate:"required"`
}

// UpdateTaskRequest represents a request to update a task.
type UpdateTaskRequest struct {
	Title       *string             `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
	Description *string             `json:"description,omitempty"`
	Priority    *models.TaskPriority `json:"priority,omitempty"`
	DueDate     *time.Time          `json:"due_date,omitempty"`
	AssignedTo  *uuid.UUID          `json:"assigned_to,omitempty"`
}

// MoveTaskRequest represents a request to move a task.
type MoveTaskRequest struct {
	ColumnID uuid.UUID `json:"column_id" validate:"required"`
	Position int       `json:"position" validate:"gte=0"`
}

// ===============================
// Notification DTOs
// ===============================

// CreateNotificationRequest represents a request to create a notification.
type CreateNotificationRequest struct {
	UserID  uuid.UUID               `json:"user_id" validate:"required"`
	Type    models.NotificationType `json:"type" validate:"required"`
	Title   string                  `json:"title" validate:"required,min=1,max=255"`
	Message *string                 `json:"message,omitempty"`
	Data    interface{}             `json:"data,omitempty"`
}
