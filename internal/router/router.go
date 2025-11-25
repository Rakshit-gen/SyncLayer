// Package router defines all HTTP routes and WebSocket endpoints.
package router

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"taskboard-backend/internal/config"
	"taskboard-backend/internal/controllers"
	"taskboard-backend/internal/dto"
	ws "taskboard-backend/internal/websocket"
)

// Router holds all route handlers and dependencies.
type Router struct {
	app                    *fiber.App
	config                 *config.Config
	userController         *controllers.UserController
	teamController         *controllers.TeamController
	boardController        *controllers.BoardController
	columnController       *controllers.ColumnController
	taskController         *controllers.TaskController
	notificationController *controllers.NotificationController
	wsHub                  *ws.Hub
	wsEventHandler         *ws.EventHandler
}

// NewRouter creates a new Router instance.
func NewRouter(
	cfg *config.Config,
	userController *controllers.UserController,
	teamController *controllers.TeamController,
	boardController *controllers.BoardController,
	columnController *controllers.ColumnController,
	taskController *controllers.TaskController,
	notificationController *controllers.NotificationController,
	wsHub *ws.Hub,
	wsEventHandler *ws.EventHandler,
) *Router {
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	return &Router{
		app:                    app,
		config:                 cfg,
		userController:         userController,
		teamController:         teamController,
		boardController:        boardController,
		columnController:       columnController,
		taskController:         taskController,
		notificationController: notificationController,
		wsHub:                  wsHub,
		wsEventHandler:         wsEventHandler,
	}
}

// Setup configures all routes and middleware.
func (r *Router) Setup() {
	// Middleware
	r.app.Use(recover.New())
	r.app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	r.app.Use(cors.New(cors.Config{
		AllowOrigins:     r.config.CORS.Origins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-User-ID",
		AllowCredentials: true,
	}))

	// Health check
	r.app.Get("/health", r.healthCheck)

	// API routes
	api := r.app.Group("/api/v1")

	// User routes
	users := api.Group("/users")
	users.Post("/", r.userController.Create)
	users.Get("/:id", r.userController.GetByID)
	users.Get("/email/:email", r.userController.GetByEmail)
	users.Put("/:id", r.userController.Update)
	users.Get("/:userId/teams", r.teamController.GetByUserID)
	users.Get("/:userId/notifications", r.notificationController.GetByUserID)
	users.Put("/:userId/notifications/read-all", r.notificationController.MarkAllAsRead)

	// Team routes
	teams := api.Group("/teams")
	teams.Post("/", r.teamController.Create)
	teams.Get("/:id", r.teamController.GetByID)
	teams.Put("/:id", r.teamController.Update)
	teams.Delete("/:id", r.teamController.Delete)
	teams.Post("/:id/members", r.teamController.AddMember)
	teams.Put("/:id/members/:userId", r.teamController.UpdateMemberRole)
	teams.Delete("/:id/members/:userId", r.teamController.RemoveMember)
	teams.Get("/:teamId/boards", r.boardController.GetByTeamID)

	// Board routes
	boards := api.Group("/boards")
	boards.Post("/", r.boardController.Create)
	boards.Get("/:id", r.boardController.GetByID)
	boards.Put("/:id", r.boardController.Update)
	boards.Delete("/:id", r.boardController.Delete)

	// Column routes
	columns := api.Group("/columns")
	columns.Post("/", r.columnController.Create)
	columns.Get("/:id", r.columnController.GetByID)
	columns.Put("/:id", r.columnController.Update)
	columns.Put("/:id/position", r.columnController.Move)
	columns.Delete("/:id", r.columnController.Delete)

	// Task routes
	tasks := api.Group("/tasks")
	tasks.Post("/", r.taskController.Create)
	tasks.Get("/:id", r.taskController.GetByID)
	tasks.Put("/:id", r.taskController.Update)
	tasks.Put("/:id/move", r.taskController.Move)
	tasks.Delete("/:id", r.taskController.Delete)
	tasks.Get("/:id/activity", r.taskController.GetActivity)

	// Notification routes
	notifications := api.Group("/notifications")
	notifications.Put("/:id/read", r.notificationController.MarkAsRead)

	// WebSocket route
	r.app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	r.app.Get("/ws", websocket.New(r.handleWebSocket))
}

// handleWebSocket handles WebSocket connections.
func (r *Router) handleWebSocket(c *websocket.Conn) {
	// Get board ID and user ID from query params
	boardIDStr := c.Query("board_id")
	userIDStr := c.Query("user_id")

	boardID, err := uuid.Parse(boardIDStr)
	if err != nil {
		log.Printf("Invalid board_id: %s", boardIDStr)
		c.Close()
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid user_id: %s", userIDStr)
		c.Close()
		return
	}

	// Create client
	client := ws.NewClient(r.wsHub, c, boardID, userID, r.wsEventHandler)

	// Register client with hub
	r.wsHub.GetRegister() <- client

	// Send initial board sync
	if err := r.wsEventHandler.SendBoardSync(context.Background(), client); err != nil {
		log.Printf("Failed to send board sync: %v", err)
	}

	// Start read and write pumps
	go client.WritePump()
	client.ReadPump()
}

// healthCheck handles the health check endpoint.
func (r *Router) healthCheck(c *fiber.Ctx) error {
	return c.JSON(dto.HealthResponse{
		Status: "healthy",
		Services: map[string]string{
			"api": "up",
		},
	})
}

// customErrorHandler handles Fiber errors.
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    code,
			"message": message,
		},
	})
}

// GetApp returns the Fiber app instance.
func (r *Router) GetApp() *fiber.App {
	return r.app
}

// Listen starts the HTTP server.
func (r *Router) Listen(address string) error {
	return r.app.Listen(address)
}

// Shutdown gracefully shuts down the server.
func (r *Router) Shutdown() error {
	return r.app.Shutdown()
}
