-- Drop triggers
DROP TRIGGER IF EXISTS update_exercises_updated_at ON exercises;
DROP TRIGGER IF EXISTS update_equipment_updated_at ON equipment;
DROP TRIGGER IF EXISTS update_muscle_groups_updated_at ON muscle_groups;

-- Drop tables in reverse order
DROP TABLE IF EXISTS exercise_alternatives;
DROP TABLE IF EXISTS exercise_equipment;
DROP TABLE IF EXISTS exercise_muscle_groups;
DROP TABLE IF EXISTS exercises;
DROP TABLE IF EXISTS equipment;
DROP TABLE IF EXISTS muscle_groups;
