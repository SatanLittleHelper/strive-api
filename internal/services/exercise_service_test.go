package services

import (
	"context"
	"testing"
	"time"

	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockExerciseRepository struct {
	mock.Mock
}

func (m *MockExerciseRepository) GetAll(ctx context.Context, filters *models.ExerciseFilters) (*models.ExerciseListResponse, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).(*models.ExerciseListResponse), args.Error(1)
}

func (m *MockExerciseRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Exercise, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Exercise), args.Error(1)
}

func (m *MockExerciseRepository) GetByExerciseDBID(ctx context.Context, exerciseDBID int) (*models.Exercise, error) {
	args := m.Called(ctx, exerciseDBID)
	return args.Get(0).(*models.Exercise), args.Error(1)
}

func (m *MockExerciseRepository) GetMuscleGroups(ctx context.Context) ([]models.MuscleGroup, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.MuscleGroup), args.Error(1)
}

func (m *MockExerciseRepository) GetEquipment(ctx context.Context) ([]models.Equipment, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Equipment), args.Error(1)
}

func (m *MockExerciseRepository) GetCacheStatus(ctx context.Context) (*models.CacheStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*models.CacheStatus), args.Error(1)
}

func (m *MockExerciseRepository) IsCacheValid(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockExerciseRepository) ClearCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockExerciseRepository) SaveExercise(ctx context.Context, exercise *models.Exercise) error {
	args := m.Called(ctx, exercise)
	return args.Error(0)
}

func (m *MockExerciseRepository) SaveMuscleGroup(ctx context.Context, muscleGroup *models.MuscleGroup) error {
	args := m.Called(ctx, muscleGroup)
	return args.Error(0)
}

func (m *MockExerciseRepository) SaveEquipment(ctx context.Context, equipment *models.Equipment) error {
	args := m.Called(ctx, equipment)
	return args.Error(0)
}

func (m *MockExerciseRepository) SaveExerciseMuscleGroup(ctx context.Context, exerciseID, muscleGroupID uuid.UUID, isPrimary bool) error {
	args := m.Called(ctx, exerciseID, muscleGroupID, isPrimary)
	return args.Error(0)
}

func (m *MockExerciseRepository) SaveExerciseEquipment(ctx context.Context, exerciseID, equipmentID uuid.UUID) error {
	args := m.Called(ctx, exerciseID, equipmentID)
	return args.Error(0)
}

func (m *MockExerciseRepository) SaveExerciseAlternative(ctx context.Context, exerciseID, alternativeID uuid.UUID) error {
	args := m.Called(ctx, exerciseID, alternativeID)
	return args.Error(0)
}

type MockExerciseCacheService struct {
	mock.Mock
}

func (m *MockExerciseCacheService) RefreshCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockExerciseCacheService) IsCacheValid(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockExerciseCacheService) GetCacheStatus(ctx context.Context) (*models.CacheStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*models.CacheStatus), args.Error(1)
}

func (m *MockExerciseCacheService) ClearCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestExerciseService_GetExercises(t *testing.T) {
	mockRepo := &MockExerciseRepository{}
	mockCache := &MockExerciseCacheService{}
	logger := logger.New("DEBUG", "text")

	service := NewExerciseService(mockRepo, mockCache, logger)

	ctx := context.Background()
	filters := &models.ExerciseFilters{
		Page:  1,
		Limit: 10,
	}

	expectedResponse := &models.ExerciseListResponse{
		Exercises: []models.Exercise{
			{
				ID:   uuid.New(),
				Name: "Push-up",
			},
		},
		Total: 1,
		Page:  1,
		Limit: 10,
	}

	mockCache.On("IsCacheValid", ctx).Return(true, nil)
	mockRepo.On("GetAll", ctx, filters).Return(expectedResponse, nil)

	response, err := service.GetExercises(ctx, filters)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestExerciseService_GetExercises_InvalidCache(t *testing.T) {
	mockRepo := &MockExerciseRepository{}
	mockCache := &MockExerciseCacheService{}
	logger := logger.New("DEBUG", "text")

	service := NewExerciseService(mockRepo, mockCache, logger)

	ctx := context.Background()
	filters := &models.ExerciseFilters{
		Page:  1,
		Limit: 10,
	}

	expectedResponse := &models.ExerciseListResponse{
		Exercises: []models.Exercise{
			{
				ID:   uuid.New(),
				Name: "Push-up",
			},
		},
		Total: 1,
		Page:  1,
		Limit: 10,
	}

	mockCache.On("IsCacheValid", ctx).Return(false, nil)
	mockCache.On("RefreshCache", ctx).Return(nil)
	mockRepo.On("GetAll", ctx, filters).Return(expectedResponse, nil)

	response, err := service.GetExercises(ctx, filters)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestExerciseService_GetExerciseByID(t *testing.T) {
	mockRepo := &MockExerciseRepository{}
	mockCache := &MockExerciseCacheService{}
	logger := logger.New("DEBUG", "text")

	service := NewExerciseService(mockRepo, mockCache, logger)

	ctx := context.Background()
	exerciseID := uuid.New()

	expectedExercise := &models.Exercise{
		ID:   exerciseID,
		Name: "Push-up",
	}

	mockCache.On("IsCacheValid", ctx).Return(true, nil)
	mockRepo.On("GetByID", ctx, exerciseID).Return(expectedExercise, nil)

	exercise, err := service.GetExerciseByID(ctx, exerciseID)

	assert.NoError(t, err)
	assert.Equal(t, expectedExercise, exercise)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestExerciseService_GetMuscleGroups(t *testing.T) {
	mockRepo := &MockExerciseRepository{}
	mockCache := &MockExerciseCacheService{}
	logger := logger.New("DEBUG", "text")

	service := NewExerciseService(mockRepo, mockCache, logger)

	ctx := context.Background()

	expectedMuscleGroups := []models.MuscleGroup{
		{
			ID:   uuid.New(),
			Name: "Chest",
		},
	}

	mockCache.On("IsCacheValid", ctx).Return(true, nil)
	mockRepo.On("GetMuscleGroups", ctx).Return(expectedMuscleGroups, nil)

	muscleGroups, err := service.GetMuscleGroups(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedMuscleGroups, muscleGroups)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestExerciseService_GetEquipment(t *testing.T) {
	mockRepo := &MockExerciseRepository{}
	mockCache := &MockExerciseCacheService{}
	logger := logger.New("DEBUG", "text")

	service := NewExerciseService(mockRepo, mockCache, logger)

	ctx := context.Background()

	expectedEquipment := []models.Equipment{
		{
			ID:   uuid.New(),
			Name: "Dumbbells",
		},
	}

	mockCache.On("IsCacheValid", ctx).Return(true, nil)
	mockRepo.On("GetEquipment", ctx).Return(expectedEquipment, nil)

	equipment, err := service.GetEquipment(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedEquipment, equipment)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestExerciseService_GetCacheStatus(t *testing.T) {
	mockRepo := &MockExerciseRepository{}
	mockCache := &MockExerciseCacheService{}
	logger := logger.New("DEBUG", "text")

	service := NewExerciseService(mockRepo, mockCache, logger)

	ctx := context.Background()

	expectedStatus := &models.CacheStatus{
		LastUpdated:    time.Now(),
		TotalExercises: 100,
		TotalMuscles:   14,
		TotalEquipment: 8,
		IsValid:        true,
		ExpiresAt:      time.Now().Add(24 * time.Hour),
	}

	mockCache.On("GetCacheStatus", ctx).Return(expectedStatus, nil)

	status, err := service.GetCacheStatus(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedStatus, status)
	mockCache.AssertExpectations(t)
}

func TestExerciseService_RefreshCache(t *testing.T) {
	mockRepo := &MockExerciseRepository{}
	mockCache := &MockExerciseCacheService{}
	logger := logger.New("DEBUG", "text")

	service := NewExerciseService(mockRepo, mockCache, logger)

	ctx := context.Background()

	mockCache.On("RefreshCache", ctx).Return(nil)

	err := service.RefreshCache(ctx)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}
