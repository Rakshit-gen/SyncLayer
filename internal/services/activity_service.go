package services

import (
	"context"

	"github.com/google/uuid"
	"taskboard-backend/internal/models"
	"taskboard-backend/internal/repositories"
)

// ActivityService handles activity log business logic.
type ActivityService struct {
	activityRepo *repositories.ActivityRepository
}

// NewActivityService creates a new ActivityService instance.
func NewActivityService(activityRepo *repositories.ActivityRepository) *ActivityService {
	return &ActivityService{activityRepo: activityRepo}
}

// GetByTaskID retrieves activity logs for a task.
func (s *ActivityService) GetByTaskID(ctx context.Context, taskID uuid.UUID, limit int) ([]models.ActivityLogWithUser, error) {
	return s.activityRepo.GetByTaskID(ctx, taskID, limit)
}

// GetByBoardID retrieves activity logs for a board.
func (s *ActivityService) GetByBoardID(ctx context.Context, boardID uuid.UUID, limit int) ([]models.ActivityLogWithUser, error) {
	return s.activityRepo.GetByBoardID(ctx, boardID, limit)
}

// Create creates a new activity log entry.
func (s *ActivityService) Create(ctx context.Context, activity *models.ActivityLog) (*models.ActivityLog, error) {
	return s.activityRepo.Create(ctx, activity)
}

// CreateWithDetails creates an activity log with structured details.
func (s *ActivityService) CreateWithDetails(ctx context.Context, taskID, boardID, userID *uuid.UUID, action models.ActivityAction, details interface{}) (*models.ActivityLog, error) {
	return s.activityRepo.CreateWithDetails(ctx, taskID, boardID, userID, action, details)
}
