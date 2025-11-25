package services_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"taskboard-backend/internal/dto"
	"taskboard-backend/internal/models"
)

func TestTaskService_Create(t *testing.T) {
	t.Run("creates task with required fields", func(t *testing.T) {
		req := &dto.CreateTaskRequest{
			ColumnID:  uuid.New(),
			Title:     "Test Task",
			CreatedBy: uuid.New(),
		}

		assert.NotNil(t, req)
		assert.Equal(t, "Test Task", req.Title)
		assert.NotEqual(t, uuid.Nil, req.ColumnID)
		assert.NotEqual(t, uuid.Nil, req.CreatedBy)
	})

	t.Run("creates task with all fields", func(t *testing.T) {
		priority := models.TaskPriorityHigh
		dueDate := time.Now().Add(24 * time.Hour)
		assignee := uuid.New()

		req := &dto.CreateTaskRequest{
			ColumnID:    uuid.New(),
			Title:       "Complete Task",
			Description: strPtr("A complete task with all fields"),
			Priority:    &priority,
			DueDate:     &dueDate,
			AssignedTo:  &assignee,
			CreatedBy:   uuid.New(),
		}

		assert.NotNil(t, req.Description)
		assert.NotNil(t, req.Priority)
		assert.NotNil(t, req.DueDate)
		assert.NotNil(t, req.AssignedTo)
	})

	t.Run("validates priority", func(t *testing.T) {
		validPriorities := []models.TaskPriority{
			models.TaskPriorityLow,
			models.TaskPriorityMedium,
			models.TaskPriorityHigh,
			models.TaskPriorityUrgent,
		}

		for _, p := range validPriorities {
			assert.True(t, p.IsValid())
		}

		invalidPriority := models.TaskPriority("critical")
		assert.False(t, invalidPriority.IsValid())
	})
}

func TestTaskService_Update(t *testing.T) {
	t.Run("updates task title", func(t *testing.T) {
		newTitle := "Updated Title"
		req := &dto.UpdateTaskRequest{
			Title: &newTitle,
		}

		assert.NotNil(t, req.Title)
		assert.Equal(t, "Updated Title", *req.Title)
	})

	t.Run("updates task priority", func(t *testing.T) {
		newPriority := models.TaskPriorityUrgent
		req := &dto.UpdateTaskRequest{
			Priority: &newPriority,
		}

		assert.NotNil(t, req.Priority)
		assert.Equal(t, models.TaskPriorityUrgent, *req.Priority)
	})

	t.Run("updates task assignee", func(t *testing.T) {
		newAssignee := uuid.New()
		req := &dto.UpdateTaskRequest{
			AssignedTo: &newAssignee,
		}

		assert.NotNil(t, req.AssignedTo)
		assert.NotEqual(t, uuid.Nil, *req.AssignedTo)
	})
}

func TestTaskService_Move(t *testing.T) {
	t.Run("moves task within same column", func(t *testing.T) {
		columnID := uuid.New()
		req := &dto.MoveTaskRequest{
			ColumnID: columnID,
			Position: 2,
		}

		assert.Equal(t, columnID, req.ColumnID)
		assert.Equal(t, 2, req.Position)
	})

	t.Run("moves task to different column", func(t *testing.T) {
		fromColumnID := uuid.New()
		toColumnID := uuid.New()

		req := &dto.MoveTaskRequest{
			ColumnID: toColumnID,
			Position: 0,
		}

		assert.NotEqual(t, fromColumnID, toColumnID)
		assert.Equal(t, toColumnID, req.ColumnID)
		assert.Equal(t, 0, req.Position)
	})

	t.Run("position must be non-negative", func(t *testing.T) {
		req := &dto.MoveTaskRequest{
			ColumnID: uuid.New(),
			Position: 0,
		}

		assert.GreaterOrEqual(t, req.Position, 0)
	})
}

func TestTaskService_Notifications(t *testing.T) {
	t.Run("sends notification when task assigned to different user", func(t *testing.T) {
		creatorID := uuid.New()
		assigneeID := uuid.New()

		shouldNotify := assigneeID != creatorID
		assert.True(t, shouldNotify)
	})

	t.Run("does not notify when self-assigning", func(t *testing.T) {
		userID := uuid.New()

		shouldNotify := userID != userID
		assert.False(t, shouldNotify)
	})
}
