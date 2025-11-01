package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyRepository_CreateAPIKey(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	userRepo := NewUserRepository(db)
	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	// Create test user
	user, err := userRepo.CreateUser(ctx, "0x742d35cc6634c0532925a3b844bc9e7595f0beb0")
	require.NoError(t, err)

	t.Run("creates API key with valid request", func(t *testing.T) {
		req := APIKeyCreateRequest{
			UserID: user.ID,
			Name:   "Test Key",
			Scopes: []string{"read", "write"},
		}

		rawKey, response, err := repo.CreateAPIKey(ctx, req)

		require.NoError(t, err)
		assert.NotEmpty(t, rawKey)
		assert.Len(t, rawKey, 64) // 32 bytes hex-encoded
		assert.NotZero(t, response.ID)
		assert.NotEmpty(t, response.KeyHash)
		assert.Equal(t, "Test Key", response.Name)
		assert.Equal(t, []string{"read", "write"}, response.Scopes)
		assert.Nil(t, response.ExpiresAt)
		assert.False(t, response.CreatedAt.IsZero())

		// Verify key hash is different from raw key
		assert.NotEqual(t, rawKey, response.KeyHash)
		assert.Len(t, response.KeyHash, 64) // SHA256 hex-encoded
	})

	t.Run("creates API key with expiration", func(t *testing.T) {
		expiresIn := 24 * time.Hour
		req := APIKeyCreateRequest{
			UserID:    user.ID,
			Name:      "Expiring Key",
			Scopes:    []string{"read"},
			ExpiresIn: &expiresIn,
		}

		rawKey, response, err := repo.CreateAPIKey(ctx, req)

		require.NoError(t, err)
		assert.NotEmpty(t, rawKey)
		assert.NotNil(t, response.ExpiresAt)
		assert.True(t, response.ExpiresAt.After(time.Now()))
		assert.True(t, response.ExpiresAt.Before(time.Now().Add(25*time.Hour)))
	})

	t.Run("creates API key with empty scopes", func(t *testing.T) {
		req := APIKeyCreateRequest{
			UserID: user.ID,
			Name:   "No Scopes Key",
			Scopes: nil,
		}

		rawKey, response, err := repo.CreateAPIKey(ctx, req)

		require.NoError(t, err)
		assert.NotEmpty(t, rawKey)
		assert.NotNil(t, response.Scopes)
		assert.Len(t, response.Scopes, 0)
	})

	t.Run("generates unique keys", func(t *testing.T) {
		req := APIKeyCreateRequest{
			UserID: user.ID,
			Name:   "Key 1",
			Scopes: []string{"read"},
		}

		rawKey1, _, err := repo.CreateAPIKey(ctx, req)
		require.NoError(t, err)

		req.Name = "Key 2"
		rawKey2, _, err := repo.CreateAPIKey(ctx, req)
		require.NoError(t, err)

		assert.NotEqual(t, rawKey1, rawKey2)
	})

	t.Run("rejects request with missing user_id", func(t *testing.T) {
		req := APIKeyCreateRequest{
			Name:   "Invalid Key",
			Scopes: []string{"read"},
		}

		_, _, err := repo.CreateAPIKey(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user_id")
	})

	t.Run("rejects request with missing name", func(t *testing.T) {
		req := APIKeyCreateRequest{
			UserID: user.ID,
			Scopes: []string{"read"},
		}

		_, _, err := repo.CreateAPIKey(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("rejects request with non-existent user", func(t *testing.T) {
		req := APIKeyCreateRequest{
			UserID: 999999,
			Name:   "Invalid User Key",
			Scopes: []string{"read"},
		}

		_, _, err := repo.CreateAPIKey(ctx, req)
		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
		assert.Equal(t, "user", notFoundErr.Resource)
	})
}

func TestAPIKeyRepository_ValidateAPIKey(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	userRepo := NewUserRepository(db)
	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	// Create test user
	user, err := userRepo.CreateUser(ctx, "0x742d35cc6634c0532925a3b844bc9e7595f0beb1")
	require.NoError(t, err)

	t.Run("validates valid API key", func(t *testing.T) {
		req := APIKeyCreateRequest{
			UserID: user.ID,
			Name:   "Valid Key",
			Scopes: []string{"read", "write"},
		}

		rawKey, _, err := repo.CreateAPIKey(ctx, req)
		require.NoError(t, err)

		// Validate the key
		apiKey, err := repo.ValidateAPIKey(ctx, rawKey)
		require.NoError(t, err)

		assert.Equal(t, user.ID, apiKey.UserID)
		assert.Equal(t, "Valid Key", apiKey.Name)
		assert.Equal(t, []string{"read", "write"}, apiKey.Scopes)
		assert.Nil(t, apiKey.LastUsedAt)
		assert.Nil(t, apiKey.ExpiresAt)
	})

	t.Run("rejects invalid API key", func(t *testing.T) {
		_, err := repo.ValidateAPIKey(ctx, "invalid_key_that_does_not_exist")

		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
	})

	t.Run("rejects empty API key", func(t *testing.T) {
		_, err := repo.ValidateAPIKey(ctx, "")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("rejects expired API key", func(t *testing.T) {
		// Create key that expires in the past
		expiresIn := -1 * time.Hour
		req := APIKeyCreateRequest{
			UserID:    user.ID,
			Name:      "Expired Key",
			Scopes:    []string{"read"},
			ExpiresIn: &expiresIn,
		}

		rawKey, _, err := repo.CreateAPIKey(ctx, req)
		require.NoError(t, err)

		// Try to validate
		_, err = repo.ValidateAPIKey(ctx, rawKey)
		require.Error(t, err)
		var expiredErr *ExpiredError
		assert.True(t, errors.As(err, &expiredErr))
		assert.Equal(t, "api_key", expiredErr.Resource)
	})

	t.Run("validates key about to expire", func(t *testing.T) {
		// Create key that expires in 1 minute
		expiresIn := 1 * time.Minute
		req := APIKeyCreateRequest{
			UserID:    user.ID,
			Name:      "Soon Expired Key",
			Scopes:    []string{"read"},
			ExpiresIn: &expiresIn,
		}

		rawKey, _, err := repo.CreateAPIKey(ctx, req)
		require.NoError(t, err)

		// Should still be valid
		apiKey, err := repo.ValidateAPIKey(ctx, rawKey)
		require.NoError(t, err)
		assert.NotNil(t, apiKey)
	})
}

func TestAPIKeyRepository_GetAPIKey(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	userRepo := NewUserRepository(db)
	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	user, err := userRepo.CreateUser(ctx, "0x742d35cc6634c0532925a3b844bc9e7595f0beb2")
	require.NoError(t, err)

	t.Run("retrieves existing API key", func(t *testing.T) {
		req := APIKeyCreateRequest{
			UserID: user.ID,
			Name:   "Test Key",
			Scopes: []string{"read"},
		}

		_, response, err := repo.CreateAPIKey(ctx, req)
		require.NoError(t, err)

		// Get by ID
		apiKey, err := repo.GetAPIKey(ctx, response.ID)
		require.NoError(t, err)

		assert.Equal(t, response.ID, apiKey.ID)
		assert.Equal(t, user.ID, apiKey.UserID)
		assert.Equal(t, "Test Key", apiKey.Name)
	})

	t.Run("returns not found for non-existent key", func(t *testing.T) {
		_, err := repo.GetAPIKey(ctx, 999999)

		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
	})
}

func TestAPIKeyRepository_ListAPIKeys(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	userRepo := NewUserRepository(db)
	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	user1, err := userRepo.CreateUser(ctx, "0x742d35cc6634c0532925a3b844bc9e7595f0beb3")
	require.NoError(t, err)

	user2, err := userRepo.CreateUser(ctx, "0x742d35cc6634c0532925a3b844bc9e7595f0beb4")
	require.NoError(t, err)

	t.Run("lists all keys for user", func(t *testing.T) {
		// Create multiple keys for user1
		for i := 0; i < 3; i++ {
			req := APIKeyCreateRequest{
				UserID: user1.ID,
				Name:   "Key " + string(rune('A'+i)),
				Scopes: []string{"read"},
			}
			_, _, err := repo.CreateAPIKey(ctx, req)
			require.NoError(t, err)
			time.Sleep(10 * time.Millisecond) // Ensure different timestamps
		}

		// Create key for user2
		req := APIKeyCreateRequest{
			UserID: user2.ID,
			Name:   "User2 Key",
			Scopes: []string{"read"},
		}
		_, _, err := repo.CreateAPIKey(ctx, req)
		require.NoError(t, err)

		// List user1's keys
		keys, err := repo.ListAPIKeys(ctx, user1.ID)
		require.NoError(t, err)

		assert.Len(t, keys, 3)
		// Should be ordered by created_at DESC
		assert.Equal(t, "Key C", keys[0].Name)
		assert.Equal(t, "Key B", keys[1].Name)
		assert.Equal(t, "Key A", keys[2].Name)
	})

	t.Run("returns empty list for user with no keys", func(t *testing.T) {
		user3, err := userRepo.CreateUser(ctx, "0x742d35cc6634c0532925a3b844bc9e7595f0beb5")
		require.NoError(t, err)

		keys, err := repo.ListAPIKeys(ctx, user3.ID)
		require.NoError(t, err)

		assert.NotNil(t, keys)
		assert.Len(t, keys, 0)
	})
}

func TestAPIKeyRepository_UpdateLastUsed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	userRepo := NewUserRepository(db)
	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	user, err := userRepo.CreateUser(ctx, "0x742d35cc6634c0532925a3b844bc9e7595f0beb6")
	require.NoError(t, err)

	t.Run("updates last_used_at timestamp", func(t *testing.T) {
		req := APIKeyCreateRequest{
			UserID: user.ID,
			Name:   "Test Key",
			Scopes: []string{"read"},
		}

		rawKey, response, err := repo.CreateAPIKey(ctx, req)
		require.NoError(t, err)

		// Verify last_used_at is initially nil
		apiKey, err := repo.GetAPIKey(ctx, response.ID)
		require.NoError(t, err)
		assert.Nil(t, apiKey.LastUsedAt)

		// Update last used
		keyHash := HashAPIKey(rawKey)
		err = repo.UpdateLastUsed(ctx, keyHash)
		require.NoError(t, err)

		// Verify last_used_at is now set
		apiKey, err = repo.GetAPIKey(ctx, response.ID)
		require.NoError(t, err)
		assert.NotNil(t, apiKey.LastUsedAt)
		assert.True(t, apiKey.LastUsedAt.After(time.Now().Add(-1*time.Minute)))
	})

	t.Run("returns not found for invalid key hash", func(t *testing.T) {
		err := repo.UpdateLastUsed(ctx, "invalid_hash")

		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
	})
}

func TestAPIKeyRepository_DeleteAPIKey(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	userRepo := NewUserRepository(db)
	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	user, err := userRepo.CreateUser(ctx, "0x742d35cc6634c0532925a3b844bc9e7595f0beb7")
	require.NoError(t, err)

	t.Run("deletes existing API key", func(t *testing.T) {
		req := APIKeyCreateRequest{
			UserID: user.ID,
			Name:   "Delete Me",
			Scopes: []string{"read"},
		}

		_, response, err := repo.CreateAPIKey(ctx, req)
		require.NoError(t, err)

		// Delete the key
		err = repo.DeleteAPIKey(ctx, response.ID)
		require.NoError(t, err)

		// Verify it's deleted
		_, err = repo.GetAPIKey(ctx, response.ID)
		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
	})

	t.Run("returns not found for non-existent key", func(t *testing.T) {
		err := repo.DeleteAPIKey(ctx, 999999)

		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
	})
}

func TestAPIKeyRepository_RevokeExpiredKeys(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	userRepo := NewUserRepository(db)
	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	user, err := userRepo.CreateUser(ctx, "0x742d35cc6634c0532925a3b844bc9e7595f0beb8")
	require.NoError(t, err)

	t.Run("revokes expired keys", func(t *testing.T) {
		// Create expired key
		expiresIn := -1 * time.Hour
		req1 := APIKeyCreateRequest{
			UserID:    user.ID,
			Name:      "Expired Key 1",
			Scopes:    []string{"read"},
			ExpiresIn: &expiresIn,
		}
		_, _, err := repo.CreateAPIKey(ctx, req1)
		require.NoError(t, err)

		// Create another expired key
		req2 := APIKeyCreateRequest{
			UserID:    user.ID,
			Name:      "Expired Key 2",
			Scopes:    []string{"read"},
			ExpiresIn: &expiresIn,
		}
		_, _, err = repo.CreateAPIKey(ctx, req2)
		require.NoError(t, err)

		// Create valid key
		req3 := APIKeyCreateRequest{
			UserID: user.ID,
			Name:   "Valid Key",
			Scopes: []string{"read"},
		}
		_, _, err = repo.CreateAPIKey(ctx, req3)
		require.NoError(t, err)

		// Revoke expired keys
		count, err := repo.RevokeExpiredKeys(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, count)

		// Verify only valid key remains
		keys, err := repo.ListAPIKeys(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, keys, 1)
		assert.Equal(t, "Valid Key", keys[0].Name)
	})

	t.Run("returns zero when no expired keys", func(t *testing.T) {
		count, err := repo.RevokeExpiredKeys(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestHashAPIKey(t *testing.T) {
	t.Run("generates consistent hash", func(t *testing.T) {
		key := "test_api_key_12345"
		hash1 := HashAPIKey(key)
		hash2 := HashAPIKey(key)

		assert.Equal(t, hash1, hash2)
		assert.Len(t, hash1, 64) // SHA256 hex-encoded
	})

	t.Run("generates different hashes for different keys", func(t *testing.T) {
		hash1 := HashAPIKey("key1")
		hash2 := HashAPIKey("key2")

		assert.NotEqual(t, hash1, hash2)
	})
}

func TestGenerateAPIKey(t *testing.T) {
	t.Run("generates 64-character hex string", func(t *testing.T) {
		key, err := GenerateAPIKey()
		require.NoError(t, err)
		assert.Len(t, key, 64)
	})

	t.Run("generates unique keys", func(t *testing.T) {
		keys := make(map[string]bool)
		for i := 0; i < 100; i++ {
			key, err := GenerateAPIKey()
			require.NoError(t, err)
			assert.False(t, keys[key], "duplicate key generated")
			keys[key] = true
		}
	})
}
