CREATE TABLE muscle_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wger_id INTEGER NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    name_en VARCHAR(100) NOT NULL DEFAULT '',
    is_front BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wger_id INTEGER NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wger_id INTEGER NOT NULL UNIQUE,
    wger_uuid VARCHAR(255) NOT NULL DEFAULT '',
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category INTEGER NOT NULL DEFAULT 0,
    language INTEGER NOT NULL DEFAULT 0,
    license INTEGER NOT NULL DEFAULT 0,
    license_author VARCHAR(255) NOT NULL DEFAULT '',
    creation_date DATE,
    cached_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE exercise_muscle_groups (
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    muscle_group_id UUID NOT NULL REFERENCES muscle_groups(id) ON DELETE CASCADE,
    is_primary BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (exercise_id, muscle_group_id, is_primary)
);

CREATE TABLE exercise_equipment (
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (exercise_id, equipment_id)
);

CREATE TABLE exercise_variations (
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    variation_exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (exercise_id, variation_exercise_id),
    CHECK (exercise_id != variation_exercise_id)
);

CREATE INDEX idx_exercises_wger_id ON exercises(wger_id);
CREATE INDEX idx_exercises_category ON exercises(category);
CREATE INDEX idx_exercises_name ON exercises(name);
CREATE INDEX idx_exercises_cached_at ON exercises(cached_at);
CREATE INDEX idx_exercises_expires_at ON exercises(expires_at);

CREATE INDEX idx_muscle_groups_wger_id ON muscle_groups(wger_id);
CREATE INDEX idx_muscle_groups_name ON muscle_groups(name);

CREATE INDEX idx_equipment_wger_id ON equipment(wger_id);
CREATE INDEX idx_equipment_name ON equipment(name);

CREATE INDEX idx_exercise_muscle_groups_exercise_id ON exercise_muscle_groups(exercise_id);
CREATE INDEX idx_exercise_muscle_groups_muscle_group_id ON exercise_muscle_groups(muscle_group_id);

CREATE INDEX idx_exercise_equipment_exercise_id ON exercise_equipment(exercise_id);
CREATE INDEX idx_exercise_equipment_equipment_id ON exercise_equipment(equipment_id);

CREATE INDEX idx_exercise_variations_exercise_id ON exercise_variations(exercise_id);
CREATE INDEX idx_exercise_variations_variation_id ON exercise_variations(variation_exercise_id);

