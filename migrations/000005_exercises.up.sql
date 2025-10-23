-- Create muscle_groups table for caching ExerciseDB muscle groups
CREATE TABLE muscle_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exercise_db_id INTEGER NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create equipment table for caching ExerciseDB equipment
CREATE TABLE equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exercise_db_id INTEGER NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create exercises table for caching ExerciseDB exercises
CREATE TABLE exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exercise_db_id INTEGER NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    instructions TEXT,
    tips TEXT,
    category INTEGER,
    language INTEGER,
    license INTEGER,
    license_author VARCHAR(255),
    status VARCHAR(10),
    name_original VARCHAR(255),
    creation_date DATE,
    uuid VARCHAR(36),
    cached_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create exercise_muscle_groups junction table
CREATE TABLE exercise_muscle_groups (
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    muscle_group_id UUID NOT NULL REFERENCES muscle_groups(id) ON DELETE CASCADE,
    is_primary BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (exercise_id, muscle_group_id, is_primary)
);

-- Create exercise_equipment junction table
CREATE TABLE exercise_equipment (
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (exercise_id, equipment_id)
);

-- Create exercise_alternatives table for exercise alternatives
CREATE TABLE exercise_alternatives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    alternative_exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(exercise_id, alternative_exercise_id)
);

-- Create indexes for performance
CREATE INDEX idx_muscle_groups_exercise_db_id ON muscle_groups(exercise_db_id);
CREATE INDEX idx_muscle_groups_name ON muscle_groups(name);

CREATE INDEX idx_equipment_exercise_db_id ON equipment(exercise_db_id);
CREATE INDEX idx_equipment_name ON equipment(name);

CREATE INDEX idx_exercises_exercise_db_id ON exercises(exercise_db_id);
CREATE INDEX idx_exercises_name ON exercises(name);
CREATE INDEX idx_exercises_category ON exercises(category);
CREATE INDEX idx_exercises_cached_at ON exercises(cached_at);
CREATE INDEX idx_exercises_expires_at ON exercises(expires_at);

CREATE INDEX idx_exercise_muscle_groups_exercise_id ON exercise_muscle_groups(exercise_id);
CREATE INDEX idx_exercise_muscle_groups_muscle_group_id ON exercise_muscle_groups(muscle_group_id);
CREATE INDEX idx_exercise_muscle_groups_primary ON exercise_muscle_groups(is_primary);

CREATE INDEX idx_exercise_equipment_exercise_id ON exercise_equipment(exercise_id);
CREATE INDEX idx_exercise_equipment_equipment_id ON exercise_equipment(equipment_id);

CREATE INDEX idx_exercise_alternatives_exercise_id ON exercise_alternatives(exercise_id);
CREATE INDEX idx_exercise_alternatives_alternative_id ON exercise_alternatives(alternative_exercise_id);

-- Create triggers for updated_at
CREATE TRIGGER update_muscle_groups_updated_at BEFORE UPDATE ON muscle_groups
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_equipment_updated_at BEFORE UPDATE ON equipment
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_exercises_updated_at BEFORE UPDATE ON exercises
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
