package services

import (
	"context"

	"github.com/google/uuid"
	"taskboard-backend/internal/dto"
	"taskboard-backend/internal/models"
	"taskboard-backend/internal/repositories"
	apperrors "taskboard-backend/pkg/errors"
)

// TeamService handles team business logic.
type TeamService struct {
	teamRepo     *repositories.TeamRepository
	activityRepo *repositories.ActivityRepository
}

// NewTeamService creates a new TeamService instance.
func NewTeamService(teamRepo *repositories.TeamRepository, activityRepo *repositories.ActivityRepository) *TeamService {
	return &TeamService{
		teamRepo:     teamRepo,
		activityRepo: activityRepo,
	}
}

// Create creates a new team and adds the creator as admin.
func (s *TeamService) Create(ctx context.Context, req *dto.CreateTeamRequest) (*models.TeamWithMembers, error) {
	// Create the team
	team := &models.Team{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   req.CreatedBy,
	}

	createdTeam, err := s.teamRepo.Create(ctx, team)
	if err != nil {
		return nil, err
	}

	// Add the creator as admin
	member := &models.TeamMember{
		TeamID: createdTeam.ID,
		UserID: req.CreatedBy,
		Role:   models.TeamRoleAdmin,
	}

	_, err = s.teamRepo.AddMember(ctx, member)
	if err != nil {
		return nil, err
	}

	// Return team with members
	return s.teamRepo.GetByIDWithMembers(ctx, createdTeam.ID)
}

// GetByID retrieves a team by its ID.
func (s *TeamService) GetByID(ctx context.Context, id uuid.UUID) (*models.Team, error) {
	return s.teamRepo.GetByID(ctx, id)
}

// GetByIDWithMembers retrieves a team with all its members.
func (s *TeamService) GetByIDWithMembers(ctx context.Context, id uuid.UUID) (*models.TeamWithMembers, error) {
	return s.teamRepo.GetByIDWithMembers(ctx, id)
}

// GetByUserID retrieves all teams a user belongs to.
func (s *TeamService) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.TeamWithRole, error) {
	return s.teamRepo.GetByUserID(ctx, userID)
}

// Update updates a team's information.
func (s *TeamService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateTeamRequest) (*models.Team, error) {
	return s.teamRepo.Update(ctx, id, req.Name, req.Description)
}

// Delete deletes a team.
func (s *TeamService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.teamRepo.Delete(ctx, id)
}

// AddMember adds a member to a team.
func (s *TeamService) AddMember(ctx context.Context, teamID uuid.UUID, req *dto.AddTeamMemberRequest) (*models.TeamMember, error) {
	// Check if user is already a member
	isMember, err := s.teamRepo.IsMember(ctx, teamID, req.UserID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, apperrors.NewConflictError("user is already a team member")
	}

	// Validate role
	if !req.Role.IsValid() {
		return nil, apperrors.NewValidationError("invalid role", map[string]interface{}{
			"role": "must be one of: admin, editor, viewer",
		})
	}

	member := &models.TeamMember{
		TeamID: teamID,
		UserID: req.UserID,
		Role:   req.Role,
	}

	createdMember, err := s.teamRepo.AddMember(ctx, member)
	if err != nil {
		return nil, err
	}

	// Log activity
	s.activityRepo.CreateWithDetails(ctx, nil, nil, &req.UserID, models.ActivityMemberAdded, map[string]interface{}{
		"team_id": teamID,
		"role":    req.Role,
	})

	return createdMember, nil
}

// UpdateMemberRole updates a member's role in a team.
func (s *TeamService) UpdateMemberRole(ctx context.Context, teamID, userID uuid.UUID, req *dto.UpdateTeamMemberRequest) (*models.TeamMember, error) {
	// Validate role
	if !req.Role.IsValid() {
		return nil, apperrors.NewValidationError("invalid role", map[string]interface{}{
			"role": "must be one of: admin, editor, viewer",
		})
	}

	// Get current member to log the change
	currentMember, err := s.teamRepo.GetMember(ctx, teamID, userID)
	if err != nil {
		return nil, err
	}

	updatedMember, err := s.teamRepo.UpdateMemberRole(ctx, teamID, userID, req.Role)
	if err != nil {
		return nil, err
	}

	// Log activity
	s.activityRepo.CreateWithDetails(ctx, nil, nil, &userID, models.ActivityMemberRoleChanged, map[string]interface{}{
		"team_id":       teamID,
		"previous_role": currentMember.Role,
		"new_role":      req.Role,
	})

	return updatedMember, nil
}

// RemoveMember removes a member from a team.
func (s *TeamService) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	// Check if this is the last admin
	members, err := s.teamRepo.GetMembers(ctx, teamID)
	if err != nil {
		return err
	}

	adminCount := 0
	isAdmin := false
	for _, m := range members {
		if m.Role == models.TeamRoleAdmin {
			adminCount++
			if m.UserID == userID {
				isAdmin = true
			}
		}
	}

	if isAdmin && adminCount == 1 {
		return apperrors.NewBadRequestError("cannot remove the last admin from the team")
	}

	err = s.teamRepo.RemoveMember(ctx, teamID, userID)
	if err != nil {
		return err
	}

	// Log activity
	s.activityRepo.CreateWithDetails(ctx, nil, nil, &userID, models.ActivityMemberRemoved, map[string]interface{}{
		"team_id": teamID,
	})

	return nil
}

// GetMembers retrieves all members of a team.
func (s *TeamService) GetMembers(ctx context.Context, teamID uuid.UUID) ([]models.TeamMember, error) {
	return s.teamRepo.GetMembers(ctx, teamID)
}

// GetMember retrieves a specific member of a team.
func (s *TeamService) GetMember(ctx context.Context, teamID, userID uuid.UUID) (*models.TeamMember, error) {
	return s.teamRepo.GetMember(ctx, teamID, userID)
}

// IsMember checks if a user is a member of a team.
func (s *TeamService) IsMember(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	return s.teamRepo.IsMember(ctx, teamID, userID)
}
