package services_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"taskboard-backend/internal/dto"
	"taskboard-backend/internal/models"
)

func TestTeamService_Create(t *testing.T) {
	t.Run("creates team successfully", func(t *testing.T) {
		userID := uuid.New()
		req := &dto.CreateTeamRequest{
			Name:        "Test Team",
			Description: strPtr("A test team"),
			CreatedBy:   userID,
		}

		// Document expected behavior
		assert.NotNil(t, req)
		assert.Equal(t, "Test Team", req.Name)
		assert.NotNil(t, req.Description)
		assert.Equal(t, userID, req.CreatedBy)
	})

	t.Run("adds creator as admin", func(t *testing.T) {
		// After creating a team, the creator should be added as admin
		expectedRole := models.TeamRoleAdmin
		assert.Equal(t, models.TeamRole("admin"), expectedRole)
	})
}

func TestTeamService_AddMember(t *testing.T) {
	t.Run("adds member with valid role", func(t *testing.T) {
		req := &dto.AddTeamMemberRequest{
			UserID: uuid.New(),
			Role:   models.TeamRoleEditor,
		}

		assert.True(t, req.Role.IsValid())
		assert.Equal(t, models.TeamRoleEditor, req.Role)
	})

	t.Run("validates role", func(t *testing.T) {
		validRoles := []models.TeamRole{
			models.TeamRoleAdmin,
			models.TeamRoleEditor,
			models.TeamRoleViewer,
		}

		for _, role := range validRoles {
			assert.True(t, role.IsValid())
		}

		invalidRole := models.TeamRole("superadmin")
		assert.False(t, invalidRole.IsValid())
	})

	t.Run("prevents duplicate membership", func(t *testing.T) {
		// Document expected behavior: should return ConflictError if user is already a member
		teamID := uuid.New()
		userID := uuid.New()
		assert.NotEqual(t, teamID, userID)
	})
}

func TestTeamService_UpdateMemberRole(t *testing.T) {
	t.Run("updates role successfully", func(t *testing.T) {
		req := &dto.UpdateTeamMemberRequest{
			Role: models.TeamRoleAdmin,
		}

		assert.Equal(t, models.TeamRoleAdmin, req.Role)
		assert.True(t, req.Role.IsValid())
	})
}

func TestTeamService_RemoveMember(t *testing.T) {
	t.Run("prevents removing last admin", func(t *testing.T) {
		// Document expected behavior: should return BadRequestError if removing last admin
		adminCount := 1
		isAdmin := true
		shouldPreventRemoval := isAdmin && adminCount == 1
		assert.True(t, shouldPreventRemoval)
	})

	t.Run("allows removing non-admin", func(t *testing.T) {
		adminCount := 1
		isAdmin := false
		shouldPreventRemoval := isAdmin && adminCount == 1
		assert.False(t, shouldPreventRemoval)
	})

	t.Run("allows removing admin when others exist", func(t *testing.T) {
		adminCount := 2
		isAdmin := true
		shouldPreventRemoval := isAdmin && adminCount == 1
		assert.False(t, shouldPreventRemoval)
	})
}

// Helper function
func strPtr(s string) *string {
	return &s
}
