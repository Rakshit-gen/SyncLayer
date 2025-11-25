package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ActivityAction represents the type of activity.
type ActivityAction string

const (
	ActivityTaskCreated     ActivityAction = "task.created"
	ActivityTaskUpdated     ActivityAction = "task.updated"
	ActivityTaskMoved       ActivityAction = "task.moved"
	ActivityTaskDeleted     ActivityAction = "task.deleted"
	ActivityTaskAssigned    ActivityAction = "task.assigned"
	ActivityColumnCreated   ActivityAction = "column.created"
	ActivityColumnUpdated   ActivityAction = "column.updated"
	ActivityColumnMoved     ActivityAction = "column.moved"
	ActivityColumnDeleted   ActivityAction = "column.deleted"
	ActivityBoardCreated    ActivityAction = "board.created"
	ActivityBoardUpdated    ActivityAction = "board.updated"
	ActivityMemberAdded     ActivityAction = "member.added"
	ActivityMemberRemoved   ActivityAction = "member.removed"
	ActivityMemberRoleChanged ActivityAction = "member.role_changed"
)

// ActivityLog represents an activity log entry.
type ActivityLog struct {
	ID        uuid.UUID       `json:"id"`
	TaskID    *uuid.UUID      `json:"task_id,omitempty"`
	BoardID   *uuid.UUID      `json:"board_id,omitempty"`
	UserID    *uuid.UUID      `json:"user_id,omitempty"`
	Action    ActivityAction  `json:"action"`
	Details   json.RawMessage `json:"details,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// ActivityLogWithUser represents an activity log with user details.
type ActivityLogWithUser struct {
	ActivityLog
	User *User `json:"user,omitempty"`
}

// ActivityDetails holds additional information about an activity.
type ActivityDetails struct {
	PreviousValue interface{} `json:"previous_value,omitempty"`
	NewValue      interface{} `json:"new_value,omitempty"`
	Field         string      `json:"field,omitempty"`
	FromColumnID  *uuid.UUID  `json:"from_column_id,omitempty"`
	ToColumnID    *uuid.UUID  `json:"to_column_id,omitempty"`
}
