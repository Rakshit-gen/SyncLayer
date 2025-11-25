package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"taskboard-backend/internal/dto"
	"taskboard-backend/internal/services"
	apperrors "taskboard-backend/pkg/errors"
)

// ColumnController handles column HTTP endpoints.
type ColumnController struct {
	columnService *services.ColumnService
}

// NewColumnController creates a new ColumnController instance.
func NewColumnController(columnService *services.ColumnService) *ColumnController {
	return &ColumnController{columnService: columnService}
}

// Create handles POST /columns - creates a new column.
func (c *ColumnController) Create(ctx *fiber.Ctx) error {
	var req dto.CreateColumnRequest
	if err := ctx.BodyParser(&req); err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid request body"))
	}

	if req.Name == "" {
		return handleError(ctx, apperrors.NewValidationError("missing required fields", map[string]interface{}{
			"name": "required",
		}))
	}

	if req.BoardID == uuid.Nil {
		return handleError(ctx, apperrors.NewValidationError("missing required fields", map[string]interface{}{
			"board_id": "required",
		}))
	}

	// Get user ID from header
	userIDStr := ctx.Get("X-User-ID")
	userID, _ := uuid.Parse(userIDStr)

	column, err := c.columnService.Create(ctx.Context(), &req, userID)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(dto.ColumnResponse{Column: column})
}

// GetByID handles GET /columns/:id - retrieves a column.
func (c *ColumnController) GetByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid column ID"))
	}

	column, err := c.columnService.GetByIDWithTasks(ctx.Context(), id)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(column)
}

// Update handles PUT /columns/:id - updates a column.
func (c *ColumnController) Update(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid column ID"))
	}

	var req dto.UpdateColumnRequest
	if err := ctx.BodyParser(&req); err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid request body"))
	}

	// Get user ID from header
	userIDStr := ctx.Get("X-User-ID")
	userID, _ := uuid.Parse(userIDStr)

	column, err := c.columnService.Update(ctx.Context(), id, &req, userID)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.ColumnResponse{Column: column})
}

// Move handles PUT /columns/:id/position - moves a column.
func (c *ColumnController) Move(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid column ID"))
	}

	var req dto.MoveColumnRequest
	if err := ctx.BodyParser(&req); err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid request body"))
	}

	// Get user ID from header
	userIDStr := ctx.Get("X-User-ID")
	userID, _ := uuid.Parse(userIDStr)

	column, err := c.columnService.Move(ctx.Context(), id, &req, userID)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.ColumnResponse{Column: column})
}

// Delete handles DELETE /columns/:id - deletes a column.
func (c *ColumnController) Delete(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid column ID"))
	}

	// Get user ID from header
	userIDStr := ctx.Get("X-User-ID")
	userID, _ := uuid.Parse(userIDStr)

	if err := c.columnService.Delete(ctx.Context(), id, userID); err != nil {
		return handleError(ctx, err)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
