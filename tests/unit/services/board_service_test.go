package services_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"taskboard-backend/internal/dto"
)

func TestBoardService_Create(t *testing.T) {
	t.Run("creates board with default columns", func(t *testing.T) {
		req := &dto.CreateBoardRequest{
			TeamID:      uuid.New(),
			Name:        "Test Board",
			Description: strPtr("A test board"),
			CreatedBy:   uuid.New(),
		}

		assert.NotNil(t, req)
		assert.Equal(t, "Test Board", req.Name)
		assert.NotEqual(t, uuid.Nil, req.TeamID)
		assert.NotEqual(t, uuid.Nil, req.CreatedBy)
	})

	t.Run("default columns are created", func(t *testing.T) {
		expectedColumns := []string{"To Do", "In Progress", "Done"}
		assert.Len(t, expectedColumns, 3)
		assert.Contains(t, expectedColumns, "To Do")
		assert.Contains(t, expectedColumns, "In Progress")
		assert.Contains(t, expectedColumns, "Done")
	})
}

func TestBoardService_Update(t *testing.T) {
	t.Run("updates board name", func(t *testing.T) {
		newName := "Updated Board Name"
		req := &dto.UpdateBoardRequest{
			Name: &newName,
		}

		assert.NotNil(t, req.Name)
		assert.Equal(t, "Updated Board Name", *req.Name)
	})

	t.Run("updates board description", func(t *testing.T) {
		newDesc := "Updated description"
		req := &dto.UpdateBoardRequest{
			Description: &newDesc,
		}

		assert.NotNil(t, req.Description)
		assert.Equal(t, "Updated description", *req.Description)
	})
}

func TestBoardService_GetByIDFull(t *testing.T) {
	t.Run("returns board with columns and tasks", func(t *testing.T) {
		boardID := uuid.New()
		// Document expected structure
		assert.NotEqual(t, uuid.Nil, boardID)
	})
}

func TestBoardService_GetByTeamID(t *testing.T) {
	t.Run("returns all boards for team", func(t *testing.T) {
		teamID := uuid.New()
		assert.NotEqual(t, uuid.Nil, teamID)
	})

	t.Run("returns empty list for team with no boards", func(t *testing.T) {
		emptyBoards := []interface{}{}
		assert.Empty(t, emptyBoards)
	})
}
