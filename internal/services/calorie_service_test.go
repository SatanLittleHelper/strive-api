package services

import (
	"context"
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

func TestCalorieService_CalculateCalories(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		data           *models.CalorieCalculationData
		expectedBMR    int
		expectedTDEE   int
		expectedTarget int
	}{
		{
			name:   "Male, 25 years, 175cm, 70kg, moderately active, maintain weight",
			userID: uuid.New(),
			data: &models.CalorieCalculationData{
				Gender:        models.GenderMale,
				Age:           25,
				Height:        175.0,
				Weight:        70.0,
				ActivityLevel: models.ActivityModeratelyActive,
				Goal:          models.GoalMaintainWeight,
			},
			expectedBMR:    1674,
			expectedTDEE:   2594,
			expectedTarget: 2594,
		},
		{
			name:   "Female, 28 years, 165cm, 65kg, moderately active, lose weight",
			userID: uuid.New(),
			data: &models.CalorieCalculationData{
				Gender:        models.GenderFemale,
				Age:           28,
				Height:        165.0,
				Weight:        65.0,
				ActivityLevel: models.ActivityModeratelyActive,
				Goal:          models.GoalLoseWeight,
			},
			expectedBMR:    1380,
			expectedTDEE:   2139,
			expectedTarget: 1711,
		},
		{
			name:   "Male, 30 years, 180cm, 80kg, very active, gain weight",
			userID: uuid.New(),
			data: &models.CalorieCalculationData{
				Gender:        models.GenderMale,
				Age:           30,
				Height:        180.0,
				Weight:        80.0,
				ActivityLevel: models.ActivityVeryActive,
				Goal:          models.GoalGainWeight,
			},
			expectedBMR:    1780,
			expectedTDEE:   3071,
			expectedTarget: 3532,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockCalorieRepository)
			mockRepo.On("SaveOrUpdate", mock.Anything, mock.AnythingOfType("*models.CalorieCalculation")).Return(nil)

			service := NewCalorieService(mockRepo)
			result, err := service.CalculateCalories(context.Background(), tt.userID, tt.data)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedBMR, result.BMR)
			assert.Equal(t, tt.expectedTDEE, result.TDEE)
			assert.Equal(t, tt.expectedTarget, result.TargetCalories)
			assert.NotNil(t, result.Macros)
			assert.Greater(t, result.Macros.ProteinGrams, 0)
			assert.Greater(t, result.Macros.FatGrams, 0)
			assert.Greater(t, result.Macros.CarbsGrams, 0)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCalorieService_GetLastCalculation(t *testing.T) {
	userID := uuid.New()
	expectedCalculation := &models.CalorieCalculation{
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

	mockRepo := new(mockCalorieRepository)
	mockRepo.On("GetByUserID", mock.Anything, userID).Return(expectedCalculation, nil)

	service := NewCalorieService(mockRepo)
	result, err := service.GetLastCalculation(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedCalculation.Gender, result.Data.Gender)
	assert.Equal(t, expectedCalculation.Gender, result.Data.Gender)
	assert.Equal(t, expectedCalculation.Age, result.Data.Age)
	assert.Equal(t, expectedCalculation.Height, result.Data.Height)
	assert.Equal(t, expectedCalculation.Weight, result.Data.Weight)
	assert.Equal(t, expectedCalculation.ActivityLevel, result.Data.ActivityLevel)
	assert.Equal(t, expectedCalculation.Goal, result.Data.Goal)
	assert.Equal(t, expectedCalculation.BMR, result.Results.BMR)
	assert.Equal(t, expectedCalculation.TDEE, result.Results.TDEE)
	assert.Equal(t, expectedCalculation.TargetCalories, result.Results.TargetCalories)
	assert.Equal(t, expectedCalculation.Formula, result.Results.Formula)
	assert.Equal(t, expectedCalculation.ProteinGrams, result.Results.Macros.ProteinGrams)
	assert.Equal(t, expectedCalculation.ProteinPercentage, result.Results.Macros.ProteinPercentage)
	assert.Equal(t, expectedCalculation.FatGrams, result.Results.Macros.FatGrams)
	assert.Equal(t, expectedCalculation.FatPercentage, result.Results.Macros.FatPercentage)
	assert.Equal(t, expectedCalculation.CarbsGrams, result.Results.Macros.CarbsGrams)
	assert.Equal(t, expectedCalculation.CarbsPercentage, result.Results.Macros.CarbsPercentage)
	assert.Equal(t, expectedCalculation.CreatedAt, result.Timestamp)

	mockRepo.AssertExpectations(t)
}

func TestCalorieService_GetLastCalculation_NotFound(t *testing.T) {
	userID := uuid.New()

	mockRepo := new(mockCalorieRepository)
	mockRepo.On("GetByUserID", mock.Anything, userID).Return(nil, assert.AnError)

	service := NewCalorieService(mockRepo)
	result, err := service.GetLastCalculation(context.Background(), userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestCalorieService_CalculateBMRMifflin(t *testing.T) {
	service := &calorieService{}

	tests := []struct {
		name     string
		data     models.CalorieCalculationData
		expected int
	}{
		{
			name: "Male, 25 years, 175cm, 70kg",
			data: models.CalorieCalculationData{
				Gender: models.GenderMale,
				Age:    25,
				Height: 175.0,
				Weight: 70.0,
			},
			expected: 1673,
		},
		{
			name: "Female, 28 years, 165cm, 65kg",
			data: models.CalorieCalculationData{
				Gender: models.GenderFemale,
				Age:    28,
				Height: 165.0,
				Weight: 65.0,
			},
			expected: 1380,
		},
		{
			name: "Male, 30 years, 180cm, 80kg",
			data: models.CalorieCalculationData{
				Gender: models.GenderMale,
				Age:    30,
				Height: 180.0,
				Weight: 80.0,
			},
			expected: 1780,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateBMRMifflin(&tt.data)
			assert.Equal(t, tt.expected, int(result))
		})
	}
}

func TestCalorieService_CalculateTDEE(t *testing.T) {
	service := &calorieService{}

	tests := []struct {
		name          string
		bmr           float64
		activityLevel string
		expected      int
	}{
		{
			name:          "Sedentary",
			bmr:           1650.0,
			activityLevel: models.ActivitySedentary,
			expected:      1980,
		},
		{
			name:          "Lightly Active",
			bmr:           1650.0,
			activityLevel: models.ActivityLightlyActive,
			expected:      2269,
		},
		{
			name:          "Moderately Active",
			bmr:           1650.0,
			activityLevel: models.ActivityModeratelyActive,
			expected:      2558,
		},
		{
			name:          "Very Active",
			bmr:           1650.0,
			activityLevel: models.ActivityVeryActive,
			expected:      2846,
		},
		{
			name:          "Extremely Active",
			bmr:           1650.0,
			activityLevel: models.ActivityExtremelyActive,
			expected:      3135,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateTDEE(tt.bmr, tt.activityLevel)
			assert.Equal(t, tt.expected, int(result))
		})
	}
}

func TestCalorieService_CalculateTargetCalories(t *testing.T) {
	service := &calorieService{}

	tests := []struct {
		name     string
		tdee     float64
		goal     string
		expected int
	}{
		{
			name:     "Lose Weight",
			tdee:     2000.0,
			goal:     models.GoalLoseWeight,
			expected: 1600,
		},
		{
			name:     "Maintain Weight",
			tdee:     2000.0,
			goal:     models.GoalMaintainWeight,
			expected: 2000,
		},
		{
			name:     "Gain Weight",
			tdee:     2000.0,
			goal:     models.GoalGainWeight,
			expected: 2300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateTargetCalories(int(tt.tdee), tt.goal)
			assert.Equal(t, tt.expected, int(result))
		})
	}
}

func TestCalorieService_CalculateMacronutrients(t *testing.T) {
	service := &calorieService{}

	tests := []struct {
		name            string
		targetCalories  int
		activityLevel   string
		goal            string
		weight          float64
		expectedProtein int
		expectedFat     int
		expectedCarbs   int
	}{
		{
			name:            "Moderately Active, Maintain Weight",
			targetCalories:  2000,
			activityLevel:   models.ActivityModeratelyActive,
			goal:            models.GoalMaintainWeight,
			weight:          80.0,
			expectedProtein: 140,
			expectedFat:     67,
			expectedCarbs:   200,
		},
		{
			name:            "Very Active, Lose Weight",
			targetCalories:  1800,
			activityLevel:   models.ActivityVeryActive,
			goal:            models.GoalLoseWeight,
			weight:          70.0,
			expectedProtein: 144,
			expectedFat:     50,
			expectedCarbs:   180,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateMacronutrients(tt.targetCalories, tt.activityLevel, tt.goal, tt.weight, "male", 30, 180.0, nil)

			assert.NotNil(t, result)
			assert.Greater(t, result.ProteinGrams, 0)
			assert.Greater(t, result.FatGrams, 0)
			assert.Greater(t, result.CarbsGrams, 0)

			// Check that percentages add up to approximately 100%
			totalPercentage := result.ProteinPercentage + result.FatPercentage + result.CarbsPercentage
			assert.InDelta(t, 100.0, totalPercentage, 1.0)
		})
	}
}
