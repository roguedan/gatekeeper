-- Create users table for storing wallet addresses and user information
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    address VARCHAR(42) NOT NULL UNIQUE, -- Ethereum address
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on address for fast lookups
CREATE INDEX IF NOT EXISTS idx_users_address ON users(address);
