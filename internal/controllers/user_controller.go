// Package controllers handles HTTP request/response logic.
package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"taskboard-backend/internal/dto"
	"taskboard-backend/internal/services"
	apperrors "taskboard-backend/pkg/errors"
)

// UserController handles user HTTP endpoints.
type UserController struct {
	userService *services.UserService
}

// NewUserController creates a new UserController instance.
func NewUserController(userService *services.UserService) *UserController {
	return &UserController{userService: userService}
}

// Create handles POST /users - creates or updates a user.
func (c *UserController) Create(ctx *fiber.Ctx) error {
	var req dto.CreateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid request body"))
	}

	if req.Email == "" || req.Name == "" {
		return handleError(ctx, apperrors.NewValidationError("missing required fields", map[string]interface{}{
			"email": "required",
			"name":  "required",
		}))
	}

	user, err := c.userService.CreateOrUpdate(ctx.Context(), &req)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(dto.UserResponse{User: user})
}

// GetByID handles GET /users/:id - retrieves a user by ID.
func (c *UserController) GetByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid user ID"))
	}

	user, err := c.userService.GetByID(ctx.Context(), id)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.UserResponse{User: user})
}

// GetByEmail handles GET /users/email/:email - retrieves a user by email.
func (c *UserController) GetByEmail(ctx *fiber.Ctx) error {
	email := ctx.Params("email")
	if email == "" {
		return handleError(ctx, apperrors.NewBadRequestError("email is required"))
	}

	user, err := c.userService.GetByEmail(ctx.Context(), email)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.UserResponse{User: user})
}

// Update handles PUT /users/:id - updates a user.
func (c *UserController) Update(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid user ID"))
	}

	var req dto.UpdateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid request body"))
	}

	user, err := c.userService.Update(ctx.Context(), id, &req)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.UserResponse{User: user})
}

// handleError converts app errors to HTTP responses.
func handleError(ctx *fiber.Ctx, err error) error {
	if appErr, ok := err.(*apperrors.AppError); ok {
		return ctx.Status(appErr.HTTPStatus).JSON(appErr.ToResponse())
	}
	
	// Default internal error
	internalErr := apperrors.NewInternalError("an unexpected error occurred", err)
	return ctx.Status(internalErr.HTTPStatus).JSON(internalErr.ToResponse())
}
