CREATE TABLE IF NOT EXISTS exercise_variations (
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    variation_exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (exercise_id, variation_exercise_id),
    CHECK (exercise_id != variation_exercise_id)
);

CREATE INDEX IF NOT EXISTS idx_exercise_variations_exercise_id ON exercise_variations(exercise_id);
CREATE INDEX IF NOT EXISTS idx_exercise_variations_variation_id ON exercise_variations(variation_exercise_id);

