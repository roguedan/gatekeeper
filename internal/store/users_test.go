package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("creates user with valid address", func(t *testing.T) {
		address := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"
		user, err := repo.CreateUser(ctx, address)

		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.Equal(t, "0x742d35cc6634c0532925a3b844bc9e7595f0beb0", user.Address) // lowercase
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	t.Run("normalizes address to lowercase", func(t *testing.T) {
		address := "0xABCDEF0123456789ABCDEF0123456789ABCDEF01"
		user, err := repo.CreateUser(ctx, address)

		require.NoError(t, err)
		assert.Equal(t, "0xabcdef0123456789abcdef0123456789abcdef01", user.Address)
	})

	t.Run("rejects duplicate address", func(t *testing.T) {
		address := "0x1234567890123456789012345678901234567890"
		_, err := repo.CreateUser(ctx, address)
		require.NoError(t, err)

		// Try to create again
		_, err = repo.CreateUser(ctx, address)
		require.Error(t, err)
		var dupErr *DuplicateError
		assert.True(t, errors.As(err, &dupErr))
		assert.Equal(t, "user", dupErr.Resource)
	})

	t.Run("rejects invalid address - missing 0x prefix", func(t *testing.T) {
		address := "742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"
		_, err := repo.CreateUser(ctx, address)

		require.Error(t, err)
		var addrErr *InvalidAddressError
		assert.True(t, errors.As(err, &addrErr))
	})

	t.Run("rejects invalid address - too short", func(t *testing.T) {
		address := "0x742d35Cc"
		_, err := repo.CreateUser(ctx, address)

		require.Error(t, err)
		var addrErr *InvalidAddressError
		assert.True(t, errors.As(err, &addrErr))
	})

	t.Run("rejects invalid address - invalid characters", func(t *testing.T) {
		address := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEbZ"
		_, err := repo.CreateUser(ctx, address)

		require.Error(t, err)
		var addrErr *InvalidAddressError
		assert.True(t, errors.As(err, &addrErr))
	})

	t.Run("rejects empty address", func(t *testing.T) {
		_, err := repo.CreateUser(ctx, "")

		require.Error(t, err)
		var addrErr *InvalidAddressError
		assert.True(t, errors.As(err, &addrErr))
	})
}

func TestUserRepository_GetUserByAddress(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("retrieves existing user", func(t *testing.T) {
		address := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1"
		created, err := repo.CreateUser(ctx, address)
		require.NoError(t, err)

		// Retrieve by address (different case)
		found, err := repo.GetUserByAddress(ctx, "0x742D35CC6634C0532925A3B844BC9E7595F0BEB1")
		require.NoError(t, err)

		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Address, found.Address)
		assert.WithinDuration(t, created.CreatedAt, found.CreatedAt, time.Second)
	})

	t.Run("returns not found for non-existent address", func(t *testing.T) {
		_, err := repo.GetUserByAddress(ctx, "0x0000000000000000000000000000000000000000")

		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
		assert.Equal(t, "user", notFoundErr.Resource)
	})

	t.Run("rejects invalid address", func(t *testing.T) {
		_, err := repo.GetUserByAddress(ctx, "invalid")

		require.Error(t, err)
		var addrErr *InvalidAddressError
		assert.True(t, errors.As(err, &addrErr))
	})
}

func TestUserRepository_GetUserByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("retrieves existing user", func(t *testing.T) {
		address := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb2"
		created, err := repo.CreateUser(ctx, address)
		require.NoError(t, err)

		found, err := repo.GetUserByID(ctx, created.ID)
		require.NoError(t, err)

		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Address, found.Address)
	})

	t.Run("returns not found for non-existent ID", func(t *testing.T) {
		_, err := repo.GetUserByID(ctx, 999999)

		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
		assert.Equal(t, "user", notFoundErr.Resource)
		assert.Equal(t, int64(999999), notFoundErr.ID)
	})
}

func TestUserRepository_UpdateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("updates user address", func(t *testing.T) {
		// Create initial user
		user, err := repo.CreateUser(ctx, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb3")
		require.NoError(t, err)
		originalUpdatedAt := user.UpdatedAt

		// Wait a bit to ensure timestamp changes
		time.Sleep(10 * time.Millisecond)

		// Update address
		user.Address = "0x1111111111111111111111111111111111111111"
		err = repo.UpdateUser(ctx, user)
		require.NoError(t, err)

		assert.Equal(t, "0x1111111111111111111111111111111111111111", user.Address)
		assert.True(t, user.UpdatedAt.After(originalUpdatedAt))

		// Verify in database
		found, err := repo.GetUserByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Address, found.Address)
	})

	t.Run("rejects nil user", func(t *testing.T) {
		err := repo.UpdateUser(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("rejects invalid address", func(t *testing.T) {
		user, err := repo.CreateUser(ctx, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb4")
		require.NoError(t, err)

		user.Address = "invalid"
		err = repo.UpdateUser(ctx, user)

		require.Error(t, err)
		var addrErr *InvalidAddressError
		assert.True(t, errors.As(err, &addrErr))
	})

	t.Run("rejects duplicate address", func(t *testing.T) {
		user1, err := repo.CreateUser(ctx, "0x2222222222222222222222222222222222222222")
		require.NoError(t, err)

		user2, err := repo.CreateUser(ctx, "0x3333333333333333333333333333333333333333")
		require.NoError(t, err)

		// Try to update user2 to have user1's address
		user2.Address = user1.Address
		err = repo.UpdateUser(ctx, user2)

		require.Error(t, err)
		var dupErr *DuplicateError
		assert.True(t, errors.As(err, &dupErr))
	})

	t.Run("returns not found for non-existent user", func(t *testing.T) {
		user := &User{
			ID:      999999,
			Address: "0x4444444444444444444444444444444444444444",
		}
		err := repo.UpdateUser(ctx, user)

		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
	})
}

func TestUserRepository_DeleteUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("deletes existing user", func(t *testing.T) {
		user, err := repo.CreateUser(ctx, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb5")
		require.NoError(t, err)

		err = repo.DeleteUser(ctx, user.ID)
		require.NoError(t, err)

		// Verify user is deleted
		_, err = repo.GetUserByID(ctx, user.ID)
		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
	})

	t.Run("returns not found for non-existent user", func(t *testing.T) {
		err := repo.DeleteUser(ctx, 999999)

		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
		assert.Equal(t, int64(999999), notFoundErr.ID)
	})
}

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name      string
		address   string
		wantValid bool
		wantAddr  string
	}{
		{
			name:      "valid address with mixed case",
			address:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
			wantValid: true,
			wantAddr:  "0x742d35cc6634c0532925a3b844bc9e7595f0beb0",
		},
		{
			name:      "valid address with lowercase",
			address:   "0x742d35cc6634c0532925a3b844bc9e7595f0beb0",
			wantValid: true,
			wantAddr:  "0x742d35cc6634c0532925a3b844bc9e7595f0beb0",
		},
		{
			name:      "valid address with uppercase",
			address:   "0x742D35CC6634C0532925A3B844BC9E7595F0BEB0",
			wantValid: true,
			wantAddr:  "0x742d35cc6634c0532925a3b844bc9e7595f0beb0",
		},
		{
			name:      "address with whitespace",
			address:   "  0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0  ",
			wantValid: true,
			wantAddr:  "0x742d35cc6634c0532925a3b844bc9e7595f0beb0",
		},
		{
			name:      "missing 0x prefix",
			address:   "742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
			wantValid: false,
		},
		{
			name:      "too short",
			address:   "0x742d35Cc",
			wantValid: false,
		},
		{
			name:      "too long",
			address:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb000",
			wantValid: false,
		},
		{
			name:      "invalid characters",
			address:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEbZ",
			wantValid: false,
		},
		{
			name:      "empty string",
			address:   "",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := validateAddress(tt.address)
			if tt.wantValid {
				require.NoError(t, err)
				assert.Equal(t, tt.wantAddr, addr)
			} else {
				require.Error(t, err)
				var addrErr *InvalidAddressError
				assert.True(t, errors.As(err, &addrErr))
			}
		})
	}
}
