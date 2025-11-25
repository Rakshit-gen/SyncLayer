package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification.
type NotificationType string

const (
	NotificationTaskAssigned   NotificationType = "task.assigned"
	NotificationTaskDueSoon    NotificationType = "task.due_soon"
	NotificationTaskOverdue    NotificationType = "task.overdue"
	NotificationTaskComment    NotificationType = "task.comment"
	NotificationMemberAdded    NotificationType = "member.added"
	NotificationMemberRemoved  NotificationType = "member.removed"
	NotificationBoardInvite    NotificationType = "board.invite"
	NotificationTeamInvite     NotificationType = "team.invite"
)

// Notification represents a notification for a user.
type Notification struct {
	ID        uuid.UUID        `json:"id"`
	UserID    uuid.UUID        `json:"user_id"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   *string          `json:"message,omitempty"`
	Data      json.RawMessage  `json:"data,omitempty"`
	Read      bool             `json:"read"`
	CreatedAt time.Time        `json:"created_at"`
}

// NotificationData holds additional notification context.
type NotificationData struct {
	TaskID    *uuid.UUID `json:"task_id,omitempty"`
	BoardID   *uuid.UUID `json:"board_id,omitempty"`
	TeamID    *uuid.UUID `json:"team_id,omitempty"`
	ActorID   *uuid.UUID `json:"actor_id,omitempty"`
	ActorName string     `json:"actor_name,omitempty"`
}
