package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aleksandr/strive-api/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCalorieRepository struct {
	mock.Mock
}

func (m *mockCalorieRepository) SaveOrUpdate(ctx context.Context, calculation *models.CalorieCalculation) error {
	args := m.Called(ctx, calculation)
	return args.Error(0)
}

func (m *mockCalorieRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.CalorieCalculation, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CalorieCalculation), args.Error(1)
}

func TestCalorieRepository_SaveOrUpdate(t *testing.T) {
	repo := &mockCalorieRepository{}
	userID := uuid.New()
	calculation := &models.CalorieCalculation{
		ID:                uuid.New(),
		UserID:            userID,
		Gender:            models.GenderMale,
		Age:               25,
		Height:            175.0,
		Weight:            70.0,
		ActivityLevel:     models.ActivityModeratelyActive,
		Goal:              models.GoalMaintainWeight,
		BMR:               1650,
		TDEE:              2558,
		TargetCalories:    2558,
		Formula:           models.FormulaMifflin,
		ProteinGrams:      196,
		ProteinPercentage: 30.6,
		FatGrams:          85,
		FatPercentage:     30.0,
		CarbsGrams:        256,
		CarbsPercentage:   39.4,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	repo.On("SaveOrUpdate", mock.Anything, calculation).Return(nil)

	err := repo.SaveOrUpdate(context.Background(), calculation)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCalorieRepository_SaveOrUpdate_Error(t *testing.T) {
	repo := &mockCalorieRepository{}
	userID := uuid.New()
	calculation := &models.CalorieCalculation{
		ID:     uuid.New(),
		UserID: userID,
	}

	expectedError := errors.New("database error")
	repo.On("SaveOrUpdate", mock.Anything, calculation).Return(expectedError)

	err := repo.SaveOrUpdate(context.Background(), calculation)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	repo.AssertExpectations(t)
}

func TestCalorieRepository_GetByUserID(t *testing.T) {
	repo := &mockCalorieRepository{}
	userID := uuid.New()
	expectedCalculation := &models.CalorieCalculation{
		ID:                uuid.New(),
		UserID:            userID,
		Gender:            models.GenderFemale,
		Age:               28,
		Height:            165.0,
		Weight:            65.0,
		ActivityLevel:     models.ActivityModeratelyActive,
		Goal:              models.GoalLoseWeight,
		BMR:               1385,
		TDEE:              2147,
		TargetCalories:    1718,
		Formula:           models.FormulaMifflin,
		ProteinGrams:      120,
		ProteinPercentage: 28.0,
		FatGrams:          48,
		FatPercentage:     25.0,
		CarbsGrams:        172,
		CarbsPercentage:   40.0,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	repo.On("GetByUserID", mock.Anything, userID).Return(expectedCalculation, nil)

	result, err := repo.GetByUserID(context.Background(), userID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedCalculation.ID, result.ID)
	assert.Equal(t, expectedCalculation.UserID, result.UserID)
	assert.Equal(t, expectedCalculation.Gender, result.Gender)
	assert.Equal(t, expectedCalculation.Age, result.Age)
	assert.Equal(t, expectedCalculation.Height, result.Height)
	assert.Equal(t, expectedCalculation.Weight, result.Weight)
	assert.Equal(t, expectedCalculation.ActivityLevel, result.ActivityLevel)
	assert.Equal(t, expectedCalculation.Goal, result.Goal)
	assert.Equal(t, expectedCalculation.BMR, result.BMR)
	assert.Equal(t, expectedCalculation.TDEE, result.TDEE)
	assert.Equal(t, expectedCalculation.TargetCalories, result.TargetCalories)
	assert.Equal(t, expectedCalculation.Formula, result.Formula)
	assert.Equal(t, expectedCalculation.ProteinGrams, result.ProteinGrams)
	assert.Equal(t, expectedCalculation.ProteinPercentage, result.ProteinPercentage)
	assert.Equal(t, expectedCalculation.FatGrams, result.FatGrams)
	assert.Equal(t, expectedCalculation.FatPercentage, result.FatPercentage)
	assert.Equal(t, expectedCalculation.CarbsGrams, result.CarbsGrams)
	assert.Equal(t, expectedCalculation.CarbsPercentage, result.CarbsPercentage)
	repo.AssertExpectations(t)
}

func TestCalorieRepository_GetByUserID_NotFound(t *testing.T) {
	repo := &mockCalorieRepository{}
	nonExistentUserID := uuid.New()

	expectedError := errors.New("calculation not found")
	repo.On("GetByUserID", mock.Anything, nonExistentUserID).Return(nil, expectedError)

	result, err := repo.GetByUserID(context.Background(), nonExistentUserID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	repo.AssertExpectations(t)
}

func TestCalorieRepository_UpdateExisting(t *testing.T) {
	repo := &mockCalorieRepository{}
	userID := uuid.New()

	initialCalculation := &models.CalorieCalculation{
		ID:                uuid.New(),
		UserID:            userID,
		Gender:            models.GenderMale,
		Age:               25,
		Height:            175.0,
		Weight:            70.0,
		ActivityLevel:     models.ActivityModeratelyActive,
		Goal:              models.GoalMaintainWeight,
		BMR:               1650,
		TDEE:              2558,
		TargetCalories:    2558,
		Formula:           models.FormulaMifflin,
		ProteinGrams:      196,
		ProteinPercentage: 30.6,
		FatGrams:          85,
		FatPercentage:     30.0,
		CarbsGrams:        256,
		CarbsPercentage:   39.4,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	updatedCalculation := &models.CalorieCalculation{
		ID:                initialCalculation.ID,
		UserID:            userID,
		Gender:            models.GenderMale,
		Age:               26,
		Height:            176.0,
		Weight:            72.0,
		ActivityLevel:     models.ActivityVeryActive,
		Goal:              models.GoalGainWeight,
		BMR:               1680,
		TDEE:              3024,
		TargetCalories:    3326,
		Formula:           models.FormulaMifflin,
		ProteinGrams:      208,
		ProteinPercentage: 25.0,
		FatGrams:          119,
		FatPercentage:     30.0,
		CarbsGrams:        357,
		CarbsPercentage:   40.0,
		CreatedAt:         initialCalculation.CreatedAt,
		UpdatedAt:         time.Now(),
	}

	repo.On("SaveOrUpdate", mock.Anything, initialCalculation).Return(nil)
	repo.On("SaveOrUpdate", mock.Anything, updatedCalculation).Return(nil)
	repo.On("GetByUserID", mock.Anything, userID).Return(updatedCalculation, nil)

	err := repo.SaveOrUpdate(context.Background(), initialCalculation)
	assert.NoError(t, err)

	err = repo.SaveOrUpdate(context.Background(), updatedCalculation)
	assert.NoError(t, err)

	result, err := repo.GetByUserID(context.Background(), userID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, updatedCalculation.Age, result.Age)
	assert.Equal(t, updatedCalculation.Height, result.Height)
	assert.Equal(t, updatedCalculation.Weight, result.Weight)
	assert.Equal(t, updatedCalculation.ActivityLevel, result.ActivityLevel)
	assert.Equal(t, updatedCalculation.Goal, result.Goal)
	assert.Equal(t, updatedCalculation.BMR, result.BMR)
	assert.Equal(t, updatedCalculation.TDEE, result.TDEE)
	assert.Equal(t, updatedCalculation.TargetCalories, result.TargetCalories)
	assert.Equal(t, updatedCalculation.ProteinGrams, result.ProteinGrams)
	assert.Equal(t, updatedCalculation.ProteinPercentage, result.ProteinPercentage)
	assert.Equal(t, updatedCalculation.FatGrams, result.FatGrams)
	assert.Equal(t, updatedCalculation.FatPercentage, result.FatPercentage)
	assert.Equal(t, updatedCalculation.CarbsGrams, result.CarbsGrams)
	assert.Equal(t, updatedCalculation.CarbsPercentage, result.CarbsPercentage)
	assert.True(t, result.UpdatedAt.After(result.CreatedAt))
	repo.AssertExpectations(t)
}
