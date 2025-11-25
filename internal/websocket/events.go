package websocket

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"taskboard-backend/internal/dto"
	"taskboard-backend/internal/models"
	"taskboard-backend/internal/services"
)

// EventType represents the type of WebSocket event.
type EventType string

// Client -> Server event types
const (
	EventBoardJoin  EventType = "board.join"
	EventBoardLeave EventType = "board.leave"

	EventTaskCreate EventType = "task.create"
	EventTaskUpdate EventType = "task.update"
	EventTaskMove   EventType = "task.move"
	EventTaskDelete EventType = "task.delete"

	EventColumnCreate EventType = "column.create"
	EventColumnUpdate EventType = "column.update"
	EventColumnMove   EventType = "column.move"
	EventColumnDelete EventType = "column.delete"

	EventPresenceCursor EventType = "presence.cursor"
)

// Server -> Client event types
const (
	EventBoardSync EventType = "board.sync"

	EventTaskCreated EventType = "task.created"
	EventTaskUpdated EventType = "task.updated"
	EventTaskMoved   EventType = "task.moved"
	EventTaskDeleted EventType = "task.deleted"

	EventColumnCreated EventType = "column.created"
	EventColumnUpdated EventType = "column.updated"
	EventColumnMoved   EventType = "column.moved"
	EventColumnDeleted EventType = "column.deleted"

	EventPresenceJoined  EventType = "presence.joined"
	EventPresenceLeft    EventType = "presence.left"
	EventPresenceCursors EventType = "presence.cursors"

	EventNotificationPush EventType = "notification.push"

	EventError EventType = "error"
)

// Event represents a WebSocket event.
type Event struct {
	Type    EventType              `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// EventHandler handles incoming WebSocket events.
type EventHandler struct {
	boardService  *services.BoardService
	columnService *services.ColumnService
	taskService   *services.TaskService
}

// NewEventHandler creates a new EventHandler instance.
func NewEventHandler(
	boardService *services.BoardService,
	columnService *services.ColumnService,
	taskService *services.TaskService,
) *EventHandler {
	return &EventHandler{
		boardService:  boardService,
		columnService: columnService,
		taskService:   taskService,
	}
}

// HandleEvent processes an incoming WebSocket event.
func (h *EventHandler) HandleEvent(client *Client, event *Event) {
	ctx := context.Background()

	switch event.Type {
	case EventTaskCreate:
		h.handleTaskCreate(ctx, client, event)
	case EventTaskUpdate:
		h.handleTaskUpdate(ctx, client, event)
	case EventTaskMove:
		h.handleTaskMove(ctx, client, event)
	case EventTaskDelete:
		h.handleTaskDelete(ctx, client, event)
	case EventColumnCreate:
		h.handleColumnCreate(ctx, client, event)
	case EventColumnUpdate:
		h.handleColumnUpdate(ctx, client, event)
	case EventColumnMove:
		h.handleColumnMove(ctx, client, event)
	case EventColumnDelete:
		h.handleColumnDelete(ctx, client, event)
	case EventPresenceCursor:
		h.handlePresenceCursor(client, event)
	default:
		log.Printf("Unknown event type: %s", event.Type)
	}
}

// handleTaskCreate handles task creation events.
func (h *EventHandler) handleTaskCreate(ctx context.Context, client *Client, event *Event) {
	payload := event.Payload

	columnIDStr, _ := payload["column_id"].(string)
	columnID, err := uuid.Parse(columnIDStr)
	if err != nil {
		client.SendError("INVALID_COLUMN_ID", "Invalid column ID")
		return
	}

	title, _ := payload["title"].(string)
	description, _ := payload["description"].(string)

	req := &dto.CreateTaskRequest{
		ColumnID:    columnID,
		Title:       title,
		Description: &description,
		CreatedBy:   client.UserID,
	}

	task, err := h.taskService.Create(ctx, req)
	if err != nil {
		client.SendError("TASK_CREATE_FAILED", err.Error())
		return
	}

	// Broadcast to all clients in the board
	broadcastEvent := &Event{
		Type: EventTaskCreated,
		Payload: map[string]interface{}{
			"task":    task,
			"user_id": client.UserID,
		},
	}
	client.hub.BroadcastEvent(client.BoardID, broadcastEvent, nil)
}

// handleTaskUpdate handles task update events.
func (h *EventHandler) handleTaskUpdate(ctx context.Context, client *Client, event *Event) {
	payload := event.Payload

	taskIDStr, _ := payload["task_id"].(string)
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		client.SendError("INVALID_TASK_ID", "Invalid task ID")
		return
	}

	changes, _ := payload["changes"].(map[string]interface{})

	req := &dto.UpdateTaskRequest{}
	if title, ok := changes["title"].(string); ok {
		req.Title = &title
	}
	if description, ok := changes["description"].(string); ok {
		req.Description = &description
	}
	if priority, ok := changes["priority"].(string); ok {
		p := models.TaskPriority(priority)
		req.Priority = &p
	}

	task, err := h.taskService.Update(ctx, taskID, req, client.UserID)
	if err != nil {
		client.SendError("TASK_UPDATE_FAILED", err.Error())
		return
	}

	// Broadcast to all clients in the board
	broadcastEvent := &Event{
		Type: EventTaskUpdated,
		Payload: map[string]interface{}{
			"task":    task,
			"user_id": client.UserID,
		},
	}
	client.hub.BroadcastEvent(client.BoardID, broadcastEvent, nil)
}

// handleTaskMove handles task move events.
func (h *EventHandler) handleTaskMove(ctx context.Context, client *Client, event *Event) {
	payload := event.Payload

	taskIDStr, _ := payload["task_id"].(string)
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		client.SendError("INVALID_TASK_ID", "Invalid task ID")
		return
	}

	columnIDStr, _ := payload["column_id"].(string)
	columnID, err := uuid.Parse(columnIDStr)
	if err != nil {
		client.SendError("INVALID_COLUMN_ID", "Invalid column ID")
		return
	}

	position, _ := payload["position"].(float64)

	// Get original column ID for broadcast
	originalTask, _ := h.taskService.GetByID(ctx, taskID)
	fromColumnID := uuid.Nil
	if originalTask != nil {
		fromColumnID = originalTask.ColumnID
	}

	req := &dto.MoveTaskRequest{
		ColumnID: columnID,
		Position: int(position),
	}

	task, err := h.taskService.Move(ctx, taskID, req, client.UserID)
	if err != nil {
		client.SendError("TASK_MOVE_FAILED", err.Error())
		return
	}

	// Broadcast to all clients in the board
	broadcastEvent := &Event{
		Type: EventTaskMoved,
		Payload: map[string]interface{}{
			"task":           task,
			"from_column_id": fromColumnID,
			"user_id":        client.UserID,
		},
	}
	client.hub.BroadcastEvent(client.BoardID, broadcastEvent, nil)
}

// handleTaskDelete handles task deletion events.
func (h *EventHandler) handleTaskDelete(ctx context.Context, client *Client, event *Event) {
	payload := event.Payload

	taskIDStr, _ := payload["task_id"].(string)
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		client.SendError("INVALID_TASK_ID", "Invalid task ID")
		return
	}

	err = h.taskService.Delete(ctx, taskID, client.UserID)
	if err != nil {
		client.SendError("TASK_DELETE_FAILED", err.Error())
		return
	}

	// Broadcast to all clients in the board
	broadcastEvent := &Event{
		Type: EventTaskDeleted,
		Payload: map[string]interface{}{
			"task_id": taskID,
			"user_id": client.UserID,
		},
	}
	client.hub.BroadcastEvent(client.BoardID, broadcastEvent, nil)
}

// handleColumnCreate handles column creation events.
func (h *EventHandler) handleColumnCreate(ctx context.Context, client *Client, event *Event) {
	payload := event.Payload

	name, _ := payload["name"].(string)
	position, _ := payload["position"].(float64)
	color, _ := payload["color"].(string)

	req := &dto.CreateColumnRequest{
		BoardID:  client.BoardID,
		Name:     name,
		Position: int(position),
		Color:    &color,
	}

	column, err := h.columnService.Create(ctx, req, client.UserID)
	if err != nil {
		client.SendError("COLUMN_CREATE_FAILED", err.Error())
		return
	}

	// Broadcast to all clients in the board
	broadcastEvent := &Event{
		Type: EventColumnCreated,
		Payload: map[string]interface{}{
			"column":  column,
			"user_id": client.UserID,
		},
	}
	client.hub.BroadcastEvent(client.BoardID, broadcastEvent, nil)
}

// handleColumnUpdate handles column update events.
func (h *EventHandler) handleColumnUpdate(ctx context.Context, client *Client, event *Event) {
	payload := event.Payload

	columnIDStr, _ := payload["column_id"].(string)
	columnID, err := uuid.Parse(columnIDStr)
	if err != nil {
		client.SendError("INVALID_COLUMN_ID", "Invalid column ID")
		return
	}

	changes, _ := payload["changes"].(map[string]interface{})

	req := &dto.UpdateColumnRequest{}
	if name, ok := changes["name"].(string); ok {
		req.Name = &name
	}
	if color, ok := changes["color"].(string); ok {
		req.Color = &color
	}

	column, err := h.columnService.Update(ctx, columnID, req, client.UserID)
	if err != nil {
		client.SendError("COLUMN_UPDATE_FAILED", err.Error())
		return
	}

	// Broadcast to all clients in the board
	broadcastEvent := &Event{
		Type: EventColumnUpdated,
		Payload: map[string]interface{}{
			"column":  column,
			"user_id": client.UserID,
		},
	}
	client.hub.BroadcastEvent(client.BoardID, broadcastEvent, nil)
}

// handleColumnMove handles column move events.
func (h *EventHandler) handleColumnMove(ctx context.Context, client *Client, event *Event) {
	payload := event.Payload

	columnIDStr, _ := payload["column_id"].(string)
	columnID, err := uuid.Parse(columnIDStr)
	if err != nil {
		client.SendError("INVALID_COLUMN_ID", "Invalid column ID")
		return
	}

	position, _ := payload["position"].(float64)

	req := &dto.MoveColumnRequest{
		Position: int(position),
	}

	column, err := h.columnService.Move(ctx, columnID, req, client.UserID)
	if err != nil {
		client.SendError("COLUMN_MOVE_FAILED", err.Error())
		return
	}

	// Broadcast to all clients in the board
	broadcastEvent := &Event{
		Type: EventColumnMoved,
		Payload: map[string]interface{}{
			"column":  column,
			"user_id": client.UserID,
		},
	}
	client.hub.BroadcastEvent(client.BoardID, broadcastEvent, nil)
}

// handleColumnDelete handles column deletion events.
func (h *EventHandler) handleColumnDelete(ctx context.Context, client *Client, event *Event) {
	payload := event.Payload

	columnIDStr, _ := payload["column_id"].(string)
	columnID, err := uuid.Parse(columnIDStr)
	if err != nil {
		client.SendError("INVALID_COLUMN_ID", "Invalid column ID")
		return
	}

	err = h.columnService.Delete(ctx, columnID, client.UserID)
	if err != nil {
		client.SendError("COLUMN_DELETE_FAILED", err.Error())
		return
	}

	// Broadcast to all clients in the board
	broadcastEvent := &Event{
		Type: EventColumnDeleted,
		Payload: map[string]interface{}{
			"column_id": columnID,
			"user_id":   client.UserID,
		},
	}
	client.hub.BroadcastEvent(client.BoardID, broadcastEvent, nil)
}

// handlePresenceCursor handles cursor position updates.
func (h *EventHandler) handlePresenceCursor(client *Client, event *Event) {
	payload := event.Payload

	x, _ := payload["x"].(float64)
	y, _ := payload["y"].(float64)

	// Broadcast cursor position to other clients
	broadcastEvent := &Event{
		Type: EventPresenceCursors,
		Payload: map[string]interface{}{
			"cursors": []map[string]interface{}{
				{
					"user_id": client.UserID,
					"x":       x,
					"y":       y,
				},
			},
		},
	}
	client.hub.BroadcastEvent(client.BoardID, broadcastEvent, client)
}

// SendBoardSync sends the full board state to a client.
func (h *EventHandler) SendBoardSync(ctx context.Context, client *Client) error {
	board, err := h.boardService.GetByIDFull(ctx, client.BoardID)
	if err != nil {
		return err
	}

	event := &Event{
		Type: EventBoardSync,
		Payload: map[string]interface{}{
			"board": board,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	client.send <- data
	return nil
}
