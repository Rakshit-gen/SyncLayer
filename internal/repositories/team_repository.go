package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"taskboard-backend/internal/models"
	apperrors "taskboard-backend/pkg/errors"
)

// TeamRepository handles team data access operations.
type TeamRepository struct {
	pool *pgxpool.Pool
}

// NewTeamRepository creates a new TeamRepository instance.
func NewTeamRepository(pool *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{pool: pool}
}

// Create creates a new team.
func (r *TeamRepository) Create(ctx context.Context, team *models.Team) (*models.Team, error) {
	query := `
		INSERT INTO teams (name, description, created_by)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, created_by, created_at, updated_at
	`

	var result models.Team
	err := r.pool.QueryRow(ctx, query, team.Name, team.Description, team.CreatedBy).Scan(
		&result.ID,
		&result.Name,
		&result.Description,
		&result.CreatedBy,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to create team", err)
	}

	return &result, nil
}

// GetByID retrieves a team by its ID.
func (r *TeamRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Team, error) {
	query := `
		SELECT id, name, description, created_by, created_at, updated_at
		FROM teams
		WHERE id = $1
	`

	var team models.Team
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&team.ID,
		&team.Name,
		&team.Description,
		&team.CreatedBy,
		&team.CreatedAt,
		&team.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewNotFoundError("team")
		}
		return nil, apperrors.NewDatabaseError("failed to get team", err)
	}

	return &team, nil
}

// GetByIDWithMembers retrieves a team with all its members.
func (r *TeamRepository) GetByIDWithMembers(ctx context.Context, id uuid.UUID) (*models.TeamWithMembers, error) {
	team, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	members, err := r.GetMembers(ctx, id)
	if err != nil {
		return nil, err
	}

	return &models.TeamWithMembers{
		Team:    *team,
		Members: members,
	}, nil
}

// GetByUserID retrieves all teams a user belongs to.
func (r *TeamRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.TeamWithRole, error) {
	query := `
		SELECT t.id, t.name, t.description, t.created_by, t.created_at, t.updated_at, tm.role
		FROM teams t
		INNER JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = $1
		ORDER BY t.name ASC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to get user teams", err)
	}
	defer rows.Close()

	var teams []models.TeamWithRole
	for rows.Next() {
		var team models.TeamWithRole
		if err := rows.Scan(
			&team.ID,
			&team.Name,
			&team.Description,
			&team.CreatedBy,
			&team.CreatedAt,
			&team.UpdatedAt,
			&team.Role,
		); err != nil {
			return nil, apperrors.NewDatabaseError("failed to scan team", err)
		}
		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.NewDatabaseError("error iterating teams", err)
	}

	return teams, nil
}

// Update updates a team's information.
func (r *TeamRepository) Update(ctx context.Context, id uuid.UUID, name *string, description *string) (*models.Team, error) {
	query := `
		UPDATE teams
		SET name = COALESCE($2, name),
			description = COALESCE($3, description),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, description, created_by, created_at, updated_at
	`

	var team models.Team
	err := r.pool.QueryRow(ctx, query, id, name, description).Scan(
		&team.ID,
		&team.Name,
		&team.Description,
		&team.CreatedBy,
		&team.CreatedAt,
		&team.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewNotFoundError("team")
		}
		return nil, apperrors.NewDatabaseError("failed to update team", err)
	}

	return &team, nil
}

// Delete deletes a team by ID.
func (r *TeamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM teams WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return apperrors.NewDatabaseError("failed to delete team", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NewNotFoundError("team")
	}

	return nil
}

// AddMember adds a member to a team.
func (r *TeamRepository) AddMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error) {
	query := `
		INSERT INTO team_members (team_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, team_id, user_id, role, joined_at
	`

	var result models.TeamMember
	err := r.pool.QueryRow(ctx, query, member.TeamID, member.UserID, member.Role).Scan(
		&result.ID,
		&result.TeamID,
		&result.UserID,
		&result.Role,
		&result.JoinedAt,
	)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to add team member", err)
	}

	return &result, nil
}

// UpdateMemberRole updates a member's role in a team.
func (r *TeamRepository) UpdateMemberRole(ctx context.Context, teamID, userID uuid.UUID, role models.TeamRole) (*models.TeamMember, error) {
	query := `
		UPDATE team_members
		SET role = $3
		WHERE team_id = $1 AND user_id = $2
		RETURNING id, team_id, user_id, role, joined_at
	`

	var member models.TeamMember
	err := r.pool.QueryRow(ctx, query, teamID, userID, role).Scan(
		&member.ID,
		&member.TeamID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewNotFoundError("team member")
		}
		return nil, apperrors.NewDatabaseError("failed to update member role", err)
	}

	return &member, nil
}

// RemoveMember removes a member from a team.
func (r *TeamRepository) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	query := `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`

	result, err := r.pool.Exec(ctx, query, teamID, userID)
	if err != nil {
		return apperrors.NewDatabaseError("failed to remove team member", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NewNotFoundError("team member")
	}

	return nil
}

// GetMembers retrieves all members of a team.
func (r *TeamRepository) GetMembers(ctx context.Context, teamID uuid.UUID) ([]models.TeamMember, error) {
	query := `
		SELECT tm.id, tm.team_id, tm.user_id, tm.role, tm.joined_at,
			   u.id, u.email, u.name, u.avatar_url, u.created_at, u.updated_at
		FROM team_members tm
		INNER JOIN users u ON tm.user_id = u.id
		WHERE tm.team_id = $1
		ORDER BY tm.joined_at ASC
	`

	rows, err := r.pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to get team members", err)
	}
	defer rows.Close()

	var members []models.TeamMember
	for rows.Next() {
		var member models.TeamMember
		var user models.User
		if err := rows.Scan(
			&member.ID,
			&member.TeamID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
			&user.ID,
			&user.Email,
			&user.Name,
			&user.AvatarURL,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, apperrors.NewDatabaseError("failed to scan team member", err)
		}
		member.User = &user
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.NewDatabaseError("error iterating team members", err)
	}

	return members, nil
}

// GetMember retrieves a specific member of a team.
func (r *TeamRepository) GetMember(ctx context.Context, teamID, userID uuid.UUID) (*models.TeamMember, error) {
	query := `
		SELECT tm.id, tm.team_id, tm.user_id, tm.role, tm.joined_at,
			   u.id, u.email, u.name, u.avatar_url, u.created_at, u.updated_at
		FROM team_members tm
		INNER JOIN users u ON tm.user_id = u.id
		WHERE tm.team_id = $1 AND tm.user_id = $2
	`

	var member models.TeamMember
	var user models.User
	err := r.pool.QueryRow(ctx, query, teamID, userID).Scan(
		&member.ID,
		&member.TeamID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
		&user.ID,
		&user.Email,
		&user.Name,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewNotFoundError("team member")
		}
		return nil, apperrors.NewDatabaseError("failed to get team member", err)
	}
	member.User = &user

	return &member, nil
}

// IsMember checks if a user is a member of a team.
func (r *TeamRepository) IsMember(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, teamID, userID).Scan(&exists)
	if err != nil {
		return false, apperrors.NewDatabaseError("failed to check membership", err)
	}

	return exists, nil
}
