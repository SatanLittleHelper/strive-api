package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aleksandr/strive-api/internal/models"
)

type CalorieRepository interface {
	SaveOrUpdate(ctx context.Context, calculation *models.CalorieCalculation) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.CalorieCalculation, error)
}

type calorieRepository struct {
	pool *pgxpool.Pool
}

func NewCalorieRepository(pool *pgxpool.Pool) CalorieRepository {
	return &calorieRepository{
		pool: pool,
	}
}

func (r *calorieRepository) SaveOrUpdate(ctx context.Context, calculation *models.CalorieCalculation) error {
	query := `
		INSERT INTO calorie_calculations (
			id, user_id, gender, age, height, weight, activity_level, goal,
			bmr, tdee, target_calories, formula,
			protein_grams, protein_percentage, fat_grams, fat_percentage,
			carbs_grams, carbs_percentage, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12,
			$13, $14, $15, $16,
			$17, $18, $19, $20
		)
		ON CONFLICT (user_id) DO UPDATE SET
			gender = EXCLUDED.gender,
			age = EXCLUDED.age,
			height = EXCLUDED.height,
			weight = EXCLUDED.weight,
			activity_level = EXCLUDED.activity_level,
			goal = EXCLUDED.goal,
			bmr = EXCLUDED.bmr,
			tdee = EXCLUDED.tdee,
			target_calories = EXCLUDED.target_calories,
			formula = EXCLUDED.formula,
			protein_grams = EXCLUDED.protein_grams,
			protein_percentage = EXCLUDED.protein_percentage,
			fat_grams = EXCLUDED.fat_grams,
			fat_percentage = EXCLUDED.fat_percentage,
			carbs_grams = EXCLUDED.carbs_grams,
			carbs_percentage = EXCLUDED.carbs_percentage,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	if calculation.CreatedAt.IsZero() {
		calculation.CreatedAt = now
	}
	calculation.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query,
		calculation.ID,
		calculation.UserID,
		calculation.Gender,
		calculation.Age,
		calculation.Height,
		calculation.Weight,
		calculation.ActivityLevel,
		calculation.Goal,
		calculation.BMR,
		calculation.TDEE,
		calculation.TargetCalories,
		calculation.Formula,
		calculation.ProteinGrams,
		calculation.ProteinPercentage,
		calculation.FatGrams,
		calculation.FatPercentage,
		calculation.CarbsGrams,
		calculation.CarbsPercentage,
		calculation.CreatedAt,
		calculation.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save calorie calculation: %w", err)
	}

	return nil
}

func (r *calorieRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.CalorieCalculation, error) {
	query := `
		SELECT id, user_id, gender, age, height, weight, activity_level, goal,
			   bmr, tdee, target_calories, formula,
			   protein_grams, protein_percentage, fat_grams, fat_percentage,
			   carbs_grams, carbs_percentage, created_at, updated_at
		FROM calorie_calculations
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var calculation models.CalorieCalculation
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&calculation.ID,
		&calculation.UserID,
		&calculation.Gender,
		&calculation.Age,
		&calculation.Height,
		&calculation.Weight,
		&calculation.ActivityLevel,
		&calculation.Goal,
		&calculation.BMR,
		&calculation.TDEE,
		&calculation.TargetCalories,
		&calculation.Formula,
		&calculation.ProteinGrams,
		&calculation.ProteinPercentage,
		&calculation.FatGrams,
		&calculation.FatPercentage,
		&calculation.CarbsGrams,
		&calculation.CarbsPercentage,
		&calculation.CreatedAt,
		&calculation.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get calorie calculation: %w", err)
	}

	return &calculation, nil
}
