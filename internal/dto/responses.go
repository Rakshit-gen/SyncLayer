package dto

import "taskboard-backend/internal/models"

// ===============================
// Generic Responses
// ===============================

// SuccessResponse represents a generic success response.
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PaginatedResponse represents a paginated response.
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int64       `json:"total_count"`
	TotalPages int         `json:"total_pages"`
}

// ===============================
// User Responses
// ===============================

// UserResponse wraps a user for API response.
type UserResponse struct {
	User *models.User `json:"user"`
}

// UsersResponse wraps multiple users for API response.
type UsersResponse struct {
	Users []models.User `json:"users"`
}

// ===============================
// Team Responses
// ===============================

// TeamResponse wraps a team for API response.
type TeamResponse struct {
	Team *models.Team `json:"team"`
}

// TeamWithMembersResponse wraps a team with members for API response.
type TeamWithMembersResponse struct {
	Team *models.TeamWithMembers `json:"team"`
}

// TeamsResponse wraps multiple teams for API response.
type TeamsResponse struct {
	Teams []models.TeamWithRole `json:"teams"`
}

// TeamMemberResponse wraps a team member for API response.
type TeamMemberResponse struct {
	Member *models.TeamMember `json:"member"`
}

// ===============================
// Board Responses
// ===============================

// BoardResponse wraps a board for API response.
type BoardResponse struct {
	Board *models.Board `json:"board"`
}

// BoardFullResponse wraps a full board (with columns and tasks) for API response.
type BoardFullResponse struct {
	Board *models.BoardFull `json:"board"`
}

// BoardsResponse wraps multiple boards for API response.
type BoardsResponse struct {
	Boards []models.Board `json:"boards"`
}

// ===============================
// Column Responses
// ===============================

// ColumnResponse wraps a column for API response.
type ColumnResponse struct {
	Column *models.Column `json:"column"`
}

// ColumnsResponse wraps multiple columns for API response.
type ColumnsResponse struct {
	Columns []models.Column `json:"columns"`
}

// ===============================
// Task Responses
// ===============================

// TaskResponse wraps a task for API response.
type TaskResponse struct {
	Task *models.Task `json:"task"`
}

// TasksResponse wraps multiple tasks for API response.
type TasksResponse struct {
	Tasks []models.Task `json:"tasks"`
}

// TaskWithAssigneeResponse wraps a task with assignee for API response.
type TaskWithAssigneeResponse struct {
	Task *models.TaskWithAssignee `json:"task"`
}

// ===============================
// Activity Responses
// ===============================

// ActivityResponse wraps an activity log for API response.
type ActivityResponse struct {
	Activity *models.ActivityLog `json:"activity"`
}

// ActivitiesResponse wraps multiple activity logs for API response.
type ActivitiesResponse struct {
	Activities []models.ActivityLogWithUser `json:"activities"`
}

// ===============================
// Notification Responses
// ===============================

// NotificationResponse wraps a notification for API response.
type NotificationResponse struct {
	Notification *models.Notification `json:"notification"`
}

// NotificationsResponse wraps multiple notifications for API response.
type NotificationsResponse struct {
	Notifications []models.Notification `json:"notifications"`
	UnreadCount   int64                 `json:"unread_count"`
}

// ===============================
// Health Check Response
// ===============================

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}
