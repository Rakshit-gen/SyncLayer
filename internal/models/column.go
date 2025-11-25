package models

import (
	"time"

	"github.com/google/uuid"
)

// Column represents a column in a board.
type Column struct {
	ID        uuid.UUID `json:"id"`
	BoardID   uuid.UUID `json:"board_id"`
	Name      string    `json:"name"`
	Position  int       `json:"position"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ColumnWithTasks represents a column with its tasks.
type ColumnWithTasks struct {
	Column
	Tasks []Task `json:"tasks"`
}
