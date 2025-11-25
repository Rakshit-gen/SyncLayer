package repositories

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"taskboard-backend/internal/models"
	apperrors "taskboard-backend/pkg/errors"
)

// NotificationRepository handles notification data access operations.
type NotificationRepository struct {
	pool *pgxpool.Pool
}

// NewNotificationRepository creates a new NotificationRepository instance.
func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

// Create creates a new notification.
func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) (*models.Notification, error) {
	query := `
		INSERT INTO notifications (user_id, type, title, message, data)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, type, title, message, data, read, created_at
	`

	var result models.Notification
	err := r.pool.QueryRow(ctx, query,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.Data,
	).Scan(
		&result.ID,
		&result.UserID,
		&result.Type,
		&result.Title,
		&result.Message,
		&result.Data,
		&result.Read,
		&result.CreatedAt,
	)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to create notification", err)
	}

	return &result, nil
}

// GetByID retrieves a notification by its ID.
func (r *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data, read, created_at
		FROM notifications
		WHERE id = $1
	`

	var notification models.Notification
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&notification.Data,
		&notification.Read,
		&notification.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewNotFoundError("notification")
		}
		return nil, apperrors.NewDatabaseError("failed to get notification", err)
	}

	return &notification, nil
}

// GetByUserID retrieves notifications for a user.
func (r *NotificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, unreadOnly bool, limit int) ([]models.Notification, error) {
	if limit <= 0 {
		limit = 50
	}

	var query string
	var args []interface{}

	if unreadOnly {
		query = `
			SELECT id, user_id, type, title, message, data, read, created_at
			FROM notifications
			WHERE user_id = $1 AND read = false
			ORDER BY created_at DESC
			LIMIT $2
		`
		args = []interface{}{userID, limit}
	} else {
		query = `
			SELECT id, user_id, type, title, message, data, read, created_at
			FROM notifications
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		args = []interface{}{userID, limit}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to get notifications", err)
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var notification models.Notification
		if err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&notification.Data,
			&notification.Read,
			&notification.CreatedAt,
		); err != nil {
			return nil, apperrors.NewDatabaseError("failed to scan notification", err)
		}
		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.NewDatabaseError("error iterating notifications", err)
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read.
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	query := `
		UPDATE notifications
		SET read = true
		WHERE id = $1
		RETURNING id, user_id, type, title, message, data, read, created_at
	`

	var notification models.Notification
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&notification.Data,
		&notification.Read,
		&notification.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewNotFoundError("notification")
		}
		return nil, apperrors.NewDatabaseError("failed to mark notification as read", err)
	}

	return &notification, nil
}

// MarkAllAsRead marks all notifications for a user as read.
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE notifications SET read = true WHERE user_id = $1 AND read = false`

	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return apperrors.NewDatabaseError("failed to mark all notifications as read", err)
	}

	return nil
}

// GetUnreadCount returns the count of unread notifications for a user.
func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = false`

	var count int64
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, apperrors.NewDatabaseError("failed to get unread count", err)
	}

	return count, nil
}

// Delete deletes a notification by ID.
func (r *NotificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return apperrors.NewDatabaseError("failed to delete notification", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NewNotFoundError("notification")
	}

	return nil
}

// CreateWithData is a helper function to create a notification with structured data.
func (r *NotificationRepository) CreateWithData(ctx context.Context, userID uuid.UUID, notifType models.NotificationType, title string, message *string, data interface{}) (*models.Notification, error) {
	var dataJSON json.RawMessage
	if data != nil {
		bytes, err := json.Marshal(data)
		if err != nil {
			return nil, apperrors.NewInternalError("failed to marshal notification data", err)
		}
		dataJSON = bytes
	}

	notification := &models.Notification{
		UserID:  userID,
		Type:    notifType,
		Title:   title,
		Message: message,
		Data:    dataJSON,
	}

	return r.Create(ctx, notification)
}
