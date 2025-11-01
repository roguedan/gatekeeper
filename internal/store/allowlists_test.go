package store

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllowlistRepository_CreateAllowlist(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewAllowlistRepository(db)
	ctx := context.Background()

	t.Run("creates allowlist with name and description", func(t *testing.T) {
		allowlist, err := repo.CreateAllowlist(ctx, "Premium Users", "Users with premium access")

		require.NoError(t, err)
		assert.NotZero(t, allowlist.ID)
		assert.Equal(t, "Premium Users", allowlist.Name)
		assert.Equal(t, "Users with premium access", allowlist.Description)
		assert.False(t, allowlist.CreatedAt.IsZero())
		assert.False(t, allowlist.UpdatedAt.IsZero())
	})

	t.Run("creates allowlist with empty description", func(t *testing.T) {
		allowlist, err := repo.CreateAllowlist(ctx, "Basic Users", "")

		require.NoError(t, err)
		assert.Equal(t, "Basic Users", allowlist.Name)
		assert.Equal(t, "", allowlist.Description)
	})

	t.Run("rejects duplicate name", func(t *testing.T) {
		name := "Duplicate List"
		_, err := repo.CreateAllowlist(ctx, name, "First")
		require.NoError(t, err)

		_, err = repo.CreateAllowlist(ctx, name, "Second")
		require.Error(t, err)
		var dupErr *DuplicateError
		assert.True(t, errors.As(err, &dupErr))
		assert.Equal(t, "allowlist", dupErr.Resource)
	})

	t.Run("rejects empty name", func(t *testing.T) {
		_, err := repo.CreateAllowlist(ctx, "", "Description")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})
}

func TestAllowlistRepository_GetAllowlist(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewAllowlistRepository(db)
	ctx := context.Background()

	t.Run("retrieves existing allowlist", func(t *testing.T) {
		created, err := repo.CreateAllowlist(ctx, "Test List", "Test Description")
		require.NoError(t, err)

		found, err := repo.GetAllowlist(ctx, created.ID)
		require.NoError(t, err)

		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Name, found.Name)
		assert.Equal(t, created.Description, found.Description)
	})

	t.Run("returns not found for non-existent allowlist", func(t *testing.T) {
		_, err := repo.GetAllowlist(ctx, 999999)

		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
		assert.Equal(t, "allowlist", notFoundErr.Resource)
	})
}

func TestAllowlistRepository_ListAllowlists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewAllowlistRepository(db)
	ctx := context.Background()

	t.Run("lists all allowlists with entry counts", func(t *testing.T) {
		// Create allowlists
		list1, err := repo.CreateAllowlist(ctx, "List 1", "First list")
		require.NoError(t, err)

		list2, err := repo.CreateAllowlist(ctx, "List 2", "Second list")
		require.NoError(t, err)

		// Add addresses to list1
		err = repo.AddAddress(ctx, list1.ID, "0x1111111111111111111111111111111111111111")
		require.NoError(t, err)
		err = repo.AddAddress(ctx, list1.ID, "0x2222222222222222222222222222222222222222")
		require.NoError(t, err)

		// Add address to list2
		err = repo.AddAddress(ctx, list2.ID, "0x3333333333333333333333333333333333333333")
		require.NoError(t, err)

		// List all
		lists, err := repo.ListAllowlists(ctx)
		require.NoError(t, err)

		assert.Len(t, lists, 2)

		// Find list1 and list2 in results
		var foundList1, foundList2 *AllowlistWithCount
		for i := range lists {
			if lists[i].ID == list1.ID {
				foundList1 = &lists[i]
			}
			if lists[i].ID == list2.ID {
				foundList2 = &lists[i]
			}
		}

		require.NotNil(t, foundList1)
		require.NotNil(t, foundList2)
		assert.Equal(t, int64(2), foundList1.EntryCount)
		assert.Equal(t, int64(1), foundList2.EntryCount)
	})

	t.Run("returns empty list when no allowlists exist", func(t *testing.T) {
		// Create fresh DB
		freshDB := setupTestDB(t)
		defer freshDB.Close()
		freshRepo := NewAllowlistRepository(freshDB)

		lists, err := freshRepo.ListAllowlists(ctx)
		require.NoError(t, err)

		assert.NotNil(t, lists)
		assert.Len(t, lists, 0)
	})
}

func TestAllowlistRepository_UpdateAllowlist(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewAllowlistRepository(db)
	ctx := context.Background()

	t.Run("updates allowlist name and description", func(t *testing.T) {
		allowlist, err := repo.CreateAllowlist(ctx, "Original Name", "Original Description")
		require.NoError(t, err)

		allowlist.Name = "Updated Name"
		allowlist.Description = "Updated Description"
		err = repo.UpdateAllowlist(ctx, allowlist)
		require.NoError(t, err)

		// Verify update
		found, err := repo.GetAllowlist(ctx, allowlist.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
		assert.Equal(t, "Updated Description", found.Description)
	})

	t.Run("rejects nil allowlist", func(t *testing.T) {
		err := repo.UpdateAllowlist(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("rejects empty name", func(t *testing.T) {
		allowlist, err := repo.CreateAllowlist(ctx, "Test", "Test")
		require.NoError(t, err)

		allowlist.Name = ""
		err = repo.UpdateAllowlist(ctx, allowlist)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("rejects duplicate name", func(t *testing.T) {
		list1, err := repo.CreateAllowlist(ctx, "List A", "A")
		require.NoError(t, err)

		list2, err := repo.CreateAllowlist(ctx, "List B", "B")
		require.NoError(t, err)

		// Try to rename list2 to list1's name
		list2.Name = list1.Name
		err = repo.UpdateAllowlist(ctx, list2)
		require.Error(t, err)
		var dupErr *DuplicateError
		assert.True(t, errors.As(err, &dupErr))
	})

	t.Run("returns not found for non-existent allowlist", func(t *testing.T) {
		allowlist := &Allowlist{
			ID:          999999,
			Name:        "Test",
			Description: "Test",
		}
		err := repo.UpdateAllowlist(ctx, allowlist)

		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
	})
}

func TestAllowlistRepository_DeleteAllowlist(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewAllowlistRepository(db)
	ctx := context.Background()

	t.Run("deletes allowlist and entries (cascade)", func(t *testing.T) {
		allowlist, err := repo.CreateAllowlist(ctx, "Delete Me", "Test")
		require.NoError(t, err)

		// Add some addresses
		err = repo.AddAddress(ctx, allowlist.ID, "0x1111111111111111111111111111111111111111")
		require.NoError(t, err)
		err = repo.AddAddress(ctx, allowlist.ID, "0x2222222222222222222222222222222222222222")
		require.NoError(t, err)

		// Delete allowlist
		err = repo.DeleteAllowlist(ctx, allowlist.ID)
		require.NoError(t, err)

		// Verify allowlist is deleted
		_, err = repo.GetAllowlist(ctx, allowlist.ID)
		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))

		// Verify entries are deleted (check should return false)
		exists, err := repo.CheckAddress(ctx, allowlist.ID, "0x1111111111111111111111111111111111111111")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("returns not found for non-existent allowlist", func(t *testing.T) {
		err := repo.DeleteAllowlist(ctx, 999999)

		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
	})
}

func TestAllowlistRepository_AddAddress(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewAllowlistRepository(db)
	ctx := context.Background()

	allowlist, err := repo.CreateAllowlist(ctx, "Test List", "Test")
	require.NoError(t, err)

	t.Run("adds valid address", func(t *testing.T) {
		address := "0x742d35cc6634c0532925a3b844bc9e7595f0beb0"
		err := repo.AddAddress(ctx, allowlist.ID, address)
		require.NoError(t, err)

		// Verify address was added
		exists, err := repo.CheckAddress(ctx, allowlist.ID, address)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("normalizes address to lowercase", func(t *testing.T) {
		address := "0xABCDEF0123456789ABCDEF0123456789ABCDEF01"
		err := repo.AddAddress(ctx, allowlist.ID, address)
		require.NoError(t, err)

		// Check with different case
		exists, err := repo.CheckAddress(ctx, allowlist.ID, "0xabcdef0123456789abcdef0123456789abcdef01")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("is idempotent (adding duplicate is OK)", func(t *testing.T) {
		address := "0x1111111111111111111111111111111111111111"
		err := repo.AddAddress(ctx, allowlist.ID, address)
		require.NoError(t, err)

		// Add again
		err = repo.AddAddress(ctx, allowlist.ID, address)
		require.NoError(t, err)
	})

	t.Run("rejects invalid address", func(t *testing.T) {
		err := repo.AddAddress(ctx, allowlist.ID, "invalid")
		require.Error(t, err)
		var addrErr *InvalidAddressError
		assert.True(t, errors.As(err, &addrErr))
	})

	t.Run("returns not found for non-existent allowlist", func(t *testing.T) {
		err := repo.AddAddress(ctx, 999999, "0x1111111111111111111111111111111111111111")
		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
	})
}

func TestAllowlistRepository_RemoveAddress(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewAllowlistRepository(db)
	ctx := context.Background()

	allowlist, err := repo.CreateAllowlist(ctx, "Test List", "Test")
	require.NoError(t, err)

	t.Run("removes existing address", func(t *testing.T) {
		address := "0x1111111111111111111111111111111111111111"
		err := repo.AddAddress(ctx, allowlist.ID, address)
		require.NoError(t, err)

		err = repo.RemoveAddress(ctx, allowlist.ID, address)
		require.NoError(t, err)

		// Verify address was removed
		exists, err := repo.CheckAddress(ctx, allowlist.ID, address)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("handles case-insensitive removal", func(t *testing.T) {
		address := "0x2222222222222222222222222222222222222222"
		err := repo.AddAddress(ctx, allowlist.ID, address)
		require.NoError(t, err)

		// Remove with uppercase
		err = repo.RemoveAddress(ctx, allowlist.ID, "0X2222222222222222222222222222222222222222")
		require.NoError(t, err)

		exists, err := repo.CheckAddress(ctx, allowlist.ID, address)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("returns not found for non-existent address", func(t *testing.T) {
		err := repo.RemoveAddress(ctx, allowlist.ID, "0x9999999999999999999999999999999999999999")
		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
	})

	t.Run("rejects invalid address", func(t *testing.T) {
		err := repo.RemoveAddress(ctx, allowlist.ID, "invalid")
		require.Error(t, err)
		var addrErr *InvalidAddressError
		assert.True(t, errors.As(err, &addrErr))
	})
}

func TestAllowlistRepository_AddAddresses(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewAllowlistRepository(db)
	ctx := context.Background()

	allowlist, err := repo.CreateAllowlist(ctx, "Test List", "Test")
	require.NoError(t, err)

	t.Run("adds multiple addresses", func(t *testing.T) {
		addresses := []string{
			"0x1111111111111111111111111111111111111111",
			"0x2222222222222222222222222222222222222222",
			"0x3333333333333333333333333333333333333333",
		}

		err := repo.AddAddresses(ctx, allowlist.ID, addresses)
		require.NoError(t, err)

		// Verify all addresses were added
		for _, addr := range addresses {
			exists, err := repo.CheckAddress(ctx, allowlist.ID, addr)
			require.NoError(t, err)
			assert.True(t, exists, "address %s should exist", addr)
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		err := repo.AddAddresses(ctx, allowlist.ID, []string{})
		require.NoError(t, err)
	})

	t.Run("normalizes all addresses", func(t *testing.T) {
		addresses := []string{
			"0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
			"0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
		}

		err := repo.AddAddresses(ctx, allowlist.ID, addresses)
		require.NoError(t, err)

		// Check with lowercase
		exists, err := repo.CheckAddress(ctx, allowlist.ID, "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("is idempotent (adding duplicates is OK)", func(t *testing.T) {
		addresses := []string{
			"0x4444444444444444444444444444444444444444",
			"0x4444444444444444444444444444444444444444",
		}

		err := repo.AddAddresses(ctx, allowlist.ID, addresses)
		require.NoError(t, err)

		// Should only have one entry
		allAddrs, err := repo.GetAddresses(ctx, allowlist.ID)
		require.NoError(t, err)

		count := 0
		for _, addr := range allAddrs {
			if addr == "0x4444444444444444444444444444444444444444" {
				count++
			}
		}
		assert.Equal(t, 1, count)
	})

	t.Run("rejects if any address is invalid", func(t *testing.T) {
		addresses := []string{
			"0x5555555555555555555555555555555555555555",
			"invalid",
			"0x6666666666666666666666666666666666666666",
		}

		err := repo.AddAddresses(ctx, allowlist.ID, addresses)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid address")
	})

	t.Run("returns not found for non-existent allowlist", func(t *testing.T) {
		addresses := []string{"0x7777777777777777777777777777777777777777"}
		err := repo.AddAddresses(ctx, 999999, addresses)
		require.Error(t, err)
		var notFoundErr *NotFoundError
		assert.True(t, errors.As(err, &notFoundErr))
	})
}

func TestAllowlistRepository_CheckAddress(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewAllowlistRepository(db)
	ctx := context.Background()

	allowlist, err := repo.CreateAllowlist(ctx, "Test List", "Test")
	require.NoError(t, err)

	t.Run("returns true for existing address", func(t *testing.T) {
		address := "0x1111111111111111111111111111111111111111"
		err := repo.AddAddress(ctx, allowlist.ID, address)
		require.NoError(t, err)

		exists, err := repo.CheckAddress(ctx, allowlist.ID, address)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("returns false for non-existent address", func(t *testing.T) {
		exists, err := repo.CheckAddress(ctx, allowlist.ID, "0x9999999999999999999999999999999999999999")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("is case-insensitive", func(t *testing.T) {
		address := "0xabcdef0123456789abcdef0123456789abcdef01"
		err := repo.AddAddress(ctx, allowlist.ID, address)
		require.NoError(t, err)

		// Check with uppercase
		exists, err := repo.CheckAddress(ctx, allowlist.ID, "0XABCDEF0123456789ABCDEF0123456789ABCDEF01")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("rejects invalid address", func(t *testing.T) {
		_, err := repo.CheckAddress(ctx, allowlist.ID, "invalid")
		require.Error(t, err)
		var addrErr *InvalidAddressError
		assert.True(t, errors.As(err, &addrErr))
	})
}

func TestAllowlistRepository_GetAddresses(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewAllowlistRepository(db)
	ctx := context.Background()

	allowlist, err := repo.CreateAllowlist(ctx, "Test List", "Test")
	require.NoError(t, err)

	t.Run("returns all addresses sorted", func(t *testing.T) {
		addresses := []string{
			"0x3333333333333333333333333333333333333333",
			"0x1111111111111111111111111111111111111111",
			"0x2222222222222222222222222222222222222222",
		}

		err := repo.AddAddresses(ctx, allowlist.ID, addresses)
		require.NoError(t, err)

		result, err := repo.GetAddresses(ctx, allowlist.ID)
		require.NoError(t, err)

		assert.Len(t, result, 3)
		// Should be sorted alphabetically
		assert.Equal(t, "0x1111111111111111111111111111111111111111", result[0])
		assert.Equal(t, "0x2222222222222222222222222222222222222222", result[1])
		assert.Equal(t, "0x3333333333333333333333333333333333333333", result[2])
	})

	t.Run("returns empty list for allowlist with no addresses", func(t *testing.T) {
		emptyList, err := repo.CreateAllowlist(ctx, "Empty List", "No addresses")
		require.NoError(t, err)

		addresses, err := repo.GetAddresses(ctx, emptyList.ID)
		require.NoError(t, err)

		assert.NotNil(t, addresses)
		assert.Len(t, addresses, 0)
	})
}
