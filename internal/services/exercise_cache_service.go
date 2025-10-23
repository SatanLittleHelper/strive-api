package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aleksandr/strive-api/internal/clients"
	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/models"
	"github.com/aleksandr/strive-api/internal/repositories"
	"github.com/google/uuid"
)

type ExerciseCacheService interface {
	RefreshCache(ctx context.Context) error
	IsCacheValid(ctx context.Context) (bool, error)
	GetCacheStatus(ctx context.Context) (*models.CacheStatus, error)
	ClearCache(ctx context.Context) error
}

type exerciseCacheService struct {
	exerciseRepo     repositories.ExerciseRepository
	exerciseDBClient *clients.ExerciseDBClient
	logger           *logger.Logger
	cacheTTL         time.Duration
}

func NewExerciseCacheService(
	exerciseRepo repositories.ExerciseRepository,
	exerciseDBClient *clients.ExerciseDBClient,
	logger *logger.Logger,
	cacheTTL time.Duration,
) ExerciseCacheService {
	return &exerciseCacheService{
		exerciseRepo:     exerciseRepo,
		exerciseDBClient: exerciseDBClient,
		logger:           logger,
		cacheTTL:         cacheTTL,
	}
}

func (s *exerciseCacheService) RefreshCache(ctx context.Context) error {
	s.logger.Info("Starting cache refresh from ExerciseDB")

	if err := s.exerciseRepo.ClearCache(ctx); err != nil {
		return fmt.Errorf("failed to clear existing cache: %w", err)
	}

	if err := s.cacheMuscleGroups(ctx); err != nil {
		return fmt.Errorf("failed to cache muscle groups: %w", err)
	}

	if err := s.cacheEquipment(ctx); err != nil {
		return fmt.Errorf("failed to cache equipment: %w", err)
	}

	if err := s.cacheExercises(ctx); err != nil {
		return fmt.Errorf("failed to cache exercises: %w", err)
	}

	s.logger.Info("Cache refresh completed successfully")
	return nil
}

func (s *exerciseCacheService) IsCacheValid(ctx context.Context) (bool, error) {
	return s.exerciseRepo.IsCacheValid(ctx)
}

func (s *exerciseCacheService) GetCacheStatus(ctx context.Context) (*models.CacheStatus, error) {
	return s.exerciseRepo.GetCacheStatus(ctx)
}

func (s *exerciseCacheService) ClearCache(ctx context.Context) error {
	s.logger.Info("Clearing exercise cache")
	return s.exerciseRepo.ClearCache(ctx)
}

func (s *exerciseCacheService) cacheMuscleGroups(ctx context.Context) error {
	s.logger.Debug("Caching muscle groups from ExerciseDB")

	muscleGroups, err := s.exerciseDBClient.GetMuscleGroups(ctx)
	if err != nil {
		return fmt.Errorf("failed to get muscle groups from ExerciseDB: %w", err)
	}

	for _, mg := range muscleGroups {
		muscleGroup := &models.MuscleGroup{
			ID:           uuid.New(),
			ExerciseDBID: mg.ID,
			Name:         mg.Name,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := s.exerciseRepo.SaveMuscleGroup(ctx, muscleGroup); err != nil {
			s.logger.Error("Failed to save muscle group", "error", err, "muscle_group", mg.Name)
			continue
		}
	}

	s.logger.Info("Successfully cached muscle groups", "count", len(muscleGroups))
	return nil
}

func (s *exerciseCacheService) cacheEquipment(ctx context.Context) error {
	s.logger.Debug("Caching equipment from ExerciseDB")

	equipment, err := s.exerciseDBClient.GetEquipment(ctx)
	if err != nil {
		return fmt.Errorf("failed to get equipment from ExerciseDB: %w", err)
	}

	for _, eq := range equipment {
		equipmentModel := &models.Equipment{
			ID:           uuid.New(),
			ExerciseDBID: eq.ID,
			Name:         eq.Name,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := s.exerciseRepo.SaveEquipment(ctx, equipmentModel); err != nil {
			s.logger.Error("Failed to save equipment", "error", err, "equipment", eq.Name)
			continue
		}
	}

	s.logger.Info("Successfully cached equipment", "count", len(equipment))
	return nil
}

func (s *exerciseCacheService) cacheExercises(ctx context.Context) error {
	s.logger.Debug("Caching exercises from ExerciseDB")

	exercises, err := s.exerciseDBClient.GetAllExercises(ctx)
	if err != nil {
		return fmt.Errorf("failed to get exercises from ExerciseDB: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(s.cacheTTL)

	for i := range exercises {
		ex := &exercises[i]
		exercise := &models.Exercise{
			ID:            uuid.New(),
			ExerciseDBID:  ex.ID,
			Name:          ex.Name,
			Description:   stringPtr(ex.Description),
			Instructions:  stringPtr(ex.Description),
			Tips:          stringPtr(ex.Description),
			Category:      intPtr(ex.Category),
			Language:      intPtr(ex.Language),
			License:       intPtr(ex.License),
			LicenseAuthor: stringPtr(ex.LicenseAuthor),
			Status:        stringPtr(ex.Status),
			NameOriginal:  stringPtr(ex.NameOriginal),
			CreationDate:  parseDate(ex.CreationDate),
			UUID:          stringPtr(ex.UUID),
			CachedAt:      now,
			ExpiresAt:     expiresAt,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := s.exerciseRepo.SaveExercise(ctx, exercise); err != nil {
			s.logger.Error("Failed to save exercise", "error", err, "exercise", ex.Name)
			continue
		}

		if err := s.cacheExerciseMuscleGroups(ctx, exercise.ID, ex.Muscles, ex.MusclesSecondary); err != nil {
			s.logger.Error("Failed to cache exercise muscle groups", "error", err, "exercise", ex.Name)
		}

		if err := s.cacheExerciseEquipment(ctx, exercise.ID, ex.Equipment); err != nil {
			s.logger.Error("Failed to cache exercise equipment", "error", err, "exercise", ex.Name)
		}
	}

	s.logger.Info("Successfully cached exercises", "count", len(exercises))
	return nil
}

func (s *exerciseCacheService) cacheExerciseMuscleGroups(
	ctx context.Context,
	exerciseID uuid.UUID,
	primaryMuscles, secondaryMuscles []int,
) error {
	muscleGroups, err := s.exerciseRepo.GetMuscleGroups(ctx)
	if err != nil {
		return fmt.Errorf("failed to get muscle groups: %w", err)
	}

	muscleGroupMap := make(map[int]uuid.UUID)
	for _, mg := range muscleGroups {
		muscleGroupMap[mg.ExerciseDBID] = mg.ID
	}

	for _, muscleID := range primaryMuscles {
		if muscleGroupID, exists := muscleGroupMap[muscleID]; exists {
			if err := s.exerciseRepo.SaveExerciseMuscleGroup(ctx, exerciseID, muscleGroupID, true); err != nil {
				s.logger.Warn("Failed to save primary muscle group", "error", err, "muscle_id", muscleID)
			}
		}
	}

	for _, muscleID := range secondaryMuscles {
		if muscleGroupID, exists := muscleGroupMap[muscleID]; exists {
			if err := s.exerciseRepo.SaveExerciseMuscleGroup(ctx, exerciseID, muscleGroupID, false); err != nil {
				s.logger.Warn("Failed to save secondary muscle group", "error", err, "muscle_id", muscleID)
			}
		}
	}

	return nil
}

func (s *exerciseCacheService) cacheExerciseEquipment(ctx context.Context, exerciseID uuid.UUID, equipmentIDs []int) error {
	equipment, err := s.exerciseRepo.GetEquipment(ctx)
	if err != nil {
		return fmt.Errorf("failed to get equipment: %w", err)
	}

	equipmentMap := make(map[int]uuid.UUID)
	for _, eq := range equipment {
		equipmentMap[eq.ExerciseDBID] = eq.ID
	}

	for _, equipmentID := range equipmentIDs {
		if eqID, exists := equipmentMap[equipmentID]; exists {
			if err := s.exerciseRepo.SaveExerciseEquipment(ctx, exerciseID, eqID); err != nil {
				s.logger.Warn("Failed to save exercise equipment", "error", err, "equipment_id", equipmentID)
			}
		}
	}

	return nil
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func intPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

func parseDate(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}

	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil
	}

	return &parsed
}
