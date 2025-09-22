-- Add theme column to users table
ALTER TABLE users ADD COLUMN theme VARCHAR(10) DEFAULT 'light' NOT NULL;

-- Add constraint to ensure only valid theme values
ALTER TABLE users ADD CONSTRAINT users_theme_check CHECK (theme IN ('light', 'dark'));
