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
	GetByExerciseDBID(ctx context.Context, exerciseDBID int) (*models.Exercise, error)
	GetMuscleGroups(ctx context.Context) ([]models.MuscleGroup, error)
	GetEquipment(ctx context.Context) ([]models.Equipment, error)
	GetCacheStatus(ctx context.Context) (*models.CacheStatus, error)
	IsCacheValid(ctx context.Context) (bool, error)
	ClearCache(ctx context.Context) error
	SaveExercise(ctx context.Context, exercise *models.Exercise) error
	SaveMuscleGroup(ctx context.Context, muscleGroup *models.MuscleGroup) error
	SaveEquipment(ctx context.Context, equipment *models.Equipment) error
	SaveExerciseMuscleGroup(ctx context.Context, exerciseID, muscleGroupID uuid.UUID, isPrimary bool) error
	SaveExerciseEquipment(ctx context.Context, exerciseID, equipmentID uuid.UUID) error
	SaveExerciseAlternative(ctx context.Context, exerciseID, alternativeID uuid.UUID) error
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
		SELECT DISTINCT e.id, e.exercise_db_id, e.name, e.description, e.instructions, e.tips,
		       e.category, e.language, e.license, e.license_author, e.status, e.name_original,
		       e.creation_date, e.uuid, e.cached_at, e.expires_at, e.created_at, e.updated_at
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
			&exercise.ExerciseDBID,
			&exercise.Name,
			&exercise.Description,
			&exercise.Instructions,
			&exercise.Tips,
			&exercise.Category,
			&exercise.Language,
			&exercise.License,
			&exercise.LicenseAuthor,
			&exercise.Status,
			&exercise.NameOriginal,
			&exercise.CreationDate,
			&exercise.UUID,
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

		alternatives, err := r.getAlternativesForExercise(ctx, exercise.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get alternatives for exercise: %w", err)
		}
		exercise.Alternatives = alternatives

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
		SELECT id, exercise_db_id, name, description, instructions, tips,
		       category, language, license, license_author, status, name_original,
		       creation_date, uuid, cached_at, expires_at, created_at, updated_at
		FROM exercises
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	exercise := &models.Exercise{}
	err := row.Scan(
		&exercise.ID,
		&exercise.ExerciseDBID,
		&exercise.Name,
		&exercise.Description,
		&exercise.Instructions,
		&exercise.Tips,
		&exercise.Category,
		&exercise.Language,
		&exercise.License,
		&exercise.LicenseAuthor,
		&exercise.Status,
		&exercise.NameOriginal,
		&exercise.CreationDate,
		&exercise.UUID,
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

	alternatives, err := r.getAlternativesForExercise(ctx, exercise.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alternatives for exercise: %w", err)
	}
	exercise.Alternatives = alternatives

	return exercise, nil
}

func (r *exerciseRepository) GetByExerciseDBID(ctx context.Context, exerciseDBID int) (*models.Exercise, error) {
	query := `
		SELECT id, exercise_db_id, name, description, instructions, tips,
		       category, language, license, license_author, status, name_original,
		       creation_date, uuid, cached_at, expires_at, created_at, updated_at
		FROM exercises
		WHERE exercise_db_id = $1
	`

	row := r.pool.QueryRow(ctx, query, exerciseDBID)
	exercise := &models.Exercise{}
	err := row.Scan(
		&exercise.ID,
		&exercise.ExerciseDBID,
		&exercise.Name,
		&exercise.Description,
		&exercise.Instructions,
		&exercise.Tips,
		&exercise.Category,
		&exercise.Language,
		&exercise.License,
		&exercise.LicenseAuthor,
		&exercise.Status,
		&exercise.NameOriginal,
		&exercise.CreationDate,
		&exercise.UUID,
		&exercise.CachedAt,
		&exercise.ExpiresAt,
		&exercise.CreatedAt,
		&exercise.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get exercise by exercise_db_id: %w", err)
	}

	return exercise, nil
}

func (r *exerciseRepository) GetMuscleGroups(ctx context.Context) ([]models.MuscleGroup, error) {
	query := `
		SELECT id, exercise_db_id, name, created_at, updated_at
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
		err := rows.Scan(&mg.ID, &mg.ExerciseDBID, &mg.Name, &mg.CreatedAt, &mg.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan muscle group: %w", err)
		}
		muscleGroups = append(muscleGroups, mg)
	}

	return muscleGroups, nil
}

func (r *exerciseRepository) GetEquipment(ctx context.Context) ([]models.Equipment, error) {
	query := `
		SELECT id, exercise_db_id, name, created_at, updated_at
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
		err := rows.Scan(&eq.ID, &eq.ExerciseDBID, &eq.Name, &eq.CreatedAt, &eq.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan equipment: %w", err)
		}
		equipment = append(equipment, eq)
	}

	return equipment, nil
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
		"DELETE FROM exercise_alternatives",
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
			id, exercise_db_id, name, description, instructions, tips,
			category, language, license, license_author, status, name_original,
			creation_date, uuid, cached_at, expires_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
		ON CONFLICT (exercise_db_id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			instructions = EXCLUDED.instructions,
			tips = EXCLUDED.tips,
			category = EXCLUDED.category,
			language = EXCLUDED.language,
			license = EXCLUDED.license,
			license_author = EXCLUDED.license_author,
			status = EXCLUDED.status,
			name_original = EXCLUDED.name_original,
			creation_date = EXCLUDED.creation_date,
			uuid = EXCLUDED.uuid,
			cached_at = EXCLUDED.cached_at,
			expires_at = EXCLUDED.expires_at,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		exercise.ID,
		exercise.ExerciseDBID,
		exercise.Name,
		exercise.Description,
		exercise.Instructions,
		exercise.Tips,
		exercise.Category,
		exercise.Language,
		exercise.License,
		exercise.LicenseAuthor,
		exercise.Status,
		exercise.NameOriginal,
		exercise.CreationDate,
		exercise.UUID,
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
		INSERT INTO muscle_groups (id, exercise_db_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (exercise_db_id) DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		muscleGroup.ID,
		muscleGroup.ExerciseDBID,
		muscleGroup.Name,
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
		INSERT INTO equipment (id, exercise_db_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (exercise_db_id) DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		equipment.ID,
		equipment.ExerciseDBID,
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

func (r *exerciseRepository) SaveExerciseAlternative(ctx context.Context, exerciseID, alternativeID uuid.UUID) error {
	query := `
		INSERT INTO exercise_alternatives (exercise_id, alternative_exercise_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (exercise_id, alternative_exercise_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, exerciseID, alternativeID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save exercise alternative: %w", err)
	}

	return nil
}

func (r *exerciseRepository) getMuscleGroupsForExercise(ctx context.Context, exerciseID uuid.UUID) ([]models.MuscleGroup, error) {
	query := `
		SELECT mg.id, mg.exercise_db_id, mg.name, mg.created_at, mg.updated_at
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
		err := rows.Scan(&mg.ID, &mg.ExerciseDBID, &mg.Name, &mg.CreatedAt, &mg.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan muscle group: %w", err)
		}
		muscleGroups = append(muscleGroups, mg)
	}

	return muscleGroups, nil
}

func (r *exerciseRepository) getEquipmentForExercise(ctx context.Context, exerciseID uuid.UUID) ([]models.Equipment, error) {
	query := `
		SELECT e.id, e.exercise_db_id, e.name, e.created_at, e.updated_at
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
		err := rows.Scan(&eq.ID, &eq.ExerciseDBID, &eq.Name, &eq.CreatedAt, &eq.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan equipment: %w", err)
		}
		equipment = append(equipment, eq)
	}

	return equipment, nil
}

func (r *exerciseRepository) getAlternativesForExercise(ctx context.Context, exerciseID uuid.UUID) ([]models.Exercise, error) {
	query := `
		SELECT e.id, e.exercise_db_id, e.name, e.description, e.instructions, e.tips,
		       e.category, e.language, e.license, e.license_author, e.status, e.name_original,
		       e.creation_date, e.uuid, e.cached_at, e.expires_at, e.created_at, e.updated_at
		FROM exercises e
		JOIN exercise_alternatives ea ON e.id = ea.alternative_exercise_id
		WHERE ea.exercise_id = $1
		ORDER BY e.name
	`

	rows, err := r.pool.Query(ctx, query, exerciseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alternatives for exercise: %w", err)
	}
	defer rows.Close()

	alternatives := []models.Exercise{}
	for rows.Next() {
		alt := models.Exercise{}
		err := rows.Scan(
			&alt.ID,
			&alt.ExerciseDBID,
			&alt.Name,
			&alt.Description,
			&alt.Instructions,
			&alt.Tips,
			&alt.Category,
			&alt.Language,
			&alt.License,
			&alt.LicenseAuthor,
			&alt.Status,
			&alt.NameOriginal,
			&alt.CreationDate,
			&alt.UUID,
			&alt.CachedAt,
			&alt.ExpiresAt,
			&alt.CreatedAt,
			&alt.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alternative exercise: %w", err)
		}
		alternatives = append(alternatives, alt)
	}

	return alternatives, nil
}
