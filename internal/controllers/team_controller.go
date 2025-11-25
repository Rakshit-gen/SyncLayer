package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"taskboard-backend/internal/dto"
	"taskboard-backend/internal/services"
	apperrors "taskboard-backend/pkg/errors"
)

// TeamController handles team HTTP endpoints.
type TeamController struct {
	teamService *services.TeamService
}

// NewTeamController creates a new TeamController instance.
func NewTeamController(teamService *services.TeamService) *TeamController {
	return &TeamController{teamService: teamService}
}

// Create handles POST /teams - creates a new team.
func (c *TeamController) Create(ctx *fiber.Ctx) error {
	var req dto.CreateTeamRequest
	if err := ctx.BodyParser(&req); err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid request body"))
	}

	if req.Name == "" {
		return handleError(ctx, apperrors.NewValidationError("missing required fields", map[string]interface{}{
			"name": "required",
		}))
	}

	if req.CreatedBy == uuid.Nil {
		return handleError(ctx, apperrors.NewValidationError("missing required fields", map[string]interface{}{
			"created_by": "required",
		}))
	}

	team, err := c.teamService.Create(ctx.Context(), &req)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(dto.TeamWithMembersResponse{Team: team})
}

// GetByID handles GET /teams/:id - retrieves a team by ID with members.
func (c *TeamController) GetByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid team ID"))
	}

	team, err := c.teamService.GetByIDWithMembers(ctx.Context(), id)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.TeamWithMembersResponse{Team: team})
}

// GetByUserID handles GET /users/:userId/teams - retrieves teams for a user.
func (c *TeamController) GetByUserID(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("userId"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid user ID"))
	}

	teams, err := c.teamService.GetByUserID(ctx.Context(), userID)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.TeamsResponse{Teams: teams})
}

// Update handles PUT /teams/:id - updates a team.
func (c *TeamController) Update(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid team ID"))
	}

	var req dto.UpdateTeamRequest
	if err := ctx.BodyParser(&req); err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid request body"))
	}

	team, err := c.teamService.Update(ctx.Context(), id, &req)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.TeamResponse{Team: team})
}

// Delete handles DELETE /teams/:id - deletes a team.
func (c *TeamController) Delete(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid team ID"))
	}

	if err := c.teamService.Delete(ctx.Context(), id); err != nil {
		return handleError(ctx, err)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

// AddMember handles POST /teams/:id/members - adds a member to a team.
func (c *TeamController) AddMember(ctx *fiber.Ctx) error {
	teamID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid team ID"))
	}

	var req dto.AddTeamMemberRequest
	if err := ctx.BodyParser(&req); err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid request body"))
	}

	if req.UserID == uuid.Nil {
		return handleError(ctx, apperrors.NewValidationError("missing required fields", map[string]interface{}{
			"user_id": "required",
		}))
	}

	member, err := c.teamService.AddMember(ctx.Context(), teamID, &req)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(dto.TeamMemberResponse{Member: member})
}

// UpdateMemberRole handles PUT /teams/:id/members/:userId - updates a member's role.
func (c *TeamController) UpdateMemberRole(ctx *fiber.Ctx) error {
	teamID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid team ID"))
	}

	userID, err := uuid.Parse(ctx.Params("userId"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid user ID"))
	}

	var req dto.UpdateTeamMemberRequest
	if err := ctx.BodyParser(&req); err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid request body"))
	}

	member, err := c.teamService.UpdateMemberRole(ctx.Context(), teamID, userID, &req)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.TeamMemberResponse{Member: member})
}

// RemoveMember handles DELETE /teams/:id/members/:userId - removes a member.
func (c *TeamController) RemoveMember(ctx *fiber.Ctx) error {
	teamID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid team ID"))
	}

	userID, err := uuid.Parse(ctx.Params("userId"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid user ID"))
	}

	if err := c.teamService.RemoveMember(ctx.Context(), teamID, userID); err != nil {
		return handleError(ctx, err)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
