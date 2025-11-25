package services

import (
	"context"

	"github.com/google/uuid"
	"taskboard-backend/internal/dto"
	"taskboard-backend/internal/models"
	"taskboard-backend/internal/repositories"
)

// TaskService handles task business logic.
type TaskService struct {
	taskRepo         *repositories.TaskRepository
	activityRepo     *repositories.ActivityRepository
	notificationRepo *repositories.NotificationRepository
}

// NewTaskService creates a new TaskService instance.
func NewTaskService(
	taskRepo *repositories.TaskRepository,
	activityRepo *repositories.ActivityRepository,
	notificationRepo *repositories.NotificationRepository,
) *TaskService {
	return &TaskService{
		taskRepo:         taskRepo,
		activityRepo:     activityRepo,
		notificationRepo: notificationRepo,
	}
}

// Create creates a new task.
func (s *TaskService) Create(ctx context.Context, req *dto.CreateTaskRequest) (*models.Task, error) {
	priority := models.TaskPriorityMedium
	if req.Priority != nil && req.Priority.IsValid() {
		priority = *req.Priority
	}

	// If position is not specified, add to the end
	position := 0
	maxPos, err := s.taskRepo.GetMaxPosition(ctx, req.ColumnID)
	if err == nil {
		position = maxPos + 1
	}

	task := &models.Task{
		ColumnID:    req.ColumnID,
		Title:       req.Title,
		Description: req.Description,
		Position:    position,
		Priority:    priority,
		DueDate:     req.DueDate,
		AssignedTo:  req.AssignedTo,
		CreatedBy:   &req.CreatedBy,
	}

	createdTask, err := s.taskRepo.Create(ctx, task)
	if err != nil {
		return nil, err
	}

	// Get board ID for activity logging
	boardID, _ := s.taskRepo.GetBoardIDByTaskID(ctx, createdTask.ID)

	// Log activity
	s.activityRepo.CreateWithDetails(ctx, &createdTask.ID, &boardID, &req.CreatedBy, models.ActivityTaskCreated, map[string]interface{}{
		"task_title": createdTask.Title,
	})

	// Send notification if assigned to someone
	if req.AssignedTo != nil && *req.AssignedTo != req.CreatedBy {
		message := "You have been assigned a new task: " + createdTask.Title
		s.notificationRepo.CreateWithData(ctx, *req.AssignedTo, models.NotificationTaskAssigned, "New Task Assigned", &message, models.NotificationData{
			TaskID:  &createdTask.ID,
			BoardID: &boardID,
		})
	}

	return createdTask, nil
}

// GetByID retrieves a task by its ID.
func (s *TaskService) GetByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	return s.taskRepo.GetByID(ctx, id)
}

// GetByColumnID retrieves all tasks for a column.
func (s *TaskService) GetByColumnID(ctx context.Context, columnID uuid.UUID) ([]models.Task, error) {
	return s.taskRepo.GetByColumnID(ctx, columnID)
}

// Update updates a task's information.
func (s *TaskService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateTaskRequest, userID uuid.UUID) (*models.Task, error) {
	// Get current task for activity logging and notifications
	currentTask, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	activityDetails := map[string]interface{}{}

	if req.Title != nil {
		updates["title"] = *req.Title
		if *req.Title != currentTask.Title {
			activityDetails["previous_title"] = currentTask.Title
			activityDetails["new_title"] = *req.Title
		}
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
		if *req.Priority != currentTask.Priority {
			activityDetails["previous_priority"] = currentTask.Priority
			activityDetails["new_priority"] = *req.Priority
		}
	}
	if req.DueDate != nil {
		updates["due_date"] = *req.DueDate
	}
	if req.AssignedTo != nil {
		updates["assigned_to"] = *req.AssignedTo
		// Check if assignee changed
		if currentTask.AssignedTo == nil || *req.AssignedTo != *currentTask.AssignedTo {
			activityDetails["previous_assignee"] = currentTask.AssignedTo
			activityDetails["new_assignee"] = *req.AssignedTo
		}
	}

	updatedTask, err := s.taskRepo.Update(ctx, id, updates)
	if err != nil {
		return nil, err
	}

	// Get board ID for activity logging
	boardID, _ := s.taskRepo.GetBoardIDByTaskID(ctx, id)

	// Log activity if there are changes
	if len(activityDetails) > 0 {
		s.activityRepo.CreateWithDetails(ctx, &id, &boardID, &userID, models.ActivityTaskUpdated, activityDetails)
	}

	// Send notification if assigned to someone new
	if req.AssignedTo != nil {
		isNewAssignment := currentTask.AssignedTo == nil || *req.AssignedTo != *currentTask.AssignedTo
		if isNewAssignment && *req.AssignedTo != userID {
			message := "You have been assigned to task: " + updatedTask.Title
			s.notificationRepo.CreateWithData(ctx, *req.AssignedTo, models.NotificationTaskAssigned, "Task Assigned", &message, models.NotificationData{
				TaskID:  &id,
				BoardID: &boardID,
				ActorID: &userID,
			})
		}
	}

	return updatedTask, nil
}

// Move moves a task to a new column and/or position.
func (s *TaskService) Move(ctx context.Context, id uuid.UUID, req *dto.MoveTaskRequest, userID uuid.UUID) (*models.Task, error) {
	// Get current task for activity logging
	currentTask, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	fromColumnID := currentTask.ColumnID

	movedTask, err := s.taskRepo.Move(ctx, id, req.ColumnID, req.Position)
	if err != nil {
		return nil, err
	}

	// Get board ID for activity logging
	boardID, _ := s.taskRepo.GetBoardIDByTaskID(ctx, id)

	// Log activity
	s.activityRepo.CreateWithDetails(ctx, &id, &boardID, &userID, models.ActivityTaskMoved, map[string]interface{}{
		"from_column_id": fromColumnID,
		"to_column_id":   req.ColumnID,
		"new_position":   req.Position,
	})

	return movedTask, nil
}

// Delete deletes a task.
func (s *TaskService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get task for activity logging
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Get board ID for activity logging
	boardID, _ := s.taskRepo.GetBoardIDByTaskID(ctx, id)

	err = s.taskRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Log activity
	s.activityRepo.CreateWithDetails(ctx, nil, &boardID, &userID, models.ActivityTaskDeleted, map[string]interface{}{
		"task_title": task.Title,
	})

	return nil
}

// GetBoardIDByTaskID retrieves the board ID for a given task.
func (s *TaskService) GetBoardIDByTaskID(ctx context.Context, taskID uuid.UUID) (uuid.UUID, error) {
	return s.taskRepo.GetBoardIDByTaskID(ctx, taskID)
}
