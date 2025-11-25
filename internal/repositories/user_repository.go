// Package repositories provides data access layer implementations.
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

// UserRepository handles user data access operations.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create creates a new user or updates if email exists (upsert).
func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (email, name, avatar_url)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE SET
			name = EXCLUDED.name,
			avatar_url = COALESCE(EXCLUDED.avatar_url, users.avatar_url),
			updated_at = NOW()
		RETURNING id, email, name, avatar_url, created_at, updated_at
	`

	var result models.User
	err := r.pool.QueryRow(ctx, query, user.Email, user.Name, user.AvatarURL).Scan(
		&result.ID,
		&result.Email,
		&result.Name,
		&result.AvatarURL,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to create user", err)
	}

	return &result, nil
}

// GetByID retrieves a user by their ID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, name, avatar_url, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewNotFoundError("user")
		}
		return nil, apperrors.NewDatabaseError("failed to get user", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by their email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, name, avatar_url, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewNotFoundError("user")
		}
		return nil, apperrors.NewDatabaseError("failed to get user by email", err)
	}

	return &user, nil
}

// Update updates a user's information.
func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, name *string, avatarURL *string) (*models.User, error) {
	query := `
		UPDATE users
		SET name = COALESCE($2, name),
			avatar_url = COALESCE($3, avatar_url),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, name, avatar_url, created_at, updated_at
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, id, name, avatarURL).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewNotFoundError("user")
		}
		return nil, apperrors.NewDatabaseError("failed to update user", err)
	}

	return &user, nil
}

// Delete deletes a user by ID.
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return apperrors.NewDatabaseError("failed to delete user", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NewNotFoundError("user")
	}

	return nil
}

// GetByIDs retrieves multiple users by their IDs.
func (r *UserRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.User, error) {
	if len(ids) == 0 {
		return []models.User{}, nil
	}

	query := `
		SELECT id, email, name, avatar_url, created_at, updated_at
		FROM users
		WHERE id = ANY($1)
	`

	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, apperrors.NewDatabaseError("failed to get users", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.AvatarURL,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, apperrors.NewDatabaseError("failed to scan user", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.NewDatabaseError("error iterating users", err)
	}

	return users, nil
}
