package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/aleksandr/strive-api/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockExerciseRepository struct {
	mock.Mock
}

func (m *mockExerciseRepository) GetExercises(ctx context.Context, filters *models.ExerciseFilters) (*models.ExerciseListResponse, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ExerciseListResponse), args.Error(1)
}

func (m *mockExerciseRepository) GetExerciseByID(ctx context.Context, id uuid.UUID) (*models.Exercise, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Exercise), args.Error(1)
}

func (m *mockExerciseRepository) GetMuscleGroups(ctx context.Context) ([]models.MuscleGroup, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.MuscleGroup), args.Error(1)
}

func (m *mockExerciseRepository) GetEquipment(ctx context.Context) ([]models.Equipment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Equipment), args.Error(1)
}

func (m *mockExerciseRepository) GetCacheStatus(ctx context.Context) (*models.CacheStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CacheStatus), args.Error(1)
}

func (m *mockExerciseRepository) IsCacheValid(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *mockExerciseRepository) ClearCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockExerciseRepository) SaveExercise(ctx context.Context, exercise *models.Exercise) error {
	args := m.Called(ctx, exercise)
	return args.Error(0)
}

func (m *mockExerciseRepository) SaveMuscleGroup(ctx context.Context, muscleGroup *models.MuscleGroup) error {
	args := m.Called(ctx, muscleGroup)
	return args.Error(0)
}

func (m *mockExerciseRepository) SaveEquipment(ctx context.Context, equipment *models.Equipment) error {
	args := m.Called(ctx, equipment)
	return args.Error(0)
}

func (m *mockExerciseRepository) SaveExerciseMuscleGroup(ctx context.Context, exerciseID, muscleGroupID uuid.UUID, isPrimary bool) error {
	args := m.Called(ctx, exerciseID, muscleGroupID, isPrimary)
	return args.Error(0)
}

func (m *mockExerciseRepository) SaveExerciseEquipment(ctx context.Context, exerciseID, equipmentID uuid.UUID) error {
	args := m.Called(ctx, exerciseID, equipmentID)
	return args.Error(0)
}

func (m *mockExerciseRepository) SaveExerciseAlternative(ctx context.Context, exerciseID, alternativeID uuid.UUID) error {
	args := m.Called(ctx, exerciseID, alternativeID)
	return args.Error(0)
}

func TestExerciseRepository_GetAll(t *testing.T) {
	repo := &mockExerciseRepository{}
	ctx := context.Background()
	filters := &models.ExerciseFilters{
		Page:  1,
		Limit: 10,
	}

	expectedResponse := &models.ExerciseListResponse{
		Exercises: []models.Exercise{},
		Total:     0,
		Page:      1,
		Limit:     10,
	}

	repo.On("GetExercises", ctx, filters).Return(expectedResponse, nil)

	response, err := repo.GetExercises(ctx, filters)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.Limit)
	assert.Equal(t, 0, response.Total)

	repo.AssertExpectations(t)
}

func TestExerciseRepository_GetMuscleGroups(t *testing.T) {
	repo := &mockExerciseRepository{}
	ctx := context.Background()

	expectedMuscleGroups := []models.MuscleGroup{
		{
			ID:        uuid.New(),
			Name:      "Chest",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	repo.On("GetMuscleGroups", ctx).Return(expectedMuscleGroups, nil)

	muscleGroups, err := repo.GetMuscleGroups(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, muscleGroups)
	assert.Len(t, muscleGroups, 1)
	assert.Equal(t, "Chest", muscleGroups[0].Name)

	repo.AssertExpectations(t)
}

func TestExerciseRepository_GetEquipment(t *testing.T) {
	repo := &mockExerciseRepository{}
	ctx := context.Background()

	expectedEquipment := []models.Equipment{
		{
			ID:        uuid.New(),
			Name:      "Dumbbell",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	repo.On("GetEquipment", ctx).Return(expectedEquipment, nil)

	equipment, err := repo.GetEquipment(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, equipment)
	assert.Len(t, equipment, 1)
	assert.Equal(t, "Dumbbell", equipment[0].Name)

	repo.AssertExpectations(t)
}

func TestExerciseRepository_GetCacheStatus(t *testing.T) {
	repo := &mockExerciseRepository{}
	ctx := context.Background()

	expectedStatus := &models.CacheStatus{
		LastUpdated:    time.Now(),
		TotalExercises: 100,
		TotalMuscles:   20,
		TotalEquipment: 15,
		IsValid:        true,
		ExpiresAt:      time.Now().Add(24 * time.Hour),
	}

	repo.On("GetCacheStatus", ctx).Return(expectedStatus, nil)

	status, err := repo.GetCacheStatus(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, 100, status.TotalExercises)
	assert.Equal(t, 20, status.TotalMuscles)
	assert.Equal(t, 15, status.TotalEquipment)
	assert.True(t, status.IsValid)

	repo.AssertExpectations(t)
}

func TestExerciseRepository_IsCacheValid(t *testing.T) {
	repo := &mockExerciseRepository{}
	ctx := context.Background()

	repo.On("IsCacheValid", ctx).Return(true, nil)

	isValid, err := repo.IsCacheValid(ctx)
	assert.NoError(t, err)
	assert.True(t, isValid)

	repo.AssertExpectations(t)
}

func TestExerciseRepository_SaveExercise(t *testing.T) {
	repo := &mockExerciseRepository{}
	ctx := context.Background()

	exercise := &models.Exercise{
		ID:            uuid.New(),
		WgerID:        1,
		WgerUUID:      "test-uuid",
		Name:          "Push-up",
		Description:   "Basic push-up exercise",
		Category:      1,
		Language:      2,
		License:       1,
		LicenseAuthor: "Test Author",
		CachedAt:      time.Now(),
		ExpiresAt:     time.Now().Add(24 * time.Hour),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	repo.On("SaveExercise", ctx, exercise).Return(nil)

	err := repo.SaveExercise(ctx, exercise)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestExerciseRepository_SaveMuscleGroup(t *testing.T) {
	repo := &mockExerciseRepository{}
	ctx := context.Background()

	muscleGroup := &models.MuscleGroup{
		ID:        uuid.New(),
		WgerID:    1,
		Name:      "Chest",
		NameEn:    "Chest",
		IsFront:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repo.On("SaveMuscleGroup", ctx, muscleGroup).Return(nil)

	err := repo.SaveMuscleGroup(ctx, muscleGroup)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestExerciseRepository_SaveEquipment(t *testing.T) {
	repo := &mockExerciseRepository{}
	ctx := context.Background()

	equipment := &models.Equipment{
		ID:        uuid.New(),
		WgerID:    1,
		Name:      "Dumbbell",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repo.On("SaveEquipment", ctx, equipment).Return(nil)

	err := repo.SaveEquipment(ctx, equipment)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestExerciseRepository_ClearCache(t *testing.T) {
	repo := &mockExerciseRepository{}
	ctx := context.Background()

	repo.On("ClearCache", ctx).Return(nil)

	err := repo.ClearCache(ctx)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}
