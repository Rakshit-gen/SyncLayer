package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"taskboard-backend/internal/dto"
	"taskboard-backend/internal/services"
	apperrors "taskboard-backend/pkg/errors"
)

// TaskController handles task HTTP endpoints.
type TaskController struct {
	taskService     *services.TaskService
	activityService *services.ActivityService
}

// NewTaskController creates a new TaskController instance.
func NewTaskController(taskService *services.TaskService, activityService *services.ActivityService) *TaskController {
	return &TaskController{
		taskService:     taskService,
		activityService: activityService,
	}
}

// Create handles POST /tasks - creates a new task.
func (c *TaskController) Create(ctx *fiber.Ctx) error {
	var req dto.CreateTaskRequest
	if err := ctx.BodyParser(&req); err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid request body"))
	}

	if req.Title == "" {
		return handleError(ctx, apperrors.NewValidationError("missing required fields", map[string]interface{}{
			"title": "required",
		}))
	}

	if req.ColumnID == uuid.Nil {
		return handleError(ctx, apperrors.NewValidationError("missing required fields", map[string]interface{}{
			"column_id": "required",
		}))
	}

	if req.CreatedBy == uuid.Nil {
		return handleError(ctx, apperrors.NewValidationError("missing required fields", map[string]interface{}{
			"created_by": "required",
		}))
	}

	task, err := c.taskService.Create(ctx.Context(), &req)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(dto.TaskResponse{Task: task})
}

// GetByID handles GET /tasks/:id - retrieves a task.
func (c *TaskController) GetByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid task ID"))
	}

	task, err := c.taskService.GetByID(ctx.Context(), id)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.TaskResponse{Task: task})
}

// Update handles PUT /tasks/:id - updates a task.
func (c *TaskController) Update(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid task ID"))
	}

	var req dto.UpdateTaskRequest
	if err := ctx.BodyParser(&req); err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid request body"))
	}

	// Get user ID from header
	userIDStr := ctx.Get("X-User-ID")
	userID, _ := uuid.Parse(userIDStr)

	task, err := c.taskService.Update(ctx.Context(), id, &req, userID)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.TaskResponse{Task: task})
}

// Move handles PUT /tasks/:id/move - moves a task.
func (c *TaskController) Move(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid task ID"))
	}

	var req dto.MoveTaskRequest
	if err := ctx.BodyParser(&req); err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid request body"))
	}

	if req.ColumnID == uuid.Nil {
		return handleError(ctx, apperrors.NewValidationError("missing required fields", map[string]interface{}{
			"column_id": "required",
		}))
	}

	// Get user ID from header
	userIDStr := ctx.Get("X-User-ID")
	userID, _ := uuid.Parse(userIDStr)

	task, err := c.taskService.Move(ctx.Context(), id, &req, userID)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.TaskResponse{Task: task})
}

// Delete handles DELETE /tasks/:id - deletes a task.
func (c *TaskController) Delete(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid task ID"))
	}

	// Get user ID from header
	userIDStr := ctx.Get("X-User-ID")
	userID, _ := uuid.Parse(userIDStr)

	if err := c.taskService.Delete(ctx.Context(), id, userID); err != nil {
		return handleError(ctx, err)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

// GetActivity handles GET /tasks/:id/activity - retrieves task activity.
func (c *TaskController) GetActivity(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid task ID"))
	}

	activities, err := c.activityService.GetByTaskID(ctx.Context(), id, 50)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.ActivitiesResponse{Activities: activities})
}
