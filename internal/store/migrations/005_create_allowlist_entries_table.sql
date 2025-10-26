-- Create allowlist_entries table for storing addresses in allowlists
CREATE TABLE IF NOT EXISTS allowlist_entries (
    id BIGSERIAL PRIMARY KEY,
    allowlist_id BIGINT NOT NULL REFERENCES allowlists(id) ON DELETE CASCADE,
    address VARCHAR(42) NOT NULL, -- Ethereum address
    added_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(allowlist_id, address) -- Prevent duplicate addresses in the same allowlist
);

-- Create indexes for efficient queries
CREATE INDEX idx_allowlist_entries_allowlist_id ON allowlist_entries(allowlist_id);
CREATE INDEX idx_allowlist_entries_address ON allowlist_entries(address);
CREATE INDEX idx_allowlist_entries_allowlist_address ON allowlist_entries(allowlist_id, address);
