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

// ColumnRepository handles column data access operations.
type ColumnRepository struct {
	pool *pgxpool.Pool
}

// NewColumnRepository creates a new ColumnRepository instance.
func NewColumnRepository(pool *pgxpool.Pool) *ColumnRepository {
	return &ColumnRepository{pool: pool}
}

// Create creates a new column.
func (r *ColumnRepository) Create(ctx context.Context, column *models.Column) (*models.Column, error) {
	// First, shift existing columns to make room
	shiftQuery := `
		UPDATE board_columns
		SET position = position + 1
		WHERE board_id = $1 AND position >= $2
	`
	_, err := r.pool.Exec(ctx, shiftQuery, column.BoardID, column.Position)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to shift columns", err)
	}

	// Insert the new column
	query := `
		INSERT INTO board_columns (board_id, name, position, color)
		VALUES ($1, $2, $3, $4)
		RETURNING id, board_id, name, position, color, created_at, updated_at
	`

	var result models.Column
	err = r.pool.QueryRow(ctx, query, column.BoardID, column.Name, column.Position, column.Color).Scan(
		&result.ID,
		&result.BoardID,
		&result.Name,
		&result.Position,
		&result.Color,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to create column", err)
	}

	return &result, nil
}

// GetByID retrieves a column by its ID.
func (r *ColumnRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Column, error) {
	query := `
		SELECT id, board_id, name, position, color, created_at, updated_at
		FROM board_columns
		WHERE id = $1
	`

	var column models.Column
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&column.ID,
		&column.BoardID,
		&column.Name,
		&column.Position,
		&column.Color,
		&column.CreatedAt,
		&column.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewNotFoundError("column")
		}
		return nil, apperrors.NewDatabaseError("failed to get column", err)
	}

	return &column, nil
}

// GetByIDWithTasks retrieves a column with its tasks.
func (r *ColumnRepository) GetByIDWithTasks(ctx context.Context, id uuid.UUID) (*models.ColumnWithTasks, error) {
	column, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tasksQuery := `
		SELECT id, column_id, title, description, position, priority, due_date, assigned_to, created_by, created_at, updated_at
		FROM tasks
		WHERE column_id = $1
		ORDER BY position ASC
	`

	rows, err := r.pool.Query(ctx, tasksQuery, id)
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

	return &models.ColumnWithTasks{
		Column: *column,
		Tasks:  tasks,
	}, nil
}

// GetByBoardID retrieves all columns for a board.
func (r *ColumnRepository) GetByBoardID(ctx context.Context, boardID uuid.UUID) ([]models.Column, error) {
	query := `
		SELECT id, board_id, name, position, color, created_at, updated_at
		FROM board_columns
		WHERE board_id = $1
		ORDER BY position ASC
	`

	rows, err := r.pool.Query(ctx, query, boardID)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to get columns", err)
	}
	defer rows.Close()

	var columns []models.Column
	for rows.Next() {
		var column models.Column
		if err := rows.Scan(
			&column.ID,
			&column.BoardID,
			&column.Name,
			&column.Position,
			&column.Color,
			&column.CreatedAt,
			&column.UpdatedAt,
		); err != nil {
			return nil, apperrors.NewDatabaseError("failed to scan column", err)
		}
		columns = append(columns, column)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.NewDatabaseError("error iterating columns", err)
	}

	return columns, nil
}

// Update updates a column's information.
func (r *ColumnRepository) Update(ctx context.Context, id uuid.UUID, name *string, color *string) (*models.Column, error) {
	query := `
		UPDATE board_columns
		SET name = COALESCE($2, name),
			color = COALESCE($3, color),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, board_id, name, position, color, created_at, updated_at
	`

	var column models.Column
	err := r.pool.QueryRow(ctx, query, id, name, color).Scan(
		&column.ID,
		&column.BoardID,
		&column.Name,
		&column.Position,
		&column.Color,
		&column.CreatedAt,
		&column.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewNotFoundError("column")
		}
		return nil, apperrors.NewDatabaseError("failed to update column", err)
	}

	return &column, nil
}

// Move moves a column to a new position.
func (r *ColumnRepository) Move(ctx context.Context, id uuid.UUID, newPosition int) (*models.Column, error) {
	// Get current column
	column, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldPosition := column.Position

	if oldPosition == newPosition {
		return column, nil
	}

	// Reorder columns
	if newPosition > oldPosition {
		// Moving down: shift items up
		shiftQuery := `
			UPDATE board_columns
			SET position = position - 1
			WHERE board_id = $1 AND position > $2 AND position <= $3
		`
		_, err = r.pool.Exec(ctx, shiftQuery, column.BoardID, oldPosition, newPosition)
	} else {
		// Moving up: shift items down
		shiftQuery := `
			UPDATE board_columns
			SET position = position + 1
			WHERE board_id = $1 AND position >= $2 AND position < $3
		`
		_, err = r.pool.Exec(ctx, shiftQuery, column.BoardID, newPosition, oldPosition)
	}
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to shift columns", err)
	}

	// Update the column's position
	updateQuery := `
		UPDATE board_columns
		SET position = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, board_id, name, position, color, created_at, updated_at
	`

	err = r.pool.QueryRow(ctx, updateQuery, id, newPosition).Scan(
		&column.ID,
		&column.BoardID,
		&column.Name,
		&column.Position,
		&column.Color,
		&column.CreatedAt,
		&column.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to update column position", err)
	}

	return column, nil
}

// Delete deletes a column by ID.
func (r *ColumnRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Get the column first to know its position
	column, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete the column
	query := `DELETE FROM board_columns WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return apperrors.NewDatabaseError("failed to delete column", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NewNotFoundError("column")
	}

	// Shift remaining columns
	shiftQuery := `
		UPDATE board_columns
		SET position = position - 1
		WHERE board_id = $1 AND position > $2
	`
	_, err = r.pool.Exec(ctx, shiftQuery, column.BoardID, column.Position)
	if err != nil {
		return apperrors.NewDatabaseError("failed to reorder columns", err)
	}

	return nil
}

// GetMaxPosition returns the maximum position in a board.
func (r *ColumnRepository) GetMaxPosition(ctx context.Context, boardID uuid.UUID) (int, error) {
	query := `SELECT COALESCE(MAX(position), -1) FROM board_columns WHERE board_id = $1`

	var maxPos int
	err := r.pool.QueryRow(ctx, query, boardID).Scan(&maxPos)
	if err != nil {
		return 0, apperrors.NewDatabaseError("failed to get max position", err)
	}

	return maxPos, nil
}
