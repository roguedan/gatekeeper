package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// APIKey represents an API key in the database
type APIKey struct {
	ID         int64      `db:"id"`
	UserID     int64      `db:"user_id"`
	KeyHash    string     `db:"key_hash"`
	Name       string     `db:"name"`
	Scopes     []string   `db:"scopes"`
	LastUsedAt *time.Time `db:"last_used_at"`
	ExpiresAt  *time.Time `db:"expires_at"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
}

// APIKeyCreateRequest represents a request to create a new API key
type APIKeyCreateRequest struct {
	UserID    int64
	Name      string
	Scopes    []string
	ExpiresIn *time.Duration
}

// APIKeyResponse represents the response when creating an API key
type APIKeyResponse struct {
	ID        int64      `json:"id"`
	KeyHash   string     `json:"key_hash"`
	Name      string     `json:"name"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// APIKeyRepository handles database operations for API keys
type APIKeyRepository struct {
	db *DB
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// GenerateAPIKey generates a new cryptographically secure API key
// Returns the raw key (hex-encoded, 64 characters)
func GenerateAPIKey() (string, error) {
	// Generate 32 bytes (256 bits) of entropy
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}

	// Hex encode to get 64-character string
	return hex.EncodeToString(keyBytes), nil
}

// HashAPIKey creates a SHA256 hash of an API key
func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// CreateAPIKey generates a new API key, hashes it, and stores it in the database
func (r *APIKeyRepository) CreateAPIKey(ctx context.Context, req APIKeyCreateRequest) (string, *APIKeyResponse, error) {
	// Validate request
	if req.UserID == 0 {
		return "", nil, fmt.Errorf("user_id is required")
	}
	if req.Name == "" {
		return "", nil, fmt.Errorf("name is required")
	}
	if req.Scopes == nil {
		req.Scopes = []string{}
	}

	// Generate raw key
	rawKey, err := GenerateAPIKey()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Hash the key for storage
	keyHash := HashAPIKey(rawKey)

	// Calculate expiration if specified
	var expiresAt *time.Time
	if req.ExpiresIn != nil {
		expiry := time.Now().Add(*req.ExpiresIn)
		expiresAt = &expiry
	}

	// Insert into database
	query := `
		INSERT INTO api_keys (user_id, key_hash, name, scopes, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, key_hash, name, scopes, expires_at, created_at
	`

	var response APIKeyResponse
	err = r.db.QueryRowContext(
		ctx,
		query,
		req.UserID,
		keyHash,
		req.Name,
		pq.Array(req.Scopes),
		expiresAt,
	).Scan(
		&response.ID,
		&response.KeyHash,
		&response.Name,
		pq.Array(&response.Scopes),
		&response.ExpiresAt,
		&response.CreatedAt,
	)
	if err != nil {
		// Check for foreign key violation (user doesn't exist)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			return "", nil, &NotFoundError{
				Resource: "user",
				ID:       req.UserID,
			}
		}
		// Check for duplicate key violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return "", nil, &DuplicateError{
				Resource: "api_key",
				Field:    "key_hash",
				Value:    keyHash,
			}
		}
		return "", nil, fmt.Errorf("failed to insert API key: %w", err)
	}

	// Return both the raw key and the response
	// The raw key should ONLY be shown to the user once at creation time
	return rawKey, &response, nil
}

// ValidateAPIKey verifies an API key and returns the associated key metadata
func (r *APIKeyRepository) ValidateAPIKey(ctx context.Context, rawKey string) (*APIKey, error) {
	if rawKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	// Hash the provided key
	keyHash := HashAPIKey(rawKey)

	// Look up the key in the database
	query := `
		SELECT id, user_id, key_hash, name, scopes, last_used_at, expires_at, created_at, updated_at
		FROM api_keys
		WHERE key_hash = $1
	`

	apiKey := &APIKey{}
	err := r.db.QueryRowContext(ctx, query, keyHash).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.KeyHash,
		&apiKey.Name,
		pq.Array(&apiKey.Scopes),
		&apiKey.LastUsedAt,
		&apiKey.ExpiresAt,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Fail closed - don't reveal whether key exists
			return nil, &NotFoundError{
				Resource: "api_key",
				ID:       "***",
			}
		}
		return nil, fmt.Errorf("failed to query API key: %w", err)
	}

	// Check if key is expired
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, &ExpiredError{
			Resource: "api_key",
			ID:       apiKey.ID,
		}
	}

	return apiKey, nil
}

// GetAPIKey retrieves an API key by ID (metadata only, no raw key)
func (r *APIKeyRepository) GetAPIKey(ctx context.Context, id int64) (*APIKey, error) {
	var apiKey APIKey
	query := `
		SELECT id, user_id, key_hash, name, scopes, last_used_at, expires_at, created_at, updated_at
		FROM api_keys
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.KeyHash,
		&apiKey.Name,
		pq.Array(&apiKey.Scopes),
		&apiKey.LastUsedAt,
		&apiKey.ExpiresAt,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &NotFoundError{
				Resource: "api_key",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to query API key: %w", err)
	}

	return &apiKey, nil
}

// ListAPIKeys returns all API keys for a user
func (r *APIKeyRepository) ListAPIKeys(ctx context.Context, userID int64) ([]APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, name, scopes, last_used_at, expires_at, created_at, updated_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var key APIKey
		err := rows.Scan(
			&key.ID,
			&key.UserID,
			&key.KeyHash,
			&key.Name,
			pq.Array(&key.Scopes),
			&key.LastUsedAt,
			&key.ExpiresAt,
			&key.CreatedAt,
			&key.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		keys = append(keys, key)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	// Return empty slice instead of nil if no keys found
	if keys == nil {
		keys = []APIKey{}
	}

	return keys, nil
}

// GetAPIKeyByID retrieves an API key by its ID
func (r *APIKeyRepository) GetAPIKeyByID(ctx context.Context, id int64) (*APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, name, scopes, last_used_at, expires_at, created_at, updated_at
		FROM api_keys
		WHERE id = $1
	`

	apiKey := &APIKey{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.KeyHash,
		&apiKey.Name,
		pq.Array(&apiKey.Scopes),
		&apiKey.LastUsedAt,
		&apiKey.ExpiresAt,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &NotFoundError{
				Resource: "api_key",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to query API key: %w", err)
	}

	return apiKey, nil
}

// DeleteAPIKey deletes an API key (revokes it)
func (r *APIKeyRepository) DeleteAPIKey(ctx context.Context, id int64) error {
	query := `DELETE FROM api_keys WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &NotFoundError{
			Resource: "api_key",
			ID:       id,
		}
	}

	return nil
}

// UpdateLastUsed updates the last_used_at timestamp for an API key
func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, keyHash string) error {
	query := `
		UPDATE api_keys
		SET last_used_at = CURRENT_TIMESTAMP
		WHERE key_hash = $1
	`

	result, err := r.db.ExecContext(ctx, query, keyHash)
	if err != nil {
		return fmt.Errorf("failed to update last used timestamp: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &NotFoundError{
			Resource: "api_key",
			ID:       keyHash,
		}
	}

	return nil
}

// RevokeExpiredKeys deletes all expired API keys and returns the count
func (r *APIKeyRepository) RevokeExpiredKeys(ctx context.Context) (int, error) {
	query := `
		DELETE FROM api_keys
		WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP
	`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to revoke expired keys: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to check rows affected: %w", err)
	}

	return int(rowsAffected), nil
}
