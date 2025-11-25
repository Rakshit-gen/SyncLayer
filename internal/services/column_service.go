package services

import (
	"context"

	"github.com/google/uuid"
	"taskboard-backend/internal/dto"
	"taskboard-backend/internal/models"
	"taskboard-backend/internal/repositories"
)

// ColumnService handles column business logic.
type ColumnService struct {
	columnRepo   *repositories.ColumnRepository
	activityRepo *repositories.ActivityRepository
}

// NewColumnService creates a new ColumnService instance.
func NewColumnService(
	columnRepo *repositories.ColumnRepository,
	activityRepo *repositories.ActivityRepository,
) *ColumnService {
	return &ColumnService{
		columnRepo:   columnRepo,
		activityRepo: activityRepo,
	}
}

// Create creates a new column.
func (s *ColumnService) Create(ctx context.Context, req *dto.CreateColumnRequest, userID uuid.UUID) (*models.Column, error) {
	color := "#6366f1" // Default color
	if req.Color != nil {
		color = *req.Color
	}

	// If position is not specified, add to the end
	position := req.Position
	if position == 0 {
		maxPos, err := s.columnRepo.GetMaxPosition(ctx, req.BoardID)
		if err == nil {
			position = maxPos + 1
		}
	}

	column := &models.Column{
		BoardID:  req.BoardID,
		Name:     req.Name,
		Position: position,
		Color:    color,
	}

	createdColumn, err := s.columnRepo.Create(ctx, column)
	if err != nil {
		return nil, err
	}

	// Log activity
	s.activityRepo.CreateWithDetails(ctx, nil, &req.BoardID, &userID, models.ActivityColumnCreated, map[string]interface{}{
		"column_name": createdColumn.Name,
		"column_id":   createdColumn.ID,
	})

	return createdColumn, nil
}

// GetByID retrieves a column by its ID.
func (s *ColumnService) GetByID(ctx context.Context, id uuid.UUID) (*models.Column, error) {
	return s.columnRepo.GetByID(ctx, id)
}

// GetByIDWithTasks retrieves a column with its tasks.
func (s *ColumnService) GetByIDWithTasks(ctx context.Context, id uuid.UUID) (*models.ColumnWithTasks, error) {
	return s.columnRepo.GetByIDWithTasks(ctx, id)
}

// GetByBoardID retrieves all columns for a board.
func (s *ColumnService) GetByBoardID(ctx context.Context, boardID uuid.UUID) ([]models.Column, error) {
	return s.columnRepo.GetByBoardID(ctx, boardID)
}

// Update updates a column's information.
func (s *ColumnService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateColumnRequest, userID uuid.UUID) (*models.Column, error) {
	// Get current column for activity logging
	currentColumn, err := s.columnRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updatedColumn, err := s.columnRepo.Update(ctx, id, req.Name, req.Color)
	if err != nil {
		return nil, err
	}

	// Log activity
	details := map[string]interface{}{"column_id": id}
	if req.Name != nil && *req.Name != currentColumn.Name {
		details["previous_name"] = currentColumn.Name
		details["new_name"] = *req.Name
	}
	if len(details) > 1 {
		s.activityRepo.CreateWithDetails(ctx, nil, &currentColumn.BoardID, &userID, models.ActivityColumnUpdated, details)
	}

	return updatedColumn, nil
}

// Move moves a column to a new position.
func (s *ColumnService) Move(ctx context.Context, id uuid.UUID, req *dto.MoveColumnRequest, userID uuid.UUID) (*models.Column, error) {
	// Get current column for activity logging
	currentColumn, err := s.columnRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	movedColumn, err := s.columnRepo.Move(ctx, id, req.Position)
	if err != nil {
		return nil, err
	}

	// Log activity
	if currentColumn.Position != req.Position {
		s.activityRepo.CreateWithDetails(ctx, nil, &currentColumn.BoardID, &userID, models.ActivityColumnMoved, map[string]interface{}{
			"column_id":         id,
			"previous_position": currentColumn.Position,
			"new_position":      req.Position,
		})
	}

	return movedColumn, nil
}

// Delete deletes a column.
func (s *ColumnService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get column for activity logging
	column, err := s.columnRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	err = s.columnRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Log activity
	s.activityRepo.CreateWithDetails(ctx, nil, &column.BoardID, &userID, models.ActivityColumnDeleted, map[string]interface{}{
		"column_name": column.Name,
	})

	return nil
}
