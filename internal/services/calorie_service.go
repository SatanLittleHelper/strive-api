package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/aleksandr/strive-api/internal/models"
	"github.com/aleksandr/strive-api/internal/repositories"
)

type CalorieService interface {
	CalculateCalories(ctx context.Context, userID uuid.UUID, data *models.CalorieCalculationData) (*models.CalorieResults, error)
	GetLastCalculation(ctx context.Context, userID uuid.UUID) (*models.CalorieCalculationResponse, error)
}

type calorieService struct {
	calorieRepo repositories.CalorieRepository
}

func NewCalorieService(calorieRepo repositories.CalorieRepository) CalorieService {
	return &calorieService{
		calorieRepo: calorieRepo,
	}
}

func (s *calorieService) CalculateCalories(
	ctx context.Context,
	userID uuid.UUID,
	data *models.CalorieCalculationData,
) (*models.CalorieResults, error) {
	bmr := s.calculateBMRMifflin(data)
	tdee := s.calculateTDEE(bmr, data.ActivityLevel)
	targetCalories := s.calculateTargetCalories(tdee, data.Goal)
	macros := s.calculateMacronutrients(
		targetCalories,
		data.ActivityLevel,
		data.Goal,
		data.Weight,
		data.Gender,
		data.Age,
		data.Height,
		data.BodyFatPercentage,
	)

	results := &models.CalorieResults{
		BMR:            int(math.Round(bmr)),
		TDEE:           tdee,
		TargetCalories: targetCalories,
		Formula:        models.FormulaMifflin,
		Macros:         macros,
	}

	calculation := &models.CalorieCalculation{
		ID:                uuid.New(),
		UserID:            userID,
		Gender:            data.Gender,
		Age:               data.Age,
		Height:            data.Height,
		Weight:            data.Weight,
		ActivityLevel:     data.ActivityLevel,
		Goal:              data.Goal,
		BMR:               results.BMR,
		TDEE:              results.TDEE,
		TargetCalories:    results.TargetCalories,
		Formula:           results.Formula,
		ProteinGrams:      macros.ProteinGrams,
		ProteinPercentage: macros.ProteinPercentage,
		FatGrams:          macros.FatGrams,
		FatPercentage:     macros.FatPercentage,
		CarbsGrams:        macros.CarbsGrams,
		CarbsPercentage:   macros.CarbsPercentage,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.calorieRepo.SaveOrUpdate(ctx, calculation); err != nil {
		return nil, fmt.Errorf("failed to save calculation: %w", err)
	}

	return results, nil
}

func (s *calorieService) GetLastCalculation(ctx context.Context, userID uuid.UUID) (*models.CalorieCalculationResponse, error) {
	calculation, err := s.calorieRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get calculation: %w", err)
	}

	if calculation == nil {
		return nil, fmt.Errorf("no calculation found for user")
	}

	data := models.CalorieCalculationData{
		Gender:        calculation.Gender,
		Age:           calculation.Age,
		Height:        calculation.Height,
		Weight:        calculation.Weight,
		ActivityLevel: calculation.ActivityLevel,
		Goal:          calculation.Goal,
	}

	results := models.CalorieResults{
		BMR:            calculation.BMR,
		TDEE:           calculation.TDEE,
		TargetCalories: calculation.TargetCalories,
		Formula:        calculation.Formula,
		Macros: models.Macronutrients{
			ProteinGrams:      calculation.ProteinGrams,
			ProteinPercentage: calculation.ProteinPercentage,
			FatGrams:          calculation.FatGrams,
			FatPercentage:     calculation.FatPercentage,
			CarbsGrams:        calculation.CarbsGrams,
			CarbsPercentage:   calculation.CarbsPercentage,
		},
	}

	return &models.CalorieCalculationResponse{
		Data:      data,
		Results:   results,
		Timestamp: calculation.CreatedAt,
	}, nil
}

func (s *calorieService) calculateBMRMifflin(data *models.CalorieCalculationData) float64 {
	const (
		weightMultiplier = 10.0
		heightMultiplier = 6.25
		ageMultiplier    = 5.0
		maleOffset       = 5.0
		femaleOffset     = -161.0
	)

	base := weightMultiplier*data.Weight + heightMultiplier*data.Height - ageMultiplier*float64(data.Age)

	if data.Gender == models.GenderMale {
		return base + maleOffset
	}
	return base + femaleOffset
}

func (s *calorieService) calculateTDEE(bmr float64, activityLevel string) int {
	return int(math.Round(bmr * models.GetActivityMultiplier(activityLevel)))
}

func (s *calorieService) calculateTargetCalories(tdee int, goal string) int {
	modifier := models.GetGoalModifier(goal)
	return int(math.Round(float64(tdee) * (1 + modifier)))
}

func (s *calorieService) calculateBodyFatPercentage(gender string, age int, weight, height float64) float64 {
	bmi := weight / ((height / 100) * (height / 100))
	genderValue := 1.0
	if gender == "female" {
		genderValue = 0.0
	}

	bodyFatPercentage := 1.20*bmi + 0.23*float64(age) - 10.8*genderValue - 5.4

	// Более реалистичные ограничения для процента жира
	if gender == "male" {
		if bodyFatPercentage < 8.0 {
			bodyFatPercentage = 8.0
		}
		if bodyFatPercentage > 25.0 {
			bodyFatPercentage = 25.0
		}
	} else {
		if bodyFatPercentage < 12.0 {
			bodyFatPercentage = 12.0
		}
		if bodyFatPercentage > 35.0 {
			bodyFatPercentage = 35.0
		}
	}

	return bodyFatPercentage
}

func (s *calorieService) getBodyFatPercentage(provided *float64, gender string, age int, weight, height float64) float64 {
	if provided != nil {
		return *provided
	}
	return s.calculateBodyFatPercentage(gender, age, weight, height)
}

func (s *calorieService) calculateMacronutrients(
	targetCalories int,
	activityLevel, goal string,
	weight float64,
	gender string,
	age int,
	height float64,
	bodyFatPercentage *float64,
) models.Macronutrients {
	actualBodyFatPercentage := s.getBodyFatPercentage(bodyFatPercentage, gender, age, weight, height)
	leanBodyMass := weight * (1 - actualBodyFatPercentage/100)

	baseProteinPerKg := models.BaseProteinByActivity[activityLevel]
	proteinAdjustment := models.ProteinAdjustmentByGoal[goal]

	proteinPerKg := baseProteinPerKg + proteinAdjustment

	proteinGrams := int(math.Round(proteinPerKg * leanBodyMass))
	proteinCalories := proteinGrams * 4

	fatPercentage := models.FatPercentageByGoal[goal]
	fatCalories := int(math.Round(float64(targetCalories) * fatPercentage / 100))
	fatGrams := int(math.Round(float64(fatCalories) / 9))

	carbsCalories := targetCalories - proteinCalories - fatCalories
	carbsGrams := int(math.Round(float64(carbsCalories) / 4))

	if carbsGrams < 0 {
		carbsGrams = 0
		carbsCalories = 0
		adjustedFatCalories := targetCalories - proteinCalories
		fatGrams = int(math.Round(float64(adjustedFatCalories) / 9))
		fatCalories = fatGrams * 9
	}

	actualProteinPercentage := float64(proteinCalories) / float64(targetCalories) * 100
	actualFatPercentage := float64(fatCalories) / float64(targetCalories) * 100
	actualCarbsPercentage := float64(carbsCalories) / float64(targetCalories) * 100

	return models.Macronutrients{
		ProteinGrams:      proteinGrams,
		ProteinPercentage: math.Round(actualProteinPercentage*100) / 100,
		FatGrams:          fatGrams,
		FatPercentage:     math.Round(actualFatPercentage*100) / 100,
		CarbsGrams:        carbsGrams,
		CarbsPercentage:   math.Round(actualCarbsPercentage*100) / 100,
	}
}
