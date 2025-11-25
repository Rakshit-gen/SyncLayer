package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"taskboard-backend/internal/models"
	apperrors "taskboard-backend/pkg/errors"
)

// TaskRepository handles task data access operations.
type TaskRepository struct {
	pool *pgxpool.Pool
}

// NewTaskRepository creates a new TaskRepository instance.
func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

// Create creates a new task.
func (r *TaskRepository) Create(ctx context.Context, task *models.Task) (*models.Task, error) {
	// Shift existing tasks to make room
	shiftQuery := `
		UPDATE tasks
		SET position = position + 1
		WHERE column_id = $1 AND position >= $2
	`
	_, err := r.pool.Exec(ctx, shiftQuery, task.ColumnID, task.Position)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to shift tasks", err)
	}

	// Insert the new task
	query := `
		INSERT INTO tasks (column_id, title, description, position, priority, due_date, assigned_to, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, column_id, title, description, position, priority, due_date, assigned_to, created_by, created_at, updated_at
	`

	var result models.Task
	err = r.pool.QueryRow(ctx, query,
		task.ColumnID,
		task.Title,
		task.Description,
		task.Position,
		task.Priority,
		task.DueDate,
		task.AssignedTo,
		task.CreatedBy,
	).Scan(
		&result.ID,
		&result.ColumnID,
		&result.Title,
		&result.Description,
		&result.Position,
		&result.Priority,
		&result.DueDate,
		&result.AssignedTo,
		&result.CreatedBy,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to create task", err)
	}

	return &result, nil
}

// GetByID retrieves a task by its ID.
func (r *TaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	query := `
		SELECT id, column_id, title, description, position, priority, due_date, assigned_to, created_by, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	var task models.Task
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.ColumnID,
		&task.Title,
		&task.Description,
		&task.Position,
		&task.Priority,
		&task.DueDate,
		&task.AssignedTo,
		&task.CreatedBy,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewNotFoundError("task")
		}
		return nil, apperrors.NewDatabaseError("failed to get task", err)
	}

	return &task, nil
}

// GetByColumnID retrieves all tasks for a column.
func (r *TaskRepository) GetByColumnID(ctx context.Context, columnID uuid.UUID) ([]models.Task, error) {
	query := `
		SELECT id, column_id, title, description, position, priority, due_date, assigned_to, created_by, created_at, updated_at
		FROM tasks
		WHERE column_id = $1
		ORDER BY position ASC
	`

	rows, err := r.pool.Query(ctx, query, columnID)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to get tasks", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(
			&task.ID,
			&task.ColumnID,
			&task.Title,
			&task.Description,
			&task.Position,
			&task.Priority,
			&task.DueDate,
			&task.AssignedTo,
			&task.CreatedBy,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, apperrors.NewDatabaseError("failed to scan task", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.NewDatabaseError("error iterating tasks", err)
	}

	return tasks, nil
}

// Update updates a task's information.
func (r *TaskRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Task, error) {
	// Build dynamic update query
	query := `
		UPDATE tasks
		SET title = COALESCE($2, title),
			description = COALESCE($3, description),
			priority = COALESCE($4, priority),
			due_date = COALESCE($5, due_date),
			assigned_to = COALESCE($6, assigned_to),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, column_id, title, description, position, priority, due_date, assigned_to, created_by, created_at, updated_at
	`

	var task models.Task
	err := r.pool.QueryRow(ctx, query, id,
		updates["title"],
		updates["description"],
		updates["priority"],
		updates["due_date"],
		updates["assigned_to"],
	).Scan(
		&task.ID,
		&task.ColumnID,
		&task.Title,
		&task.Description,
		&task.Position,
		&task.Priority,
		&task.DueDate,
		&task.AssignedTo,
		&task.CreatedBy,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewNotFoundError("task")
		}
		return nil, apperrors.NewDatabaseError("failed to update task", err)
	}

	return &task, nil
}

// Move moves a task to a new column and/or position.
func (r *TaskRepository) Move(ctx context.Context, id uuid.UUID, newColumnID uuid.UUID, newPosition int) (*models.Task, error) {
	// Get current task
	task, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldColumnID := task.ColumnID
	oldPosition := task.Position

	// If moving within the same column
	if oldColumnID == newColumnID {
		if oldPosition == newPosition {
			return task, nil
		}

		if newPosition > oldPosition {
			// Moving down
			shiftQuery := `
				UPDATE tasks
				SET position = position - 1
				WHERE column_id = $1 AND position > $2 AND position <= $3
			`
			_, err = r.pool.Exec(ctx, shiftQuery, oldColumnID, oldPosition, newPosition)
		} else {
			// Moving up
			shiftQuery := `
				UPDATE tasks
				SET position = position + 1
				WHERE column_id = $1 AND position >= $2 AND position < $3
			`
			_, err = r.pool.Exec(ctx, shiftQuery, oldColumnID, newPosition, oldPosition)
		}
		if err != nil {
			return nil, apperrors.NewDatabaseError("failed to shift tasks", err)
		}
	} else {
		// Moving to a different column
		// First, remove from old column
		shiftOldQuery := `
			UPDATE tasks
			SET position = position - 1
			WHERE column_id = $1 AND position > $2
		`
		_, err = r.pool.Exec(ctx, shiftOldQuery, oldColumnID, oldPosition)
		if err != nil {
			return nil, apperrors.NewDatabaseError("failed to shift old column tasks", err)
		}

		// Then, make room in new column
		shiftNewQuery := `
			UPDATE tasks
			SET position = position + 1
			WHERE column_id = $1 AND position >= $2
		`
		_, err = r.pool.Exec(ctx, shiftNewQuery, newColumnID, newPosition)
		if err != nil {
			return nil, apperrors.NewDatabaseError("failed to shift new column tasks", err)
		}
	}

	// Update the task
	updateQuery := `
		UPDATE tasks
		SET column_id = $2, position = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, column_id, title, description, position, priority, due_date, assigned_to, created_by, created_at, updated_at
	`

	err = r.pool.QueryRow(ctx, updateQuery, id, newColumnID, newPosition).Scan(
		&task.ID,
		&task.ColumnID,
		&task.Title,
		&task.Description,
		&task.Position,
		&task.Priority,
		&task.DueDate,
		&task.AssignedTo,
		&task.CreatedBy,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to update task position", err)
	}

	return task, nil
}

// Delete deletes a task by ID.
func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Get the task first
	task, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete the task
	query := `DELETE FROM tasks WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return apperrors.NewDatabaseError("failed to delete task", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NewNotFoundError("task")
	}

	// Shift remaining tasks
	shiftQuery := `
		UPDATE tasks
		SET position = position - 1
		WHERE column_id = $1 AND position > $2
	`
	_, err = r.pool.Exec(ctx, shiftQuery, task.ColumnID, task.Position)
	if err != nil {
		return apperrors.NewDatabaseError("failed to reorder tasks", err)
	}

	return nil
}

// GetMaxPosition returns the maximum position in a column.
func (r *TaskRepository) GetMaxPosition(ctx context.Context, columnID uuid.UUID) (int, error) {
	query := `SELECT COALESCE(MAX(position), -1) FROM tasks WHERE column_id = $1`

	var maxPos int
	err := r.pool.QueryRow(ctx, query, columnID).Scan(&maxPos)
	if err != nil {
		return 0, apperrors.NewDatabaseError("failed to get max position", err)
	}

	return maxPos, nil
}

// GetBoardIDByTaskID retrieves the board ID for a given task.
func (r *TaskRepository) GetBoardIDByTaskID(ctx context.Context, taskID uuid.UUID) (uuid.UUID, error) {
	query := `
		SELECT bc.board_id
		FROM tasks t
		INNER JOIN board_columns bc ON t.column_id = bc.id
		WHERE t.id = $1
	`

	var boardID uuid.UUID
	err := r.pool.QueryRow(ctx, query, taskID).Scan(&boardID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, apperrors.NewNotFoundError("task")
		}
		return uuid.Nil, apperrors.NewDatabaseError("failed to get board ID", err)
	}

	return boardID, nil
}
