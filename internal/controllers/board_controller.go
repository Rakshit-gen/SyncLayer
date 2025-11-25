package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"taskboard-backend/internal/dto"
	"taskboard-backend/internal/services"
	apperrors "taskboard-backend/pkg/errors"
)

// BoardController handles board HTTP endpoints.
type BoardController struct {
	boardService *services.BoardService
}

// NewBoardController creates a new BoardController instance.
func NewBoardController(boardService *services.BoardService) *BoardController {
	return &BoardController{boardService: boardService}
}

// Create handles POST /boards - creates a new board.
func (c *BoardController) Create(ctx *fiber.Ctx) error {
	var req dto.CreateBoardRequest
	if err := ctx.BodyParser(&req); err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid request body"))
	}

	if req.Name == "" {
		return handleError(ctx, apperrors.NewValidationError("missing required fields", map[string]interface{}{
			"name": "required",
		}))
	}

	if req.TeamID == uuid.Nil {
		return handleError(ctx, apperrors.NewValidationError("missing required fields", map[string]interface{}{
			"team_id": "required",
		}))
	}

	if req.CreatedBy == uuid.Nil {
		return handleError(ctx, apperrors.NewValidationError("missing required fields", map[string]interface{}{
			"created_by": "required",
		}))
	}

	board, err := c.boardService.Create(ctx.Context(), &req)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(dto.BoardFullResponse{Board: board})
}

// GetByID handles GET /boards/:id - retrieves a board with columns and tasks.
func (c *BoardController) GetByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid board ID"))
	}

	board, err := c.boardService.GetByIDFull(ctx.Context(), id)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.BoardFullResponse{Board: board})
}

// GetByTeamID handles GET /teams/:teamId/boards - retrieves boards for a team.
func (c *BoardController) GetByTeamID(ctx *fiber.Ctx) error {
	teamID, err := uuid.Parse(ctx.Params("teamId"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid team ID"))
	}

	boards, err := c.boardService.GetByTeamID(ctx.Context(), teamID)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.BoardsResponse{Boards: boards})
}

// Update handles PUT /boards/:id - updates a board.
func (c *BoardController) Update(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid board ID"))
	}

	var req dto.UpdateBoardRequest
	if err := ctx.BodyParser(&req); err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid request body"))
	}

	// Get user ID from header (frontend sends this)
	userIDStr := ctx.Get("X-User-ID")
	userID, _ := uuid.Parse(userIDStr)

	board, err := c.boardService.Update(ctx.Context(), id, &req, userID)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.BoardResponse{Board: board})
}

// Delete handles DELETE /boards/:id - deletes a board.
func (c *BoardController) Delete(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid board ID"))
	}

	if err := c.boardService.Delete(ctx.Context(), id); err != nil {
		return handleError(ctx, err)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
