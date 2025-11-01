-- Create allowlists table for managing address allowlists
CREATE TABLE IF NOT EXISTS allowlists (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on name for fast lookups
CREATE INDEX IF NOT EXISTS idx_allowlists_name ON allowlists(name);
