package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/aleksandr/strive-api/internal/models"
	"github.com/aleksandr/strive-api/internal/repositories"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidTheme = errors.New("invalid theme value")
)

type UserService interface {
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateUserTheme(ctx context.Context, userID uuid.UUID, theme string) error
}

type userService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return user, nil
}

func (s *userService) UpdateUserTheme(ctx context.Context, userID uuid.UUID, theme string) error {
	if theme != "light" && theme != "dark" {
		return ErrInvalidTheme
	}

	if err := s.userRepo.UpdateTheme(ctx, userID, theme); err != nil {
		return fmt.Errorf("failed to update user theme: %w", err)
	}

	return nil
}
