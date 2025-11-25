package services

import (
	"context"

	"github.com/google/uuid"
	"taskboard-backend/internal/dto"
	"taskboard-backend/internal/models"
	"taskboard-backend/internal/repositories"
)

// BoardService handles board business logic.
type BoardService struct {
	boardRepo    *repositories.BoardRepository
	columnRepo   *repositories.ColumnRepository
	activityRepo *repositories.ActivityRepository
}

// NewBoardService creates a new BoardService instance.
func NewBoardService(
	boardRepo *repositories.BoardRepository,
	columnRepo *repositories.ColumnRepository,
	activityRepo *repositories.ActivityRepository,
) *BoardService {
	return &BoardService{
		boardRepo:    boardRepo,
		columnRepo:   columnRepo,
		activityRepo: activityRepo,
	}
}

// Create creates a new board with default columns.
func (s *BoardService) Create(ctx context.Context, req *dto.CreateBoardRequest) (*models.BoardFull, error) {
	// Create the board
	board := &models.Board{
		TeamID:      req.TeamID,
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   &req.CreatedBy,
	}

	createdBoard, err := s.boardRepo.Create(ctx, board)
	if err != nil {
		return nil, err
	}

	// Create default columns
	defaultColumns := []struct {
		Name  string
		Color string
	}{
		{"To Do", "#6b7280"},
		{"In Progress", "#3b82f6"},
		{"Done", "#22c55e"},
	}

	for i, col := range defaultColumns {
		column := &models.Column{
			BoardID:  createdBoard.ID,
			Name:     col.Name,
			Position: i,
			Color:    col.Color,
		}
		_, err := s.columnRepo.Create(ctx, column)
		if err != nil {
			// Log error but continue - board was created
			continue
		}
	}

	// Log activity
	s.activityRepo.CreateWithDetails(ctx, nil, &createdBoard.ID, &req.CreatedBy, models.ActivityBoardCreated, map[string]interface{}{
		"board_name": createdBoard.Name,
	})

	// Return full board
	return s.boardRepo.GetByIDFull(ctx, createdBoard.ID)
}

// GetByID retrieves a board by its ID.
func (s *BoardService) GetByID(ctx context.Context, id uuid.UUID) (*models.Board, error) {
	return s.boardRepo.GetByID(ctx, id)
}

// GetByIDFull retrieves a board with all its columns and tasks.
func (s *BoardService) GetByIDFull(ctx context.Context, id uuid.UUID) (*models.BoardFull, error) {
	return s.boardRepo.GetByIDFull(ctx, id)
}

// GetByTeamID retrieves all boards for a team.
func (s *BoardService) GetByTeamID(ctx context.Context, teamID uuid.UUID) ([]models.Board, error) {
	return s.boardRepo.GetByTeamID(ctx, teamID)
}

// Update updates a board's information.
func (s *BoardService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateBoardRequest, userID uuid.UUID) (*models.Board, error) {
	// Get current board for activity logging
	currentBoard, err := s.boardRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updatedBoard, err := s.boardRepo.Update(ctx, id, req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	// Log activity
	details := map[string]interface{}{}
	if req.Name != nil && *req.Name != currentBoard.Name {
		details["previous_name"] = currentBoard.Name
		details["new_name"] = *req.Name
	}
	if len(details) > 0 {
		s.activityRepo.CreateWithDetails(ctx, nil, &id, &userID, models.ActivityBoardUpdated, details)
	}

	return updatedBoard, nil
}

// Delete deletes a board.
func (s *BoardService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.boardRepo.Delete(ctx, id)
}
