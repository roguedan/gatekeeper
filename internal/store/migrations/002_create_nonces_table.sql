-- Create nonces table for SIWE nonce lifecycle management
CREATE TABLE IF NOT EXISTS nonces (
    id BIGSERIAL PRIMARY KEY,
    nonce VARCHAR(255) NOT NULL UNIQUE,
    address VARCHAR(42), -- NULL if nonce not yet verified
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    verified_at TIMESTAMP WITH TIME ZONE,
    used_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_nonces_nonce ON nonces(nonce);
CREATE INDEX IF NOT EXISTS idx_nonces_address ON nonces(address);
CREATE INDEX IF NOT EXISTS idx_nonces_expires_at ON nonces(expires_at);
