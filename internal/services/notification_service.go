package services

import (
	"context"

	"github.com/google/uuid"
	"taskboard-backend/internal/models"
	"taskboard-backend/internal/repositories"
)

// NotificationService handles notification business logic.
type NotificationService struct {
	notificationRepo *repositories.NotificationRepository
}

// NewNotificationService creates a new NotificationService instance.
func NewNotificationService(notificationRepo *repositories.NotificationRepository) *NotificationService {
	return &NotificationService{notificationRepo: notificationRepo}
}

// GetByUserID retrieves notifications for a user.
func (s *NotificationService) GetByUserID(ctx context.Context, userID uuid.UUID, unreadOnly bool, limit int) ([]models.Notification, int64, error) {
	notifications, err := s.notificationRepo.GetByUserID(ctx, userID, unreadOnly, limit)
	if err != nil {
		return nil, 0, err
	}

	unreadCount, err := s.notificationRepo.GetUnreadCount(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return notifications, unreadCount, nil
}

// MarkAsRead marks a notification as read.
func (s *NotificationService) MarkAsRead(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	return s.notificationRepo.MarkAsRead(ctx, id)
}

// MarkAllAsRead marks all notifications for a user as read.
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID)
}

// GetUnreadCount returns the count of unread notifications for a user.
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.notificationRepo.GetUnreadCount(ctx, userID)
}

// Create creates a new notification.
func (s *NotificationService) Create(ctx context.Context, notification *models.Notification) (*models.Notification, error) {
	return s.notificationRepo.Create(ctx, notification)
}

// CreateWithData creates a notification with structured data.
func (s *NotificationService) CreateWithData(ctx context.Context, userID uuid.UUID, notifType models.NotificationType, title string, message *string, data interface{}) (*models.Notification, error) {
	return s.notificationRepo.CreateWithData(ctx, userID, notifType, title, message, data)
}

// Delete deletes a notification.
func (s *NotificationService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.notificationRepo.Delete(ctx, id)
}
