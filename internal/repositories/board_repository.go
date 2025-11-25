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

// BoardRepository handles board data access operations.
type BoardRepository struct {
	pool *pgxpool.Pool
}

// NewBoardRepository creates a new BoardRepository instance.
func NewBoardRepository(pool *pgxpool.Pool) *BoardRepository {
	return &BoardRepository{pool: pool}
}

// Create creates a new board.
func (r *BoardRepository) Create(ctx context.Context, board *models.Board) (*models.Board, error) {
	query := `
		INSERT INTO boards (team_id, name, description, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, team_id, name, description, created_by, created_at, updated_at
	`

	var result models.Board
	err := r.pool.QueryRow(ctx, query, board.TeamID, board.Name, board.Description, board.CreatedBy).Scan(
		&result.ID,
		&result.TeamID,
		&result.Name,
		&result.Description,
		&result.CreatedBy,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to create board", err)
	}

	return &result, nil
}

// GetByID retrieves a board by its ID.
func (r *BoardRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Board, error) {
	query := `
		SELECT id, team_id, name, description, created_by, created_at, updated_at
		FROM boards
		WHERE id = $1
	`

	var board models.Board
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&board.ID,
		&board.TeamID,
		&board.Name,
		&board.Description,
		&board.CreatedBy,
		&board.CreatedAt,
		&board.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewNotFoundError("board")
		}
		return nil, apperrors.NewDatabaseError("failed to get board", err)
	}

	return &board, nil
}

// GetByIDFull retrieves a board with all its columns and tasks.
func (r *BoardRepository) GetByIDFull(ctx context.Context, id uuid.UUID) (*models.BoardFull, error) {
	// First get the board
	board, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get columns with tasks
	columnsQuery := `
		SELECT id, board_id, name, position, color, created_at, updated_at
		FROM board_columns
		WHERE board_id = $1
		ORDER BY position ASC
	`

	columnRows, err := r.pool.Query(ctx, columnsQuery, id)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to get columns", err)
	}
	defer columnRows.Close()

	var columns []models.ColumnWithTasks
	columnIDs := make([]uuid.UUID, 0)

	for columnRows.Next() {
		var col models.ColumnWithTasks
		if err := columnRows.Scan(
			&col.ID,
			&col.BoardID,
			&col.Name,
			&col.Position,
			&col.Color,
			&col.CreatedAt,
			&col.UpdatedAt,
		); err != nil {
			return nil, apperrors.NewDatabaseError("failed to scan column", err)
		}
		col.Tasks = []models.Task{} // Initialize empty slice
		columns = append(columns, col)
		columnIDs = append(columnIDs, col.ID)
	}

	if err := columnRows.Err(); err != nil {
		return nil, apperrors.NewDatabaseError("error iterating columns", err)
	}

	// Get all tasks for these columns
	if len(columnIDs) > 0 {
		tasksQuery := `
			SELECT id, column_id, title, description, position, priority, due_date, assigned_to, created_by, created_at, updated_at
			FROM tasks
			WHERE column_id = ANY($1)
			ORDER BY position ASC
		`

		taskRows, err := r.pool.Query(ctx, tasksQuery, columnIDs)
		if err != nil {
			return nil, apperrors.NewDatabaseError("failed to get tasks", err)
		}
		defer taskRows.Close()

		// Map tasks to their columns
		tasksByColumn := make(map[uuid.UUID][]models.Task)
		for taskRows.Next() {
			var task models.Task
			if err := taskRows.Scan(
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
			tasksByColumn[task.ColumnID] = append(tasksByColumn[task.ColumnID], task)
		}

		if err := taskRows.Err(); err != nil {
			return nil, apperrors.NewDatabaseError("error iterating tasks", err)
		}

		// Assign tasks to columns
		for i := range columns {
			if tasks, ok := tasksByColumn[columns[i].ID]; ok {
				columns[i].Tasks = tasks
			}
		}
	}

	return &models.BoardFull{
		Board:   *board,
		Columns: columns,
	}, nil
}

// GetByTeamID retrieves all boards for a team.
func (r *BoardRepository) GetByTeamID(ctx context.Context, teamID uuid.UUID) ([]models.Board, error) {
	query := `
		SELECT id, team_id, name, description, created_by, created_at, updated_at
		FROM boards
		WHERE team_id = $1
		ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to get team boards", err)
	}
	defer rows.Close()

	var boards []models.Board
	for rows.Next() {
		var board models.Board
		if err := rows.Scan(
			&board.ID,
			&board.TeamID,
			&board.Name,
			&board.Description,
			&board.CreatedBy,
			&board.CreatedAt,
			&board.UpdatedAt,
		); err != nil {
			return nil, apperrors.NewDatabaseError("failed to scan board", err)
		}
		boards = append(boards, board)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.NewDatabaseError("error iterating boards", err)
	}

	return boards, nil
}

// Update updates a board's information.
func (r *BoardRepository) Update(ctx context.Context, id uuid.UUID, name *string, description *string) (*models.Board, error) {
	query := `
		UPDATE boards
		SET name = COALESCE($2, name),
			description = COALESCE($3, description),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, team_id, name, description, created_by, created_at, updated_at
	`

	var board models.Board
	err := r.pool.QueryRow(ctx, query, id, name, description).Scan(
		&board.ID,
		&board.TeamID,
		&board.Name,
		&board.Description,
		&board.CreatedBy,
		&board.CreatedAt,
		&board.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewNotFoundError("board")
		}
		return nil, apperrors.NewDatabaseError("failed to update board", err)
	}

	return &board, nil
}

// Delete deletes a board by ID.
func (r *BoardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM boards WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return apperrors.NewDatabaseError("failed to delete board", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NewNotFoundError("board")
	}

	return nil
}
