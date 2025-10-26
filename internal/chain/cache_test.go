package chain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCache_NewCache creates a new cache
func TestCache_NewCache(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewCache(ttl)

	require.NotNil(t, cache)
	assert.Equal(t, ttl, cache.ttl)
	assert.Equal(t, 0, cache.Size())
}

// TestCache_Set stores a value
func TestCache_Set(t *testing.T) {
	cache := NewCache(5 * time.Minute)

	cache.Set("key1", "value1")

	assert.Equal(t, 1, cache.Size())
}

// TestCache_Get retrieves a stored value
func TestCache_Get(t *testing.T) {
	cache := NewCache(5 * time.Minute)
	cache.Set("key1", "value1")

	value, ok := cache.Get("key1")

	require.True(t, ok)
	assert.Equal(t, "value1", value)
}

// TestCache_Get_NotFound returns false for missing key
func TestCache_Get_NotFound(t *testing.T) {
	cache := NewCache(5 * time.Minute)

	_, ok := cache.Get("missing")

	assert.False(t, ok)
}

// TestCache_Set_Overwrites replaces existing value
func TestCache_Set_Overwrites(t *testing.T) {
	cache := NewCache(5 * time.Minute)
	cache.Set("key1", "value1")

	cache.Set("key1", "value2")

	value, _ := cache.Get("key1")
	assert.Equal(t, "value2", value)
}

// TestCache_Expiration_ItemExpires after TTL
func TestCache_Expiration_ItemExpires(t *testing.T) {
	cache := NewCache(10 * time.Millisecond)
	cache.Set("key1", "value1")

	// Item should exist immediately
	_, ok := cache.Get("key1")
	assert.True(t, ok)

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Item should be expired
	_, ok = cache.Get("key1")
	assert.False(t, ok)
}

// TestCache_MultipleKeys stores multiple values
func TestCache_MultipleKeys(t *testing.T) {
	cache := NewCache(5 * time.Minute)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	assert.Equal(t, 3, cache.Size())

	v1, _ := cache.Get("key1")
	v2, _ := cache.Get("key2")
	v3, _ := cache.Get("key3")

	assert.Equal(t, "value1", v1)
	assert.Equal(t, "value2", v2)
	assert.Equal(t, "value3", v3)
}

// TestCache_Clear removes all items
func TestCache_Clear(t *testing.T) {
	cache := NewCache(5 * time.Minute)
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	assert.Equal(t, 2, cache.Size())

	cache.Clear()

	assert.Equal(t, 0, cache.Size())
}

// TestCache_Delete removes a specific item
func TestCache_Delete(t *testing.T) {
	cache := NewCache(5 * time.Minute)
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	cache.Delete("key1")

	assert.Equal(t, 1, cache.Size())
	_, ok := cache.Get("key1")
	assert.False(t, ok)

	v2, ok := cache.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, "value2", v2)
}

// TestCache_Delete_NonExistent doesn't error
func TestCache_Delete_NonExistent(t *testing.T) {
	cache := NewCache(5 * time.Minute)

	// Should not panic or error
	cache.Delete("missing")

	assert.Equal(t, 0, cache.Size())
}

// TestCache_CacheKey_Format generates correct key format
func TestCache_CacheKey_Format(t *testing.T) {
	key := CacheKey("erc20_balance", "1", "0x1234", "0x5678")

	assert.Contains(t, key, "erc20_balance")
	assert.Contains(t, key, "1")
	assert.Contains(t, key, "0x1234")
	assert.Contains(t, key, "0x5678")
}

// TestCache_CacheKey_Deterministic produces same key for same inputs
func TestCache_CacheKey_Deterministic(t *testing.T) {
	key1 := CacheKey("erc20_balance", "1", "0x1234", "0x5678")
	key2 := CacheKey("erc20_balance", "1", "0x1234", "0x5678")

	assert.Equal(t, key1, key2)
}

// TestCache_CacheKey_Different produces different key for different inputs
func TestCache_CacheKey_Different(t *testing.T) {
	key1 := CacheKey("erc20_balance", "1", "0x1234", "0x5678")
	key2 := CacheKey("erc20_balance", "1", "0x1234", "0x9999")

	assert.NotEqual(t, key1, key2)
}

// TestCache_GetOrSet gets existing or sets new
func TestCache_GetOrSet(t *testing.T) {
	cache := NewCache(5 * time.Minute)

	// First call should set
	value := cache.GetOrSet("key1", func() interface{} {
		return "computed_value"
	})

	assert.Equal(t, "computed_value", value)
	assert.Equal(t, 1, cache.Size())

	// Second call should get
	callCount := 0
	value = cache.GetOrSet("key1", func() interface{} {
		callCount++
		return "new_value"
	})

	assert.Equal(t, "computed_value", value)
	assert.Equal(t, 0, callCount) // Callback should not be called
}

// TestCache_Stats returns cache statistics
func TestCache_Stats(t *testing.T) {
	cache := NewCache(5 * time.Minute)

	// Add some items
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	stats := cache.Stats()

	assert.Equal(t, 2, stats.Size)
	assert.Equal(t, int64(2), stats.Items)
}

// TestCache_Cleanup removes expired items
func TestCache_Cleanup(t *testing.T) {
	cache := NewCache(10 * time.Millisecond)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	assert.Equal(t, 3, cache.Size())

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Manually cleanup
	cache.Cleanup()

	// All items should be removed
	assert.Equal(t, 0, cache.Size())
}

// TestCache_Cleanup_KeepsValid keeps non-expired items
func TestCache_Cleanup_KeepsValid(t *testing.T) {
	cache := NewCache(5 * time.Minute)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Cleanup should not remove non-expired items
	cache.Cleanup()

	assert.Equal(t, 2, cache.Size())
	v1, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", v1)
}

// TestCache_ThreadSafe handles concurrent access
func TestCache_ThreadSafe(t *testing.T) {
	cache := NewCache(5 * time.Minute)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(index int) {
			key := CacheKey("key", string(rune(index)), "", "")
			cache.Set(key, index)
		}(i)
	}

	// Concurrent reads
	time.Sleep(50 * time.Millisecond)
	for i := 0; i < 10; i++ {
		go func(index int) {
			key := CacheKey("key", string(rune(index)), "", "")
			_, _ = cache.Get(key)
		}(i)
	}

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 10, cache.Size())
}

// TestCache_Complex realistic scenario
func TestCache_Complex(t *testing.T) {
	cache := NewCache(100 * time.Millisecond)

	// Add ERC20 balance caches
	cache.Set(CacheKey("erc20_balance", "1", "0xToken", "0xUser1"), int64(1000))
	cache.Set(CacheKey("erc20_balance", "1", "0xToken", "0xUser2"), int64(2000))

	// Add ERC721 ownership caches
	cache.Set(CacheKey("erc721_owner", "1", "0xNFT", "42"), "0xUser1")

	assert.Equal(t, 3, cache.Size())

	// Verify retrieval
	val1, _ := cache.Get(CacheKey("erc20_balance", "1", "0xToken", "0xUser1"))
	assert.Equal(t, int64(1000), val1)

	val2, _ := cache.Get(CacheKey("erc721_owner", "1", "0xNFT", "42"))
	assert.Equal(t, "0xUser1", val2)

	// Wait for expiration
	time.Sleep(110 * time.Millisecond)

	// All should be expired
	_, ok := cache.Get(CacheKey("erc20_balance", "1", "0xToken", "0xUser1"))
	assert.False(t, ok)
}
