package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/models"
	"github.com/aleksandr/strive-api/internal/repositories"
	"github.com/google/uuid"
)

var (
	ErrExerciseNotFound = errors.New("exercise not found")
	ErrCacheNotValid    = errors.New("exercise cache is not valid")
)

type ExerciseService interface {
	GetExercises(ctx context.Context, filters *models.ExerciseFilters) (*models.ExerciseListResponse, error)
	GetExerciseByID(ctx context.Context, id uuid.UUID) (*models.Exercise, error)
	GetMuscleGroups(ctx context.Context) ([]models.MuscleGroup, error)
	GetEquipment(ctx context.Context) ([]models.Equipment, error)
	GetCacheStatus(ctx context.Context) (*models.CacheStatus, error)
	RefreshCache(ctx context.Context) error
}

type exerciseService struct {
	exerciseRepo repositories.ExerciseRepository
	cacheService ExerciseCacheService
	logger       *logger.Logger
}

func NewExerciseService(
	exerciseRepo repositories.ExerciseRepository,
	cacheService ExerciseCacheService,
	logger *logger.Logger,
) ExerciseService {
	return &exerciseService{
		exerciseRepo: exerciseRepo,
		cacheService: cacheService,
		logger:       logger,
	}
}

func (s *exerciseService) GetExercises(ctx context.Context, filters *models.ExerciseFilters) (*models.ExerciseListResponse, error) {
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	isValid, err := s.cacheService.IsCacheValid(ctx)
	if err != nil {
		s.logger.Error("Failed to check cache validity", "error", err)
		return nil, fmt.Errorf("failed to check cache validity: %w", err)
	}

	if !isValid {
		s.logger.Warn("Cache is not valid, attempting to refresh")
		if err := s.cacheService.RefreshCache(ctx); err != nil {
			s.logger.Error("Failed to refresh cache", "error", err)
			return nil, fmt.Errorf("failed to refresh cache: %w", err)
		}
	}

	response, err := s.exerciseRepo.GetAll(ctx, filters)
	if err != nil {
		s.logger.Error("Failed to get exercises", "error", err)
		return nil, fmt.Errorf("failed to get exercises: %w", err)
	}

	s.logger.Debug("Successfully retrieved exercises", "count", len(response.Exercises), "total", response.Total)
	return response, nil
}

func (s *exerciseService) GetExerciseByID(ctx context.Context, id uuid.UUID) (*models.Exercise, error) {
	isValid, err := s.cacheService.IsCacheValid(ctx)
	if err != nil {
		s.logger.Error("Failed to check cache validity", "error", err)
		return nil, fmt.Errorf("failed to check cache validity: %w", err)
	}

	if !isValid {
		s.logger.Warn("Cache is not valid, attempting to refresh")
		if err := s.cacheService.RefreshCache(ctx); err != nil {
			s.logger.Error("Failed to refresh cache", "error", err)
			return nil, fmt.Errorf("failed to refresh cache: %w", err)
		}
	}

	exercise, err := s.exerciseRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get exercise by ID", "error", err, "exercise_id", id)
		return nil, fmt.Errorf("failed to get exercise: %w", err)
	}

	s.logger.Debug("Successfully retrieved exercise", "exercise_id", id, "name", exercise.Name)
	return exercise, nil
}

func (s *exerciseService) GetMuscleGroups(ctx context.Context) ([]models.MuscleGroup, error) {
	isValid, err := s.cacheService.IsCacheValid(ctx)
	if err != nil {
		s.logger.Error("Failed to check cache validity", "error", err)
		return nil, fmt.Errorf("failed to check cache validity: %w", err)
	}

	if !isValid {
		s.logger.Warn("Cache is not valid, attempting to refresh")
		if err := s.cacheService.RefreshCache(ctx); err != nil {
			s.logger.Error("Failed to refresh cache", "error", err)
			return nil, fmt.Errorf("failed to refresh cache: %w", err)
		}
	}

	muscleGroups, err := s.exerciseRepo.GetMuscleGroups(ctx)
	if err != nil {
		s.logger.Error("Failed to get muscle groups", "error", err)
		return nil, fmt.Errorf("failed to get muscle groups: %w", err)
	}

	s.logger.Debug("Successfully retrieved muscle groups", "count", len(muscleGroups))
	return muscleGroups, nil
}

func (s *exerciseService) GetEquipment(ctx context.Context) ([]models.Equipment, error) {
	isValid, err := s.cacheService.IsCacheValid(ctx)
	if err != nil {
		s.logger.Error("Failed to check cache validity", "error", err)
		return nil, fmt.Errorf("failed to check cache validity: %w", err)
	}

	if !isValid {
		s.logger.Warn("Cache is not valid, attempting to refresh")
		if err := s.cacheService.RefreshCache(ctx); err != nil {
			s.logger.Error("Failed to refresh cache", "error", err)
			return nil, fmt.Errorf("failed to refresh cache: %w", err)
		}
	}

	equipment, err := s.exerciseRepo.GetEquipment(ctx)
	if err != nil {
		s.logger.Error("Failed to get equipment", "error", err)
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}

	s.logger.Debug("Successfully retrieved equipment", "count", len(equipment))
	return equipment, nil
}

func (s *exerciseService) GetCacheStatus(ctx context.Context) (*models.CacheStatus, error) {
	status, err := s.cacheService.GetCacheStatus(ctx)
	if err != nil {
		s.logger.Error("Failed to get cache status", "error", err)
		return nil, fmt.Errorf("failed to get cache status: %w", err)
	}

	return status, nil
}

func (s *exerciseService) RefreshCache(ctx context.Context) error {
	s.logger.Info("Manual cache refresh requested")

	if err := s.cacheService.RefreshCache(ctx); err != nil {
		s.logger.Error("Failed to refresh cache", "error", err)
		return fmt.Errorf("failed to refresh cache: %w", err)
	}

	s.logger.Info("Cache refresh completed successfully")
	return nil
}
