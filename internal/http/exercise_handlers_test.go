package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockExerciseService struct {
	mock.Mock
}

func (m *MockExerciseService) GetExercises(ctx context.Context, filters *models.ExerciseFilters) (*models.ExerciseListResponse, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).(*models.ExerciseListResponse), args.Error(1)
}

func (m *MockExerciseService) GetExerciseByID(ctx context.Context, id uuid.UUID) (*models.Exercise, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Exercise), args.Error(1)
}

func (m *MockExerciseService) GetMuscleGroups(ctx context.Context) ([]models.MuscleGroup, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.MuscleGroup), args.Error(1)
}

func (m *MockExerciseService) GetEquipment(ctx context.Context) ([]models.Equipment, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Equipment), args.Error(1)
}

func (m *MockExerciseService) GetCacheStatus(ctx context.Context) (*models.CacheStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*models.CacheStatus), args.Error(1)
}

func (m *MockExerciseService) RefreshCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestExerciseHandlers_GetExercises(t *testing.T) {
	mockService := &MockExerciseService{}
	logger := logger.New("DEBUG", "text")

	handlers := NewExerciseHandlers(mockService, logger)

	ctx := context.Background()
	expectedResponse := &models.ExerciseListResponse{
		Exercises: []models.Exercise{
			{
				ID:   uuid.New(),
				Name: "Push-up",
			},
		},
		Total: 1,
		Page:  1,
		Limit: 20,
	}

	mockService.On("GetExercises", ctx, mock.AnythingOfType("*models.ExerciseFilters")).Return(expectedResponse, nil)

	req := httptest.NewRequest("GET", "/api/v1/exercises", http.NoBody)
	w := httptest.NewRecorder()

	handlers.GetExercises(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	mockService.AssertExpectations(t)
}

func TestExerciseHandlers_GetExercises_WithFilters(t *testing.T) {
	mockService := &MockExerciseService{}
	logger := logger.New("DEBUG", "text")

	handlers := NewExerciseHandlers(mockService, logger)

	ctx := context.Background()
	expectedResponse := &models.ExerciseListResponse{
		Exercises: []models.Exercise{
			{
				ID:   uuid.New(),
				Name: "Bench Press",
			},
		},
		Total: 1,
		Page:  1,
		Limit: 10,
	}

	mockService.On("GetExercises", ctx, mock.AnythingOfType("*models.ExerciseFilters")).Return(expectedResponse, nil)

	req := httptest.NewRequest("GET", "/api/v1/exercises?muscle_group_id=123e4567-e89b-12d3-a456-426614174000&limit=10&page=1", http.NoBody)
	w := httptest.NewRecorder()

	handlers.GetExercises(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	mockService.AssertExpectations(t)
}

func TestExerciseHandlers_GetExerciseByID(t *testing.T) {
	mockService := &MockExerciseService{}
	logger := logger.New("DEBUG", "text")

	handlers := NewExerciseHandlers(mockService, logger)

	ctx := context.Background()
	exerciseID := uuid.New()
	expectedExercise := &models.Exercise{
		ID:   exerciseID,
		Name: "Push-up",
	}

	mockService.On("GetExerciseByID", ctx, exerciseID).Return(expectedExercise, nil)

	req := httptest.NewRequest("GET", "/api/v1/exercises/"+exerciseID.String(), http.NoBody)
	w := httptest.NewRecorder()

	handlers.GetExerciseByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	mockService.AssertExpectations(t)
}

func TestExerciseHandlers_GetMuscleGroups(t *testing.T) {
	mockService := &MockExerciseService{}
	logger := logger.New("DEBUG", "text")

	handlers := NewExerciseHandlers(mockService, logger)

	ctx := context.Background()
	expectedMuscleGroups := []models.MuscleGroup{
		{
			ID:   uuid.New(),
			Name: "Chest",
		},
	}

	mockService.On("GetMuscleGroups", ctx).Return(expectedMuscleGroups, nil)

	req := httptest.NewRequest("GET", "/api/v1/exercises/muscle-groups", http.NoBody)
	w := httptest.NewRecorder()

	handlers.GetMuscleGroups(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	mockService.AssertExpectations(t)
}

func TestExerciseHandlers_GetEquipment(t *testing.T) {
	mockService := &MockExerciseService{}
	logger := logger.New("DEBUG", "text")

	handlers := NewExerciseHandlers(mockService, logger)

	ctx := context.Background()
	expectedEquipment := []models.Equipment{
		{
			ID:   uuid.New(),
			Name: "Dumbbells",
		},
	}

	mockService.On("GetEquipment", ctx).Return(expectedEquipment, nil)

	req := httptest.NewRequest("GET", "/api/v1/exercises/equipment", http.NoBody)
	w := httptest.NewRecorder()

	handlers.GetEquipment(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	mockService.AssertExpectations(t)
}

func TestExerciseHandlers_GetCacheStatus(t *testing.T) {
	mockService := &MockExerciseService{}
	logger := logger.New("DEBUG", "text")

	handlers := NewExerciseHandlers(mockService, logger)

	ctx := context.Background()
	expectedStatus := &models.CacheStatus{
		TotalExercises: 100,
		TotalMuscles:   14,
		TotalEquipment: 8,
		IsValid:        true,
	}

	mockService.On("GetCacheStatus", ctx).Return(expectedStatus, nil)

	req := httptest.NewRequest("GET", "/api/v1/exercises/cache/status", http.NoBody)
	w := httptest.NewRecorder()

	handlers.GetCacheStatus(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	mockService.AssertExpectations(t)
}

func TestExerciseHandlers_RefreshCache(t *testing.T) {
	mockService := &MockExerciseService{}
	logger := logger.New("DEBUG", "text")

	handlers := NewExerciseHandlers(mockService, logger)

	ctx := context.Background()

	mockService.On("RefreshCache", ctx).Return(nil)

	req := httptest.NewRequest("POST", "/api/v1/exercises/cache/refresh", http.NoBody)
	w := httptest.NewRecorder()

	handlers.RefreshCache(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	mockService.AssertExpectations(t)
}

func TestExerciseHandlers_InvalidMethod(t *testing.T) {
	mockService := &MockExerciseService{}
	logger := logger.New("DEBUG", "text")

	handlers := NewExerciseHandlers(mockService, logger)

	req := httptest.NewRequest("POST", "/api/v1/exercises", http.NoBody)
	w := httptest.NewRecorder()

	handlers.GetExercises(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
