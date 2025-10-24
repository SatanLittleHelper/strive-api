CREATE TABLE IF NOT EXISTS muscle_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wger_id INTEGER NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    name_en VARCHAR(100) NOT NULL DEFAULT '',
    is_front BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'muscle_groups' AND column_name = 'exercise_db_id') THEN
        ALTER TABLE muscle_groups RENAME COLUMN exercise_db_id TO wger_id;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'muscle_groups' AND column_name = 'name_en') THEN
        ALTER TABLE muscle_groups ADD COLUMN name_en VARCHAR(100) NOT NULL DEFAULT '';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'muscle_groups' AND column_name = 'is_front') THEN
        ALTER TABLE muscle_groups ADD COLUMN is_front BOOLEAN NOT NULL DEFAULT true;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wger_id INTEGER NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'equipment' AND column_name = 'exercise_db_id') THEN
        ALTER TABLE equipment RENAME COLUMN exercise_db_id TO wger_id;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS exercises (
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

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercises' AND column_name = 'exercise_db_id') THEN
        ALTER TABLE exercises RENAME COLUMN exercise_db_id TO wger_id;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercises' AND column_name = 'wger_uuid') THEN
        ALTER TABLE exercises ADD COLUMN wger_uuid VARCHAR(255) NOT NULL DEFAULT '';
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercises' AND column_name = 'instructions') THEN
        ALTER TABLE exercises DROP COLUMN instructions;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercises' AND column_name = 'tips') THEN
        ALTER TABLE exercises DROP COLUMN tips;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercises' AND column_name = 'status') THEN
        ALTER TABLE exercises DROP COLUMN status;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercises' AND column_name = 'name_original') THEN
        ALTER TABLE exercises DROP COLUMN name_original;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercises' AND column_name = 'uuid') THEN
        ALTER TABLE exercises DROP COLUMN uuid;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercises' AND column_name = 'description') THEN
        ALTER TABLE exercises ALTER COLUMN description TYPE TEXT;
        ALTER TABLE exercises ALTER COLUMN description SET NOT NULL;
        ALTER TABLE exercises ALTER COLUMN description SET DEFAULT '';
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercises' AND column_name = 'license_author') THEN
        ALTER TABLE exercises ALTER COLUMN license_author TYPE VARCHAR(255);
        ALTER TABLE exercises ALTER COLUMN license_author SET NOT NULL;
        ALTER TABLE exercises ALTER COLUMN license_author SET DEFAULT '';
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS exercise_muscle_groups (
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    muscle_group_id UUID NOT NULL REFERENCES muscle_groups(id) ON DELETE CASCADE,
    is_primary BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (exercise_id, muscle_group_id, is_primary)
);

CREATE TABLE IF NOT EXISTS exercise_equipment (
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (exercise_id, equipment_id)
);

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'exercise_alternatives') THEN
        ALTER TABLE exercise_alternatives RENAME TO exercise_variations;
        ALTER TABLE exercise_variations RENAME COLUMN alternative_exercise_id TO variation_exercise_id;
    ELSE
        CREATE TABLE IF NOT EXISTS exercise_variations (
            exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
            variation_exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            PRIMARY KEY (exercise_id, variation_exercise_id),
            CHECK (exercise_id != variation_exercise_id)
        );
    END IF;
END $$;

DROP INDEX IF EXISTS idx_exercises_exercise_db_id;
DROP INDEX IF EXISTS idx_muscle_groups_exercise_db_id;
DROP INDEX IF EXISTS idx_equipment_exercise_db_id;
DROP INDEX IF EXISTS idx_exercise_alternatives_exercise_id;
DROP INDEX IF EXISTS idx_exercise_alternatives_alternative_id;

CREATE INDEX IF NOT EXISTS idx_exercises_wger_id ON exercises(wger_id);
CREATE INDEX IF NOT EXISTS idx_exercises_category ON exercises(category);
CREATE INDEX IF NOT EXISTS idx_exercises_name ON exercises(name);
CREATE INDEX IF NOT EXISTS idx_exercises_cached_at ON exercises(cached_at);
CREATE INDEX IF NOT EXISTS idx_exercises_expires_at ON exercises(expires_at);

CREATE INDEX IF NOT EXISTS idx_muscle_groups_wger_id ON muscle_groups(wger_id);
CREATE INDEX IF NOT EXISTS idx_muscle_groups_name ON muscle_groups(name);

CREATE INDEX IF NOT EXISTS idx_equipment_wger_id ON equipment(wger_id);
CREATE INDEX IF NOT EXISTS idx_equipment_name ON equipment(name);

CREATE INDEX IF NOT EXISTS idx_exercise_muscle_groups_exercise_id ON exercise_muscle_groups(exercise_id);
CREATE INDEX IF NOT EXISTS idx_exercise_muscle_groups_muscle_group_id ON exercise_muscle_groups(muscle_group_id);

CREATE INDEX IF NOT EXISTS idx_exercise_equipment_exercise_id ON exercise_equipment(exercise_id);
CREATE INDEX IF NOT EXISTS idx_exercise_equipment_equipment_id ON exercise_equipment(equipment_id);

CREATE INDEX IF NOT EXISTS idx_exercise_variations_exercise_id ON exercise_variations(exercise_id);
CREATE INDEX IF NOT EXISTS idx_exercise_variations_variation_id ON exercise_variations(variation_exercise_id);
