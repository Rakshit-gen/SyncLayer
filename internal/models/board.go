package models

import (
	"time"

	"github.com/google/uuid"
)

// Board represents a task board.
type Board struct {
	ID          uuid.UUID  `json:"id"`
	TeamID      uuid.UUID  `json:"team_id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// BoardWithColumns represents a board with its columns.
type BoardWithColumns struct {
	Board
	Columns []Column `json:"columns"`
}

// BoardFull represents a board with columns and tasks.
type BoardFull struct {
	Board
	Columns []ColumnWithTasks `json:"columns"`
}
