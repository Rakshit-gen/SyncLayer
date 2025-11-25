package models

import (
	"time"

	"github.com/google/uuid"
)

// TaskPriority represents the priority level of a task.
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

// IsValid checks if the priority is valid.
func (p TaskPriority) IsValid() bool {
	switch p {
	case TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh, TaskPriorityUrgent:
		return true
	}
	return false
}

// Task represents a task in a column.
type Task struct {
	ID          uuid.UUID    `json:"id"`
	ColumnID    uuid.UUID    `json:"column_id"`
	Title       string       `json:"title"`
	Description *string      `json:"description,omitempty"`
	Position    int          `json:"position"`
	Priority    TaskPriority `json:"priority"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	AssignedTo  *uuid.UUID   `json:"assigned_to,omitempty"`
	CreatedBy   *uuid.UUID   `json:"created_by,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// TaskWithAssignee represents a task with assignee details.
type TaskWithAssignee struct {
	Task
	Assignee *User `json:"assignee,omitempty"`
	Creator  *User `json:"creator,omitempty"`
}
