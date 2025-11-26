// Package main is the entry point for the TaskBoard backend server.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"taskboard-backend/internal/config"
	"taskboard-backend/internal/controllers"
	"taskboard-backend/internal/database"
	"taskboard-backend/internal/repositories"
	"taskboard-backend/internal/router"
	"taskboard-backend/internal/services"
	"taskboard-backend/internal/websocket"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting SyncLayer server in %s mode...", cfg.Env)
	log.Printf("CORS Origins: %s", cfg.CORS.Origins)

	// Initialize PostgreSQL
	postgres, err := database.NewPostgresDB(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgres.Close()
	log.Println("Connected to PostgreSQL")

	// Run migrations
	if err := postgres.RunMigrations("internal/database/migrations"); err != nil {
		log.Printf("Warning: Migration error (may already exist): %v", err)
	}
	log.Println("Database migrations completed")

	// Initialize Redis
	redis, err := database.NewRedisDB(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()
	log.Println("Connected to Redis")

	// Initialize repositories
	userRepo := repositories.NewUserRepository(postgres.Pool)
	teamRepo := repositories.NewTeamRepository(postgres.Pool)
	boardRepo := repositories.NewBoardRepository(postgres.Pool)
	columnRepo := repositories.NewColumnRepository(postgres.Pool)
	taskRepo := repositories.NewTaskRepository(postgres.Pool)
	activityRepo := repositories.NewActivityRepository(postgres.Pool)
	notificationRepo := repositories.NewNotificationRepository(postgres.Pool)

	// Initialize services
	userService := services.NewUserService(userRepo)
	teamService := services.NewTeamService(teamRepo, activityRepo)
	boardService := services.NewBoardService(boardRepo, columnRepo, activityRepo)
	columnService := services.NewColumnService(columnRepo, activityRepo)
	taskService := services.NewTaskService(taskRepo, activityRepo, notificationRepo)
	activityService := services.NewActivityService(activityRepo)
	notificationService := services.NewNotificationService(notificationRepo)

	// Initialize controllers
	userController := controllers.NewUserController(userService)
	teamController := controllers.NewTeamController(teamService)
	boardController := controllers.NewBoardController(boardService)
	columnController := controllers.NewColumnController(columnService)
	taskController := controllers.NewTaskController(taskService, activityService)
	notificationController := controllers.NewNotificationController(notificationService)

	// Initialize WebSocket components
	redisPubSub := websocket.NewRedisPubSub(redis)
	wsHub := websocket.NewHub(redisPubSub)
	wsEventHandler := websocket.NewEventHandler(boardService, columnService, taskService)

	// Start WebSocket hub
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go wsHub.Run(ctx)

	// Initialize router
	r := router.NewRouter(
		cfg,
		userController,
		teamController,
		boardController,
		columnController,
		taskController,
		notificationController,
		wsHub,
		wsEventHandler,
	)
	r.Setup()

	// Start server in goroutine
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	go func() {
		log.Printf("Server listening on %s", serverAddr)
		if err := r.Listen(serverAddr); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give outstanding requests 10 seconds to complete
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Cancel the WebSocket hub context
	cancel()

	// Shutdown the HTTP server
	if err := r.Shutdown(); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Wait for shutdown context
	<-shutdownCtx.Done()

	log.Println("Server exited gracefully")
}
