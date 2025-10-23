package models

import (
	"time"

	"github.com/google/uuid"
)

type MuscleGroup struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ExerciseDBID int       `json:"-" db:"exercise_db_id"`
	Name         string    `json:"name" db:"name"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Equipment struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ExerciseDBID int       `json:"-" db:"exercise_db_id"`
	Name         string    `json:"name" db:"name"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Exercise struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	ExerciseDBID  int           `json:"-" db:"exercise_db_id"`
	Name          string        `json:"name" db:"name"`
	Description   *string       `json:"description,omitempty" db:"description"`
	Instructions  *string       `json:"instructions,omitempty" db:"instructions"`
	Tips          *string       `json:"tips,omitempty" db:"tips"`
	Category      *int          `json:"category,omitempty" db:"category"`
	Language      *int          `json:"language,omitempty" db:"language"`
	License       *int          `json:"license,omitempty" db:"license"`
	LicenseAuthor *string       `json:"license_author,omitempty" db:"license_author"`
	Status        *string       `json:"status,omitempty" db:"status"`
	NameOriginal  *string       `json:"name_original,omitempty" db:"name_original"`
	CreationDate  *time.Time    `json:"creation_date,omitempty" db:"creation_date"`
	UUID          *string       `json:"uuid,omitempty" db:"uuid"`
	CachedAt      time.Time     `json:"cached_at" db:"cached_at"`
	ExpiresAt     time.Time     `json:"expires_at" db:"expires_at"`
	MuscleGroups  []MuscleGroup `json:"muscle_groups,omitempty"`
	Equipment     []Equipment   `json:"equipment,omitempty"`
	Alternatives  []Exercise    `json:"alternatives,omitempty"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

type ExerciseListResponse struct {
	Exercises []Exercise `json:"exercises"`
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	Limit     int        `json:"limit"`
}

type ExerciseFilters struct {
	MuscleGroupID *uuid.UUID `json:"muscle_group_id,omitempty"`
	EquipmentID   *uuid.UUID `json:"equipment_id,omitempty"`
	Category      *int       `json:"category,omitempty"`
	Search        *string    `json:"search,omitempty"`
	Page          int        `json:"page"`
	Limit         int        `json:"limit"`
}

type CacheStatus struct {
	LastUpdated    time.Time `json:"last_updated"`
	TotalExercises int       `json:"total_exercises"`
	TotalMuscles   int       `json:"total_muscles"`
	TotalEquipment int       `json:"total_equipment"`
	IsValid        bool      `json:"is_valid"`
	ExpiresAt      time.Time `json:"expires_at"`
}
