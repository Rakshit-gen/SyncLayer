// Package errors provides custom error types for the application.
package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents an application error code.
type ErrorCode string

// Error codes used throughout the application.
const (
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrCodeBadRequest        ErrorCode = "BAD_REQUEST"
	ErrCodeInternalServer    ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeConflict          ErrorCode = "CONFLICT"
	ErrCodeValidation        ErrorCode = "VALIDATION_ERROR"
	ErrCodeForbidden         ErrorCode = "FORBIDDEN"
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrCodeDatabaseError     ErrorCode = "DATABASE_ERROR"
	ErrCodeWebSocketError    ErrorCode = "WEBSOCKET_ERROR"
)

// AppError represents a structured application error.
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	Err        error                  `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewNotFoundError creates a not found error.
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		HTTPStatus: http.StatusNotFound,
	}
}

// NewBadRequestError creates a bad request error.
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeBadRequest,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewInternalError creates an internal server error.
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeInternalServer,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewConflictError creates a conflict error.
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeConflict,
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

// NewValidationError creates a validation error with details.
func NewValidationError(message string, details map[string]interface{}) *AppError {
	return &AppError{
		Code:       ErrCodeValidation,
		Message:    message,
		Details:    details,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewForbiddenError creates a forbidden error.
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeForbidden,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

// NewDatabaseError creates a database error.
func NewDatabaseError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeDatabaseError,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewWebSocketError creates a WebSocket error.
func NewWebSocketError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeWebSocketError,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// ErrorResponse is the JSON structure returned to clients.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody contains the error details.
type ErrorBody struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ToResponse converts an AppError to an ErrorResponse.
func (e *AppError) ToResponse() ErrorResponse {
	return ErrorResponse{
		Error: ErrorBody{
			Code:    e.Code,
			Message: e.Message,
			Details: e.Details,
		},
	}
}
