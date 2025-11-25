package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"taskboard-backend/internal/dto"
	"taskboard-backend/internal/services"
	apperrors "taskboard-backend/pkg/errors"
)

// NotificationController handles notification HTTP endpoints.
type NotificationController struct {
	notificationService *services.NotificationService
}

// NewNotificationController creates a new NotificationController instance.
func NewNotificationController(notificationService *services.NotificationService) *NotificationController {
	return &NotificationController{notificationService: notificationService}
}

// GetByUserID handles GET /users/:userId/notifications - retrieves notifications for a user.
func (c *NotificationController) GetByUserID(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("userId"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid user ID"))
	}

	unreadOnly := ctx.Query("unread") == "true"
	limit := ctx.QueryInt("limit", 50)

	notifications, unreadCount, err := c.notificationService.GetByUserID(ctx.Context(), userID, unreadOnly, limit)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.NotificationsResponse{
		Notifications: notifications,
		UnreadCount:   unreadCount,
	})
}

// MarkAsRead handles PUT /notifications/:id/read - marks a notification as read.
func (c *NotificationController) MarkAsRead(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid notification ID"))
	}

	notification, err := c.notificationService.MarkAsRead(ctx.Context(), id)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(dto.NotificationResponse{Notification: notification})
}

// MarkAllAsRead handles PUT /users/:userId/notifications/read-all - marks all as read.
func (c *NotificationController) MarkAllAsRead(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("userId"))
	if err != nil {
		return handleError(ctx, apperrors.NewBadRequestError("invalid user ID"))
	}

	if err := c.notificationService.MarkAllAsRead(ctx.Context(), userID); err != nil {
		return handleError(ctx, err)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
