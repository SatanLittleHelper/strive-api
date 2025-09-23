CREATE TABLE calorie_calculations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    gender VARCHAR(10) NOT NULL CHECK (gender IN ('male', 'female')),
    age INTEGER NOT NULL CHECK (age >= 15 AND age <= 120),
    height DECIMAL(5,2) NOT NULL CHECK (height >= 100 AND height <= 250),
    weight DECIMAL(5,2) NOT NULL CHECK (weight >= 30 AND weight <= 300),
    activity_level VARCHAR(20) NOT NULL CHECK (activity_level IN ('sedentary', 'lightly_active', 'moderately_active', 'very_active', 'extremely_active')),
    goal VARCHAR(20) NOT NULL CHECK (goal IN ('lose_weight', 'maintain_weight', 'gain_weight')),
    bmr INTEGER NOT NULL,
    tdee INTEGER NOT NULL,
    target_calories INTEGER NOT NULL,
    formula VARCHAR(10) NOT NULL DEFAULT 'mifflin',
    protein_grams INTEGER NOT NULL,
    protein_percentage DECIMAL(5,2) NOT NULL,
    fat_grams INTEGER NOT NULL,
    fat_percentage DECIMAL(5,2) NOT NULL,
    carbs_grams INTEGER NOT NULL,
    carbs_percentage DECIMAL(5,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id)
);

CREATE INDEX idx_calorie_calculations_user_id ON calorie_calculations(user_id);
CREATE INDEX idx_calorie_calculations_created_at ON calorie_calculations(created_at);
