package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aleksandr/strive-api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExerciseRepository interface {
	GetAll(ctx context.Context, filters *models.ExerciseFilters) (*models.ExerciseListResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Exercise, error)
	GetByWgerID(ctx context.Context, wgerID int) (*models.Exercise, error)
	GetMuscleGroups(ctx context.Context) ([]models.MuscleGroup, error)
	GetMuscleGroupByWgerID(ctx context.Context, wgerID int) (*models.MuscleGroup, error)
	GetEquipment(ctx context.Context) ([]models.Equipment, error)
	GetEquipmentByWgerID(ctx context.Context, wgerID int) (*models.Equipment, error)
	GetCacheStatus(ctx context.Context) (*models.CacheStatus, error)
	IsCacheValid(ctx context.Context) (bool, error)
	ClearCache(ctx context.Context) error
	SaveExercise(ctx context.Context, exercise *models.Exercise) error
	SaveMuscleGroup(ctx context.Context, muscleGroup *models.MuscleGroup) error
	SaveEquipment(ctx context.Context, equipment *models.Equipment) error
	SaveExerciseMuscleGroup(ctx context.Context, exerciseID, muscleGroupID uuid.UUID, isPrimary bool) error
	SaveExerciseEquipment(ctx context.Context, exerciseID, equipmentID uuid.UUID) error
	SaveExerciseVariation(ctx context.Context, exerciseID, variationID uuid.UUID) error
}

type exerciseRepository struct {
	pool *pgxpool.Pool
}

func NewExerciseRepository(pool *pgxpool.Pool) ExerciseRepository {
	return &exerciseRepository{
		pool: pool,
	}
}

func (r *exerciseRepository) GetAll(ctx context.Context, filters *models.ExerciseFilters) (*models.ExerciseListResponse, error) {
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if filters.MuscleGroupID != nil {
		whereConditions = append(whereConditions,
			fmt.Sprintf("e.id IN (SELECT exercise_id FROM exercise_muscle_groups WHERE muscle_group_id = $%d)", argIndex))
		args = append(args, *filters.MuscleGroupID)
		argIndex++
	}

	if filters.EquipmentID != nil {
		whereConditions = append(whereConditions,
			fmt.Sprintf("e.id IN (SELECT exercise_id FROM exercise_equipment WHERE equipment_id = $%d)", argIndex))
		args = append(args, *filters.EquipmentID)
		argIndex++
	}

	if filters.Category != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("e.category = $%d", argIndex))
		args = append(args, *filters.Category)
		argIndex++
	}

	if filters.Search != nil && *filters.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("e.name ILIKE $%d", argIndex))
		args = append(args, "%"+*filters.Search+"%")
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT e.id)
		FROM exercises e
		%s
	`, whereClause)

	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count exercises: %w", err)
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := (filters.Page - 1) * limit

	query := fmt.Sprintf(`
		SELECT DISTINCT e.id, e.wger_id, e.wger_uuid, e.name, e.description,
		       e.category, e.language, e.license, e.license_author,
		       e.creation_date, e.cached_at, e.expires_at, e.created_at, e.updated_at
		FROM exercises e
		%s
		ORDER BY e.name
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get exercises: %w", err)
	}
	defer rows.Close()

	exercises := []models.Exercise{}
	for rows.Next() {
		exercise := models.Exercise{}
		err := rows.Scan(
			&exercise.ID,
			&exercise.WgerID,
			&exercise.WgerUUID,
			&exercise.Name,
			&exercise.Description,
			&exercise.Category,
			&exercise.Language,
			&exercise.License,
			&exercise.LicenseAuthor,
			&exercise.CreationDate,
			&exercise.CachedAt,
			&exercise.ExpiresAt,
			&exercise.CreatedAt,
			&exercise.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan exercise: %w", err)
		}

		muscleGroups, err := r.getMuscleGroupsForExercise(ctx, exercise.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get muscle groups for exercise: %w", err)
		}
		exercise.MuscleGroups = muscleGroups

		equipment, err := r.getEquipmentForExercise(ctx, exercise.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get equipment for exercise: %w", err)
		}
		exercise.Equipment = equipment

		variations, err := r.getVariationsForExercise(ctx, exercise.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get variations for exercise: %w", err)
		}
		exercise.Variations = variations

		exercises = append(exercises, exercise)
	}

	return &models.ExerciseListResponse{
		Exercises: exercises,
		Total:     total,
		Page:      filters.Page,
		Limit:     limit,
	}, nil
}

func (r *exerciseRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Exercise, error) {
	query := `
		SELECT id, wger_id, wger_uuid, name, description,
		       category, language, license, license_author,
		       creation_date, cached_at, expires_at, created_at, updated_at
		FROM exercises
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	exercise := &models.Exercise{}
	err := row.Scan(
		&exercise.ID,
		&exercise.WgerID,
		&exercise.WgerUUID,
		&exercise.Name,
		&exercise.Description,
		&exercise.Category,
		&exercise.Language,
		&exercise.License,
		&exercise.LicenseAuthor,
		&exercise.CreationDate,
		&exercise.CachedAt,
		&exercise.ExpiresAt,
		&exercise.CreatedAt,
		&exercise.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get exercise by id: %w", err)
	}

	muscleGroups, err := r.getMuscleGroupsForExercise(ctx, exercise.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get muscle groups for exercise: %w", err)
	}
	exercise.MuscleGroups = muscleGroups

	equipment, err := r.getEquipmentForExercise(ctx, exercise.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment for exercise: %w", err)
	}
	exercise.Equipment = equipment

	variations, err := r.getVariationsForExercise(ctx, exercise.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get variations for exercise: %w", err)
	}
	exercise.Variations = variations

	return exercise, nil
}

func (r *exerciseRepository) GetByWgerID(ctx context.Context, wgerID int) (*models.Exercise, error) {
	query := `
		SELECT id, wger_id, wger_uuid, name, description,
		       category, language, license, license_author,
		       creation_date, cached_at, expires_at, created_at, updated_at
		FROM exercises
		WHERE wger_id = $1
	`

	row := r.pool.QueryRow(ctx, query, wgerID)
	exercise := &models.Exercise{}
	err := row.Scan(
		&exercise.ID,
		&exercise.WgerID,
		&exercise.WgerUUID,
		&exercise.Name,
		&exercise.Description,
		&exercise.Category,
		&exercise.Language,
		&exercise.License,
		&exercise.LicenseAuthor,
		&exercise.CreationDate,
		&exercise.CachedAt,
		&exercise.ExpiresAt,
		&exercise.CreatedAt,
		&exercise.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get exercise by wger_id: %w", err)
	}

	return exercise, nil
}

func (r *exerciseRepository) GetMuscleGroups(ctx context.Context) ([]models.MuscleGroup, error) {
	query := `
		SELECT id, wger_id, name, name_en, is_front, created_at, updated_at
		FROM muscle_groups
		ORDER BY name
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get muscle groups: %w", err)
	}
	defer rows.Close()

	muscleGroups := []models.MuscleGroup{}
	for rows.Next() {
		mg := models.MuscleGroup{}
		err := rows.Scan(&mg.ID, &mg.WgerID, &mg.Name, &mg.NameEn, &mg.IsFront, &mg.CreatedAt, &mg.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan muscle group: %w", err)
		}
		muscleGroups = append(muscleGroups, mg)
	}

	return muscleGroups, nil
}

func (r *exerciseRepository) GetMuscleGroupByWgerID(ctx context.Context, wgerID int) (*models.MuscleGroup, error) {
	query := `
		SELECT id, wger_id, name, name_en, is_front, created_at, updated_at
		FROM muscle_groups
		WHERE wger_id = $1
	`

	row := r.pool.QueryRow(ctx, query, wgerID)
	mg := &models.MuscleGroup{}
	err := row.Scan(&mg.ID, &mg.WgerID, &mg.Name, &mg.NameEn, &mg.IsFront, &mg.CreatedAt, &mg.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get muscle group by wger_id: %w", err)
	}

	return mg, nil
}

func (r *exerciseRepository) GetEquipment(ctx context.Context) ([]models.Equipment, error) {
	query := `
		SELECT id, wger_id, name, created_at, updated_at
		FROM equipment
		ORDER BY name
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}
	defer rows.Close()

	equipment := []models.Equipment{}
	for rows.Next() {
		eq := models.Equipment{}
		err := rows.Scan(&eq.ID, &eq.WgerID, &eq.Name, &eq.CreatedAt, &eq.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan equipment: %w", err)
		}
		equipment = append(equipment, eq)
	}

	return equipment, nil
}

func (r *exerciseRepository) GetEquipmentByWgerID(ctx context.Context, wgerID int) (*models.Equipment, error) {
	query := `
		SELECT id, wger_id, name, created_at, updated_at
		FROM equipment
		WHERE wger_id = $1
	`

	row := r.pool.QueryRow(ctx, query, wgerID)
	eq := &models.Equipment{}
	err := row.Scan(&eq.ID, &eq.WgerID, &eq.Name, &eq.CreatedAt, &eq.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment by wger_id: %w", err)
	}

	return eq, nil
}

func (r *exerciseRepository) GetCacheStatus(ctx context.Context) (*models.CacheStatus, error) {
	query := `
		SELECT 
			MAX(e.cached_at) as last_updated,
			COUNT(DISTINCT e.id) as total_exercises,
			COUNT(DISTINCT mg.id) as total_muscles,
			COUNT(DISTINCT eq.id) as total_equipment,
			MIN(e.expires_at) as expires_at
		FROM exercises e
		LEFT JOIN muscle_groups mg ON true
		LEFT JOIN equipment eq ON true
	`

	row := r.pool.QueryRow(ctx, query)
	status := &models.CacheStatus{}
	var lastUpdated *time.Time
	var expiresAt *time.Time

	err := row.Scan(&lastUpdated, &status.TotalExercises, &status.TotalMuscles, &status.TotalEquipment, &expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache status: %w", err)
	}

	if lastUpdated != nil {
		status.LastUpdated = *lastUpdated
	}
	if expiresAt != nil {
		status.ExpiresAt = *expiresAt
		status.IsValid = time.Now().Before(*expiresAt)
	}

	return status, nil
}

func (r *exerciseRepository) IsCacheValid(ctx context.Context) (bool, error) {
	query := `
		SELECT COUNT(*) > 0 AND MIN(expires_at) > NOW()
		FROM exercises
	`

	var isValid bool
	err := r.pool.QueryRow(ctx, query).Scan(&isValid)
	if err != nil {
		return false, fmt.Errorf("failed to check cache validity: %w", err)
	}

	return isValid, nil
}

func (r *exerciseRepository) ClearCache(ctx context.Context) error {
	queries := []string{
		"DELETE FROM exercise_variations",
		"DELETE FROM exercise_equipment",
		"DELETE FROM exercise_muscle_groups",
		"DELETE FROM exercises",
		"DELETE FROM equipment",
		"DELETE FROM muscle_groups",
	}

	for _, query := range queries {
		_, err := r.pool.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}
	}

	return nil
}

func (r *exerciseRepository) SaveExercise(ctx context.Context, exercise *models.Exercise) error {
	query := `
		INSERT INTO exercises (
			id, wger_id, wger_uuid, name, description,
			category, language, license, license_author,
			creation_date, cached_at, expires_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
		ON CONFLICT (wger_id) DO UPDATE SET
			wger_uuid = EXCLUDED.wger_uuid,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			category = EXCLUDED.category,
			language = EXCLUDED.language,
			license = EXCLUDED.license,
			license_author = EXCLUDED.license_author,
			creation_date = EXCLUDED.creation_date,
			cached_at = EXCLUDED.cached_at,
			expires_at = EXCLUDED.expires_at,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		exercise.ID,
		exercise.WgerID,
		exercise.WgerUUID,
		exercise.Name,
		exercise.Description,
		exercise.Category,
		exercise.Language,
		exercise.License,
		exercise.LicenseAuthor,
		exercise.CreationDate,
		exercise.CachedAt,
		exercise.ExpiresAt,
		exercise.CreatedAt,
		exercise.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save exercise: %w", err)
	}

	return nil
}

func (r *exerciseRepository) SaveMuscleGroup(ctx context.Context, muscleGroup *models.MuscleGroup) error {
	query := `
		INSERT INTO muscle_groups (id, wger_id, name, name_en, is_front, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (wger_id) DO UPDATE SET
			name = EXCLUDED.name,
			name_en = EXCLUDED.name_en,
			is_front = EXCLUDED.is_front,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		muscleGroup.ID,
		muscleGroup.WgerID,
		muscleGroup.Name,
		muscleGroup.NameEn,
		muscleGroup.IsFront,
		muscleGroup.CreatedAt,
		muscleGroup.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save muscle group: %w", err)
	}

	return nil
}

func (r *exerciseRepository) SaveEquipment(ctx context.Context, equipment *models.Equipment) error {
	query := `
		INSERT INTO equipment (id, wger_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (wger_id) DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		equipment.ID,
		equipment.WgerID,
		equipment.Name,
		equipment.CreatedAt,
		equipment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save equipment: %w", err)
	}

	return nil
}

func (r *exerciseRepository) SaveExerciseMuscleGroup(ctx context.Context, exerciseID, muscleGroupID uuid.UUID, isPrimary bool) error {
	query := `
		INSERT INTO exercise_muscle_groups (exercise_id, muscle_group_id, is_primary, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (exercise_id, muscle_group_id, is_primary) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, exerciseID, muscleGroupID, isPrimary, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save exercise muscle group: %w", err)
	}

	return nil
}

func (r *exerciseRepository) SaveExerciseEquipment(ctx context.Context, exerciseID, equipmentID uuid.UUID) error {
	query := `
		INSERT INTO exercise_equipment (exercise_id, equipment_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (exercise_id, equipment_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, exerciseID, equipmentID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save exercise equipment: %w", err)
	}

	return nil
}

func (r *exerciseRepository) SaveExerciseVariation(ctx context.Context, exerciseID, variationID uuid.UUID) error {
	query := `
		INSERT INTO exercise_variations (exercise_id, variation_exercise_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (exercise_id, variation_exercise_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, exerciseID, variationID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save exercise variation: %w", err)
	}

	return nil
}

func (r *exerciseRepository) getMuscleGroupsForExercise(ctx context.Context, exerciseID uuid.UUID) ([]models.MuscleGroup, error) {
	query := `
		SELECT mg.id, mg.wger_id, mg.name, mg.name_en, mg.is_front, mg.created_at, mg.updated_at
		FROM muscle_groups mg
		JOIN exercise_muscle_groups emg ON mg.id = emg.muscle_group_id
		WHERE emg.exercise_id = $1
		ORDER BY emg.is_primary DESC, mg.name
	`

	rows, err := r.pool.Query(ctx, query, exerciseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get muscle groups for exercise: %w", err)
	}
	defer rows.Close()

	muscleGroups := []models.MuscleGroup{}
	for rows.Next() {
		mg := models.MuscleGroup{}
		err := rows.Scan(&mg.ID, &mg.WgerID, &mg.Name, &mg.NameEn, &mg.IsFront, &mg.CreatedAt, &mg.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan muscle group: %w", err)
		}
		muscleGroups = append(muscleGroups, mg)
	}

	return muscleGroups, nil
}

func (r *exerciseRepository) getEquipmentForExercise(ctx context.Context, exerciseID uuid.UUID) ([]models.Equipment, error) {
	query := `
		SELECT e.id, e.wger_id, e.name, e.created_at, e.updated_at
		FROM equipment e
		JOIN exercise_equipment ee ON e.id = ee.equipment_id
		WHERE ee.exercise_id = $1
		ORDER BY e.name
	`

	rows, err := r.pool.Query(ctx, query, exerciseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment for exercise: %w", err)
	}
	defer rows.Close()

	equipment := []models.Equipment{}
	for rows.Next() {
		eq := models.Equipment{}
		err := rows.Scan(&eq.ID, &eq.WgerID, &eq.Name, &eq.CreatedAt, &eq.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan equipment: %w", err)
		}
		equipment = append(equipment, eq)
	}

	return equipment, nil
}

func (r *exerciseRepository) getVariationsForExercise(ctx context.Context, exerciseID uuid.UUID) ([]models.Exercise, error) {
	query := `
		SELECT e.id, e.wger_id, e.wger_uuid, e.name, e.description,
		       e.category, e.language, e.license, e.license_author,
		       e.creation_date, e.cached_at, e.expires_at, e.created_at, e.updated_at
		FROM exercises e
		JOIN exercise_variations ev ON e.id = ev.variation_exercise_id
		WHERE ev.exercise_id = $1
		ORDER BY e.name
	`

	rows, err := r.pool.Query(ctx, query, exerciseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get variations for exercise: %w", err)
	}
	defer rows.Close()

	variations := []models.Exercise{}
	for rows.Next() {
		variation := models.Exercise{}
		err := rows.Scan(
			&variation.ID,
			&variation.WgerID,
			&variation.WgerUUID,
			&variation.Name,
			&variation.Description,
			&variation.Category,
			&variation.Language,
			&variation.License,
			&variation.LicenseAuthor,
			&variation.CreationDate,
			&variation.CachedAt,
			&variation.ExpiresAt,
			&variation.CreatedAt,
			&variation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan variation exercise: %w", err)
		}
		variations = append(variations, variation)
	}

	return variations, nil
}
