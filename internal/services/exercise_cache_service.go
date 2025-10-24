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
	exerciseRepo repositories.ExerciseRepository
	wgerClient   *clients.WgerClient
	logger       *logger.Logger
	cacheTTL     time.Duration
}

func NewExerciseCacheService(
	exerciseRepo repositories.ExerciseRepository,
	wgerClient *clients.WgerClient,
	logger *logger.Logger,
	cacheTTL time.Duration,
) ExerciseCacheService {
	return &exerciseCacheService{
		exerciseRepo: exerciseRepo,
		wgerClient:   wgerClient,
		logger:       logger,
		cacheTTL:     cacheTTL,
	}
}

func (s *exerciseCacheService) RefreshCache(ctx context.Context) error {
	s.logger.Info("Starting cache refresh from wger API")

	s.logger.Info("Step 1: Clearing existing cache")
	if err := s.exerciseRepo.ClearCache(ctx); err != nil {
		s.logger.Error("Failed to clear existing cache", "error", err)
		return fmt.Errorf("failed to clear existing cache: %w", err)
	}
	s.logger.Info("Successfully cleared existing cache")

	s.logger.Info("Step 2: Caching muscle groups")
	if err := s.cacheMuscleGroups(ctx); err != nil {
		s.logger.Error("Failed to cache muscle groups", "error", err)
		return fmt.Errorf("failed to cache muscle groups: %w", err)
	}
	s.logger.Info("Successfully cached muscle groups")

	s.logger.Info("Step 3: Caching equipment")
	if err := s.cacheEquipment(ctx); err != nil {
		s.logger.Error("Failed to cache equipment", "error", err)
		return fmt.Errorf("failed to cache equipment: %w", err)
	}
	s.logger.Info("Successfully cached equipment")

	s.logger.Info("Step 4: Caching exercises")
	if err := s.cacheExercises(ctx); err != nil {
		s.logger.Error("Failed to cache exercises", "error", err)
		return fmt.Errorf("failed to cache exercises: %w", err)
	}
	s.logger.Info("Successfully cached exercises")

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
	s.logger.Debug("Caching muscle groups from wger API")

	muscles, err := s.wgerClient.GetMuscles(ctx)
	if err != nil {
		return fmt.Errorf("failed to get muscles from wger API: %w", err)
	}

	for _, muscle := range muscles {
		muscleGroup := &models.MuscleGroup{
			ID:        uuid.New(),
			WgerID:    muscle.ID,
			Name:      muscle.Name,
			NameEn:    muscle.NameEn,
			IsFront:   muscle.IsFront,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := s.exerciseRepo.SaveMuscleGroup(ctx, muscleGroup); err != nil {
			s.logger.Error("Failed to save muscle group", "error", err, "muscle_group", muscle.Name)
			continue
		}
	}

	s.logger.Info("Successfully cached muscle groups", "count", len(muscles))
	return nil
}

func (s *exerciseCacheService) cacheEquipment(ctx context.Context) error {
	s.logger.Debug("Caching equipment from wger API")

	equipment, err := s.wgerClient.GetEquipment(ctx)
	if err != nil {
		return fmt.Errorf("failed to get equipment from wger API: %w", err)
	}

	for _, eq := range equipment {
		equipmentModel := &models.Equipment{
			ID:        uuid.New(),
			WgerID:    eq.ID,
			Name:      eq.Name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
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
	s.logger.Info("Starting to cache exercises from wger API with pagination")

	wgerCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	s.logger.Info("Created new context for wger API", "timeout", "10m")
	allExercises, err := s.wgerClient.GetAllExercises(wgerCtx)
	if err != nil {
		s.logger.Error("Failed to get exercises from wger API", "error", err)
		return fmt.Errorf("failed to get exercises from wger API: %w", err)
	}

	s.logger.Info("Successfully received exercises from wger API", "count", len(allExercises))

	now := time.Now()
	expiresAt := now.Add(s.cacheTTL)

	for i := range allExercises {
		ex := &allExercises[i]
		exercise := &models.Exercise{
			ID:            uuid.New(),
			WgerID:        ex.ID,
			WgerUUID:      ex.UUID,
			Name:          ex.Name,
			Description:   ex.Description,
			Category:      ex.Category.ID,
			Language:      ex.Language,
			License:       ex.License.ID,
			LicenseAuthor: ex.LicenseAuthor,
			CreationDate:  parseDate(ex.CreationDate),
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

		variations := ex.GetVariations()
		if len(variations) > 0 {
			s.cacheExerciseVariations(ctx, exercise.ID, variations)
		}
	}

	s.logger.Info("Successfully cached exercises", "count", len(allExercises))
	return nil
}

func (s *exerciseCacheService) cacheExerciseMuscleGroups(
	ctx context.Context,
	exerciseID uuid.UUID,
	primaryMuscles, secondaryMuscles []models.WgerMuscle,
) error {
	muscleGroups, err := s.exerciseRepo.GetMuscleGroups(ctx)
	if err != nil {
		return fmt.Errorf("failed to get muscle groups: %w", err)
	}

	muscleGroupMap := make(map[int]uuid.UUID)
	for _, mg := range muscleGroups {
		muscleGroupMap[mg.WgerID] = mg.ID
	}

	for _, muscle := range primaryMuscles {
		if muscleGroupID, exists := muscleGroupMap[muscle.ID]; exists {
			if err := s.exerciseRepo.SaveExerciseMuscleGroup(ctx, exerciseID, muscleGroupID, true); err != nil {
				s.logger.Warn("Failed to save primary muscle group", "error", err, "muscle_id", muscle.ID)
			}
		}
	}

	for _, muscle := range secondaryMuscles {
		if muscleGroupID, exists := muscleGroupMap[muscle.ID]; exists {
			if err := s.exerciseRepo.SaveExerciseMuscleGroup(ctx, exerciseID, muscleGroupID, false); err != nil {
				s.logger.Warn("Failed to save secondary muscle group", "error", err, "muscle_id", muscle.ID)
			}
		}
	}

	return nil
}

func (s *exerciseCacheService) cacheExerciseEquipment(ctx context.Context, exerciseID uuid.UUID, equipment []models.WgerEquipment) error {
	equipmentList, err := s.exerciseRepo.GetEquipment(ctx)
	if err != nil {
		return fmt.Errorf("failed to get equipment: %w", err)
	}

	equipmentMap := make(map[int]uuid.UUID)
	for _, eq := range equipmentList {
		equipmentMap[eq.WgerID] = eq.ID
	}

	for _, eq := range equipment {
		if eqID, exists := equipmentMap[eq.ID]; exists {
			if err := s.exerciseRepo.SaveExerciseEquipment(ctx, exerciseID, eqID); err != nil {
				s.logger.Warn("Failed to save exercise equipment", "error", err, "equipment_id", eq.ID)
			}
		}
	}

	return nil
}

func (s *exerciseCacheService) cacheExerciseVariations(ctx context.Context, exerciseID uuid.UUID, variationIDs []int) {
	for _, variationWgerID := range variationIDs {
		variation, err := s.exerciseRepo.GetByWgerID(ctx, variationWgerID)
		if err != nil {
			s.logger.Warn("Variation not found in cache", "wger_id", variationWgerID)
			continue
		}

		if err := s.exerciseRepo.SaveExerciseVariation(ctx, exerciseID, variation.ID); err != nil {
			s.logger.Warn("Failed to save exercise variation", "error", err, "variation_id", variationWgerID)
		}
	}
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
