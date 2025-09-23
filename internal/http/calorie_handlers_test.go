package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCalorieService struct {
	mock.Mock
}

func (m *mockCalorieService) CalculateCalories(
	ctx context.Context,
	userID uuid.UUID,
	data *models.CalorieCalculationData,
) (*models.CalorieResults, error) {
	args := m.Called(ctx, userID, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CalorieResults), args.Error(1)
}

func (m *mockCalorieService) GetLastCalculation(ctx context.Context, userID uuid.UUID) (*models.CalorieCalculationResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CalorieCalculationResponse), args.Error(1)
}

func TestCalorieHandlers_CalculateCalories_Success(t *testing.T) {
	mockService := new(mockCalorieService)
	logger := logger.New("INFO", "json")

	handlers := NewCalorieHandlers(mockService, logger)

	userID := uuid.New()
	expectedResults := &models.CalorieResults{
		BMR:            1650,
		TDEE:           2558,
		TargetCalories: 2558,
		Macros: models.Macronutrients{
			ProteinGrams:      196,
			ProteinPercentage: 30.6,
			FatGrams:          85,
			FatPercentage:     30.0,
			CarbsGrams:        256,
			CarbsPercentage:   39.4,
		},
	}

	mockService.On("CalculateCalories", mock.Anything, userID, mock.AnythingOfType("*models.CalorieCalculationData")).
		Return(expectedResults, nil)

	requestData := models.CalorieCalculationData{
		Gender:        models.GenderMale,
		Age:           25,
		Height:        175.0,
		Weight:        70.0,
		ActivityLevel: models.ActivityModeratelyActive,
		Goal:          models.GoalMaintainWeight,
	}

	jsonData, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", "/api/v1/calorie/calculate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, userID.String()))

	w := httptest.NewRecorder()
	handlers.CalculateCalories(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response models.CalorieResults
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResults.BMR, response.BMR)
	assert.Equal(t, expectedResults.TDEE, response.TDEE)
	assert.Equal(t, expectedResults.TargetCalories, response.TargetCalories)
	assert.Equal(t, expectedResults.Macros.ProteinGrams, response.Macros.ProteinGrams)
	assert.Equal(t, expectedResults.Macros.FatGrams, response.Macros.FatGrams)
	assert.Equal(t, expectedResults.Macros.CarbsGrams, response.Macros.CarbsGrams)

	mockService.AssertExpectations(t)
}

func TestCalorieHandlers_CalculateCalories_InvalidMethod(t *testing.T) {
	mockService := new(mockCalorieService)
	logger := logger.New("INFO", "json")

	handlers := NewCalorieHandlers(mockService, logger)

	req := httptest.NewRequest("GET", "/api/v1/calorie/calculate", http.NoBody)
	w := httptest.NewRecorder()
	handlers.CalculateCalories(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestCalorieHandlers_CalculateCalories_InvalidJSON(t *testing.T) {
	mockService := new(mockCalorieService)
	logger := logger.New("INFO", "json")

	handlers := NewCalorieHandlers(mockService, logger)

	req := httptest.NewRequest("POST", "/api/v1/calorie/calculate", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	userID := uuid.New()
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, userID.String()))

	w := httptest.NewRecorder()
	handlers.CalculateCalories(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCalorieHandlers_CalculateCalories_ValidationError(t *testing.T) {
	mockService := new(mockCalorieService)
	logger := logger.New("INFO", "json")

	handlers := NewCalorieHandlers(mockService, logger)

	// Invalid data - age too low
	requestData := models.CalorieCalculationData{
		Gender:        models.GenderMale,
		Age:           10, // Too young
		Height:        175.0,
		Weight:        70.0,
		ActivityLevel: models.ActivityModeratelyActive,
		Goal:          models.GoalMaintainWeight,
	}

	jsonData, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", "/api/v1/calorie/calculate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	userID := uuid.New()
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, userID.String()))

	w := httptest.NewRecorder()
	handlers.CalculateCalories(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCalorieHandlers_CalculateCalories_ServiceError(t *testing.T) {
	mockService := new(mockCalorieService)
	logger := logger.New("INFO", "json")

	handlers := NewCalorieHandlers(mockService, logger)

	userID := uuid.New()
	mockService.On("CalculateCalories", mock.Anything, userID, mock.AnythingOfType("*models.CalorieCalculationData")).
		Return(nil, assert.AnError)

	requestData := models.CalorieCalculationData{
		Gender:        models.GenderMale,
		Age:           25,
		Height:        175.0,
		Weight:        70.0,
		ActivityLevel: models.ActivityModeratelyActive,
		Goal:          models.GoalMaintainWeight,
	}

	jsonData, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", "/api/v1/calorie/calculate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, userID.String()))

	w := httptest.NewRecorder()
	handlers.CalculateCalories(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCalorieHandlers_CalculateCalories_NoUserID(t *testing.T) {
	mockService := new(mockCalorieService)
	logger := logger.New("INFO", "json")

	handlers := NewCalorieHandlers(mockService, logger)

	requestData := models.CalorieCalculationData{
		Gender:        models.GenderMale,
		Age:           25,
		Height:        175.0,
		Weight:        70.0,
		ActivityLevel: models.ActivityModeratelyActive,
		Goal:          models.GoalMaintainWeight,
	}

	jsonData, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", "/api/v1/calorie/calculate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	// No user ID in context

	w := httptest.NewRecorder()
	handlers.CalculateCalories(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCalorieHandlers_GetLastCalculation_Success(t *testing.T) {
	mockService := new(mockCalorieService)
	logger := logger.New("INFO", "json")

	handlers := NewCalorieHandlers(mockService, logger)

	userID := uuid.New()
	expectedResponse := &models.CalorieCalculationResponse{
		Data: models.CalorieCalculationData{
			Gender:        models.GenderMale,
			Age:           25,
			Height:        175.0,
			Weight:        70.0,
			ActivityLevel: models.ActivityModeratelyActive,
			Goal:          models.GoalMaintainWeight,
		},
		Results: models.CalorieResults{
			BMR:            1650,
			TDEE:           2558,
			TargetCalories: 2558,
			Formula:        models.FormulaMifflin,
			Macros: models.Macronutrients{
				ProteinGrams:      196,
				ProteinPercentage: 30.6,
				FatGrams:          85,
				FatPercentage:     30.0,
				CarbsGrams:        256,
				CarbsPercentage:   39.4,
			},
		},
		Timestamp: time.Now(),
	}

	mockService.On("GetLastCalculation", mock.Anything, userID).Return(expectedResponse, nil)

	req := httptest.NewRequest("GET", "/api/v1/calorie/last", http.NoBody)
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, userID.String()))

	w := httptest.NewRecorder()
	handlers.GetLastCalculation(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response models.CalorieCalculationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.Data.Gender, response.Data.Gender)
	assert.Equal(t, expectedResponse.Data.Age, response.Data.Age)
	assert.Equal(t, expectedResponse.Data.Height, response.Data.Height)
	assert.Equal(t, expectedResponse.Data.Weight, response.Data.Weight)
	assert.Equal(t, expectedResponse.Data.ActivityLevel, response.Data.ActivityLevel)
	assert.Equal(t, expectedResponse.Data.Goal, response.Data.Goal)
	assert.Equal(t, expectedResponse.Results.BMR, response.Results.BMR)
	assert.Equal(t, expectedResponse.Results.TDEE, response.Results.TDEE)
	assert.Equal(t, expectedResponse.Results.TargetCalories, response.Results.TargetCalories)
	assert.Equal(t, expectedResponse.Results.Formula, response.Results.Formula)
	assert.Equal(t, expectedResponse.Results.Macros.ProteinGrams, response.Results.Macros.ProteinGrams)
	assert.Equal(t, expectedResponse.Results.Macros.ProteinPercentage, response.Results.Macros.ProteinPercentage)
	assert.Equal(t, expectedResponse.Results.Macros.FatGrams, response.Results.Macros.FatGrams)
	assert.Equal(t, expectedResponse.Results.Macros.FatPercentage, response.Results.Macros.FatPercentage)
	assert.Equal(t, expectedResponse.Results.Macros.CarbsGrams, response.Results.Macros.CarbsGrams)
	assert.Equal(t, expectedResponse.Results.Macros.CarbsPercentage, response.Results.Macros.CarbsPercentage)

	mockService.AssertExpectations(t)
}

func TestCalorieHandlers_GetLastCalculation_InvalidMethod(t *testing.T) {
	mockService := new(mockCalorieService)
	logger := logger.New("INFO", "json")

	handlers := NewCalorieHandlers(mockService, logger)

	req := httptest.NewRequest("POST", "/api/v1/calorie/last", http.NoBody)
	w := httptest.NewRecorder()
	handlers.GetLastCalculation(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestCalorieHandlers_GetLastCalculation_ServiceError(t *testing.T) {
	mockService := new(mockCalorieService)
	logger := logger.New("INFO", "json")

	handlers := NewCalorieHandlers(mockService, logger)

	userID := uuid.New()
	mockService.On("GetLastCalculation", mock.Anything, userID).Return(nil, assert.AnError)

	req := httptest.NewRequest("GET", "/api/v1/calorie/last", http.NoBody)
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, userID.String()))

	w := httptest.NewRecorder()
	handlers.GetLastCalculation(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCalorieHandlers_GetLastCalculation_NoUserID(t *testing.T) {
	mockService := new(mockCalorieService)
	logger := logger.New("INFO", "json")

	handlers := NewCalorieHandlers(mockService, logger)

	req := httptest.NewRequest("GET", "/api/v1/calorie/last", http.NoBody)
	// No user ID in context

	w := httptest.NewRecorder()
	handlers.GetLastCalculation(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
