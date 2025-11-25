// Package services implements the business logic layer.
package services

import (
	"context"

	"github.com/google/uuid"
	"taskboard-backend/internal/dto"
	"taskboard-backend/internal/models"
	"taskboard-backend/internal/repositories"
)

// UserService handles user business logic.
type UserService struct {
	userRepo *repositories.UserRepository
}

// NewUserService creates a new UserService instance.
func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// CreateOrUpdate creates a new user or updates if email exists.
func (s *UserService) CreateOrUpdate(ctx context.Context, req *dto.CreateUserRequest) (*models.User, error) {
	user := &models.User{
		Email:     req.Email,
		Name:      req.Name,
		AvatarURL: req.AvatarURL,
	}

	return s.userRepo.Create(ctx, user)
}

// GetByID retrieves a user by their ID.
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// GetByEmail retrieves a user by their email.
func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

// Update updates a user's information.
func (s *UserService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateUserRequest) (*models.User, error) {
	return s.userRepo.Update(ctx, id, req.Name, req.AvatarURL)
}

// Delete deletes a user.
func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}

// GetByIDs retrieves multiple users by their IDs.
func (s *UserService) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.User, error) {
	return s.userRepo.GetByIDs(ctx, ids)
}
