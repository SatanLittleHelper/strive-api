package models

const (
	// Gender constants
	GenderMale   = "male"
	GenderFemale = "female"

	// Activity level constants
	ActivitySedentary        = "sedentary"
	ActivityLightlyActive    = "lightly_active"
	ActivityModeratelyActive = "moderately_active"
	ActivityVeryActive       = "very_active"
	ActivityExtremelyActive  = "extremely_active"

	// Goal constants
	GoalLoseWeight     = "lose_weight"
	GoalMaintainWeight = "maintain_weight"
	GoalGainWeight     = "gain_weight"

	// Formula constants
	FormulaMifflin = "mifflin"

	// Activity level multipliers for TDEE calculation
	ActivityMultiplierSedentary        = 1.2
	ActivityMultiplierLightlyActive    = 1.375
	ActivityMultiplierModeratelyActive = 1.55
	ActivityMultiplierVeryActive       = 1.725
	ActivityMultiplierExtremelyActive  = 1.9

	// Goal modifiers for target calories
	GoalModifierLoseWeight     = -0.20
	GoalModifierMaintainWeight = 0.0
	GoalModifierGainWeight     = 0.15

	// Validation constants
	ActivityLevelValidation = "required,oneof=sedentary lightly_active moderately_active very_active extremely_active"
	GoalValidation          = "required,oneof=lose_weight maintain_weight gain_weight"
	GenderValidation        = "required,oneof=male female"
	AgeValidation           = "required,min=15,max=120"
	HeightValidation        = "required,min=100,max=250"
	WeightValidation        = "required,min=30,max=300"
	BodyFatValidation       = "omitempty,min=5,max=50"
)

// Base protein per kg by activity level (на сухую массу тела)
var BaseProteinByActivity = map[string]float64{
	ActivitySedentary:        1.2, // Здоровый активный человек
	ActivityLightlyActive:    1.4, // Здоровый активный человек
	ActivityModeratelyActive: 1.6, // Здоровый активный человек
	ActivityVeryActive:       1.8, // Силовые тренировки
	ActivityExtremelyActive:  2.0, // Силовые тренировки
}

// Protein adjustments by goal
var ProteinAdjustmentByGoal = map[string]float64{
	GoalLoseWeight:     0.4, // При похудении: 1.6 + 0.4 = 2.0г/кг (сохранить мышцы)
	GoalMaintainWeight: 0.0, // Без изменений
	GoalGainWeight:     0.2, // При наборе: 1.6 + 0.2 = 1.8г/кг (набор мышц)
}

// Fat percentages by goal
var FatPercentageByGoal = map[string]float64{
	GoalLoseWeight:     25.0,
	GoalMaintainWeight: 30.0,
	GoalGainWeight:     35.0,
}

// Genders returns slice of gender values
func Genders() []string {
	return []string{GenderMale, GenderFemale}
}

// ActivityLevels returns slice of activity level values
func ActivityLevels() []string {
	return []string{
		ActivitySedentary,
		ActivityLightlyActive,
		ActivityModeratelyActive,
		ActivityVeryActive,
		ActivityExtremelyActive,
	}
}

// Goals returns slice of goal values
func Goals() []string {
	return []string{GoalLoseWeight, GoalMaintainWeight, GoalGainWeight}
}

// GetActivityMultiplier returns the multiplier for given activity level
func GetActivityMultiplier(activityLevel string) float64 {
	multipliers := map[string]float64{
		ActivitySedentary:        ActivityMultiplierSedentary,
		ActivityLightlyActive:    ActivityMultiplierLightlyActive,
		ActivityModeratelyActive: ActivityMultiplierModeratelyActive,
		ActivityVeryActive:       ActivityMultiplierVeryActive,
		ActivityExtremelyActive:  ActivityMultiplierExtremelyActive,
	}
	return multipliers[activityLevel]
}

// GetGoalModifier returns the modifier for given goal
func GetGoalModifier(goal string) float64 {
	modifiers := map[string]float64{
		GoalLoseWeight:     GoalModifierLoseWeight,
		GoalMaintainWeight: GoalModifierMaintainWeight,
		GoalGainWeight:     GoalModifierGainWeight,
	}
	return modifiers[goal]
}
