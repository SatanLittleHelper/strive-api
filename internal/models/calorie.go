package models

import (
	"time"

	"github.com/google/uuid"
)

type CalorieCalculationData struct {
	Gender            string   `json:"gender" validate:"required,oneof=male female"`
	Age               int      `json:"age" validate:"required,min=15,max=120"`
	Height            float64  `json:"height" validate:"required,min=100,max=250"`
	Weight            float64  `json:"weight" validate:"required,min=30,max=300"`
	ActivityLevel     string   `json:"activityLevel" validate:"required"`
	Goal              string   `json:"goal" validate:"required,oneof=lose_weight maintain_weight gain_weight"`
	BodyFatPercentage *float64 `json:"bodyFatPercentage,omitempty" validate:"omitempty,min=5,max=50"`
}

type Macronutrients struct {
	ProteinGrams      int     `json:"proteinGrams"`
	ProteinPercentage float64 `json:"proteinPercentage"`
	FatGrams          int     `json:"fatGrams"`
	FatPercentage     float64 `json:"fatPercentage"`
	CarbsGrams        int     `json:"carbsGrams"`
	CarbsPercentage   float64 `json:"carbsPercentage"`
}

type CalorieResults struct {
	BMR            int            `json:"bmr"`
	TDEE           int            `json:"tdee"`
	TargetCalories int            `json:"targetCalories"`
	Formula        string         `json:"formula"`
	Macros         Macronutrients `json:"macros"`
}

type CalorieCalculation struct {
	ID                uuid.UUID `json:"id" db:"id"`
	UserID            uuid.UUID `json:"user_id" db:"user_id"`
	Gender            string    `json:"gender" db:"gender"`
	Age               int       `json:"age" db:"age"`
	Height            float64   `json:"height" db:"height"`
	Weight            float64   `json:"weight" db:"weight"`
	ActivityLevel     string    `json:"activityLevel" db:"activity_level"`
	Goal              string    `json:"goal" db:"goal"`
	BMR               int       `json:"bmr" db:"bmr"`
	TDEE              int       `json:"tdee" db:"tdee"`
	TargetCalories    int       `json:"targetCalories" db:"target_calories"`
	Formula           string    `json:"formula" db:"formula"`
	ProteinGrams      int       `json:"proteinGrams" db:"protein_grams"`
	ProteinPercentage float64   `json:"proteinPercentage" db:"protein_percentage"`
	FatGrams          int       `json:"fatGrams" db:"fat_grams"`
	FatPercentage     float64   `json:"fatPercentage" db:"fat_percentage"`
	CarbsGrams        int       `json:"carbsGrams" db:"carbs_grams"`
	CarbsPercentage   float64   `json:"carbsPercentage" db:"carbs_percentage"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type CalorieCalculationResponse struct {
	Data      CalorieCalculationData `json:"data"`
	Results   CalorieResults         `json:"results"`
	Timestamp time.Time              `json:"timestamp"`
}
