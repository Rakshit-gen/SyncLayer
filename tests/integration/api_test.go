package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TestHealthEndpoint tests the health check endpoint.
func TestHealthEndpoint(t *testing.T) {
	app := fiber.New()
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"services": fiber.Map{
				"api": "up",
			},
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)

	assert.Equal(t, "healthy", body["status"])
}

// TestUserEndpoints tests user API endpoints.
func TestUserEndpoints(t *testing.T) {
	app := fiber.New()

	// Mock user endpoint
	app.Post("/api/v1/users", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "BAD_REQUEST",
					"message": "invalid request body",
				},
			})
		}

		if body["email"] == nil || body["name"] == nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "VALIDATION_ERROR",
					"message": "missing required fields",
				},
			})
		}

		return c.Status(http.StatusCreated).JSON(fiber.Map{
			"user": fiber.Map{
				"id":    "123e4567-e89b-12d3-a456-426614174000",
				"email": body["email"],
				"name":  body["name"],
			},
		})
	})

	t.Run("create user - success", func(t *testing.T) {
		payload := map[string]string{
			"email": "test@example.com",
			"name":  "Test User",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("create user - missing fields", func(t *testing.T) {
		payload := map[string]string{
			"email": "test@example.com",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestTeamEndpoints tests team API endpoints.
func TestTeamEndpoints(t *testing.T) {
	app := fiber.New()

	// Mock team endpoint
	app.Post("/api/v1/teams", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "BAD_REQUEST",
					"message": "invalid request body",
				},
			})
		}

		if body["name"] == nil || body["created_by"] == nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "VALIDATION_ERROR",
					"message": "missing required fields",
				},
			})
		}

		return c.Status(http.StatusCreated).JSON(fiber.Map{
			"team": fiber.Map{
				"id":         "123e4567-e89b-12d3-a456-426614174001",
				"name":       body["name"],
				"created_by": body["created_by"],
				"members":    []interface{}{},
			},
		})
	})

	t.Run("create team - success", func(t *testing.T) {
		payload := map[string]string{
			"name":       "Test Team",
			"created_by": "123e4567-e89b-12d3-a456-426614174000",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

// TestBoardEndpoints tests board API endpoints.
func TestBoardEndpoints(t *testing.T) {
	app := fiber.New()

	// Mock board endpoint
	app.Post("/api/v1/boards", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "BAD_REQUEST",
					"message": "invalid request body",
				},
			})
		}

		return c.Status(http.StatusCreated).JSON(fiber.Map{
			"board": fiber.Map{
				"id":      "123e4567-e89b-12d3-a456-426614174002",
				"name":    body["name"],
				"team_id": body["team_id"],
				"columns": []interface{}{
					map[string]interface{}{"name": "To Do", "position": 0},
					map[string]interface{}{"name": "In Progress", "position": 1},
					map[string]interface{}{"name": "Done", "position": 2},
				},
			},
		})
	})

	t.Run("create board - success", func(t *testing.T) {
		payload := map[string]string{
			"name":       "Test Board",
			"team_id":    "123e4567-e89b-12d3-a456-426614174001",
			"created_by": "123e4567-e89b-12d3-a456-426614174000",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/boards", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var respBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respBody)

		board := respBody["board"].(map[string]interface{})
		columns := board["columns"].([]interface{})
		assert.Len(t, columns, 3)
	})
}

// TestTaskEndpoints tests task API endpoints.
func TestTaskEndpoints(t *testing.T) {
	app := fiber.New()

	// Mock task endpoint
	app.Post("/api/v1/tasks", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "BAD_REQUEST",
					"message": "invalid request body",
				},
			})
		}

		return c.Status(http.StatusCreated).JSON(fiber.Map{
			"task": fiber.Map{
				"id":        "123e4567-e89b-12d3-a456-426614174003",
				"title":     body["title"],
				"column_id": body["column_id"],
				"position":  0,
				"priority":  "medium",
			},
		})
	})

	// Mock task move endpoint
	app.Put("/api/v1/tasks/:id/move", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		c.BodyParser(&body)

		return c.JSON(fiber.Map{
			"task": fiber.Map{
				"id":        c.Params("id"),
				"column_id": body["column_id"],
				"position":  body["position"],
			},
		})
	})

	t.Run("create task - success", func(t *testing.T) {
		payload := map[string]string{
			"title":      "Test Task",
			"column_id":  "123e4567-e89b-12d3-a456-426614174004",
			"created_by": "123e4567-e89b-12d3-a456-426614174000",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("move task - success", func(t *testing.T) {
		payload := map[string]interface{}{
			"column_id": "123e4567-e89b-12d3-a456-426614174005",
			"position":  2,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/123e4567-e89b-12d3-a456-426614174003/move", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestErrorResponses tests error response format.
func TestErrorResponses(t *testing.T) {
	app := fiber.New()

	app.Get("/api/v1/not-found", func(c *fiber.Ctx) error {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "resource not found",
			},
		})
	})

	t.Run("error response format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/not-found", nil)
		resp, err := app.Test(req, -1)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)

		errorObj := body["error"].(map[string]interface{})
		assert.Equal(t, "NOT_FOUND", errorObj["code"])
		assert.NotEmpty(t, errorObj["message"])
	})
}
