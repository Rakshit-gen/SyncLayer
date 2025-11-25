package repositories

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"taskboard-backend/internal/models"
	apperrors "taskboard-backend/pkg/errors"
)

// ActivityRepository handles activity log data access operations.
type ActivityRepository struct {
	pool *pgxpool.Pool
}

// NewActivityRepository creates a new ActivityRepository instance.
func NewActivityRepository(pool *pgxpool.Pool) *ActivityRepository {
	return &ActivityRepository{pool: pool}
}

// Create creates a new activity log entry.
func (r *ActivityRepository) Create(ctx context.Context, activity *models.ActivityLog) (*models.ActivityLog, error) {
	query := `
		INSERT INTO activity_logs (task_id, board_id, user_id, action, details)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, task_id, board_id, user_id, action, details, created_at
	`

	var result models.ActivityLog
	err := r.pool.QueryRow(ctx, query,
		activity.TaskID,
		activity.BoardID,
		activity.UserID,
		activity.Action,
		activity.Details,
	).Scan(
		&result.ID,
		&result.TaskID,
		&result.BoardID,
		&result.UserID,
		&result.Action,
		&result.Details,
		&result.CreatedAt,
	)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to create activity log", err)
	}

	return &result, nil
}

// GetByTaskID retrieves activity logs for a task.
func (r *ActivityRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID, limit int) ([]models.ActivityLogWithUser, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT al.id, al.task_id, al.board_id, al.user_id, al.action, al.details, al.created_at,
			   u.id, u.email, u.name, u.avatar_url, u.created_at, u.updated_at
		FROM activity_logs al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE al.task_id = $1
		ORDER BY al.created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, taskID, limit)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to get activity logs", err)
	}
	defer rows.Close()

	var activities []models.ActivityLogWithUser
	for rows.Next() {
		var activity models.ActivityLogWithUser
		var user models.User
		var userID, userEmail, userName *string
		var userAvatar *string
		var userCreatedAt, userUpdatedAt *interface{}

		if err := rows.Scan(
			&activity.ID,
			&activity.TaskID,
			&activity.BoardID,
			&activity.UserID,
			&activity.Action,
			&activity.Details,
			&activity.CreatedAt,
			&userID,
			&userEmail,
			&userName,
			&userAvatar,
			&userCreatedAt,
			&userUpdatedAt,
		); err != nil {
			return nil, apperrors.NewDatabaseError("failed to scan activity log", err)
		}

		if userID != nil {
			user.Email = *userEmail
			user.Name = *userName
			user.AvatarURL = userAvatar
			activity.User = &user
		}

		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.NewDatabaseError("error iterating activity logs", err)
	}

	return activities, nil
}

// GetByBoardID retrieves activity logs for a board.
func (r *ActivityRepository) GetByBoardID(ctx context.Context, boardID uuid.UUID, limit int) ([]models.ActivityLogWithUser, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT al.id, al.task_id, al.board_id, al.user_id, al.action, al.details, al.created_at,
			   u.id, u.email, u.name, u.avatar_url, u.created_at, u.updated_at
		FROM activity_logs al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE al.board_id = $1
		ORDER BY al.created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, boardID, limit)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to get activity logs", err)
	}
	defer rows.Close()

	var activities []models.ActivityLogWithUser
	for rows.Next() {
		var activity models.ActivityLogWithUser
		var user models.User
		var userID *uuid.UUID
		var userEmail, userName *string
		var userAvatar *string
		var userCreatedAt, userUpdatedAt *interface{}

		if err := rows.Scan(
			&activity.ID,
			&activity.TaskID,
			&activity.BoardID,
			&activity.UserID,
			&activity.Action,
			&activity.Details,
			&activity.CreatedAt,
			&userID,
			&userEmail,
			&userName,
			&userAvatar,
			&userCreatedAt,
			&userUpdatedAt,
		); err != nil {
			return nil, apperrors.NewDatabaseError("failed to scan activity log", err)
		}

		if userID != nil {
			user.ID = *userID
			user.Email = *userEmail
			user.Name = *userName
			user.AvatarURL = userAvatar
			activity.User = &user
		}

		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.NewDatabaseError("error iterating activity logs", err)
	}

	return activities, nil
}

// CreateWithDetails is a helper function to create an activity with structured details.
func (r *ActivityRepository) CreateWithDetails(ctx context.Context, taskID, boardID, userID *uuid.UUID, action models.ActivityAction, details interface{}) (*models.ActivityLog, error) {
	var detailsJSON json.RawMessage
	if details != nil {
		bytes, err := json.Marshal(details)
		if err != nil {
			return nil, apperrors.NewInternalError("failed to marshal activity details", err)
		}
		detailsJSON = bytes
	}

	activity := &models.ActivityLog{
		TaskID:  taskID,
		BoardID: boardID,
		UserID:  userID,
		Action:  action,
		Details: detailsJSON,
	}

	return r.Create(ctx, activity)
}
