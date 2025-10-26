package chain

import (
	"fmt"
	"sync"
	"time"
)

// CacheEntry holds a cached value with expiration
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// Cache is a thread-safe in-memory cache with TTL support
type Cache struct {
	data map[string]*CacheEntry
	ttl  time.Duration
	mu   sync.RWMutex
}

// CacheStats contains cache statistics
type CacheStats struct {
	Size  int
	Items int64
}

// NewCache creates a new cache with TTL
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		data: make(map[string]*CacheEntry),
		ttl:  ttl,
	}
}

// Set stores a value in the cache with expiration
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Get retrieves a value from cache, returns false if not found or expired
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Value, true
}

// Delete removes a key from cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

// Clear removes all items from cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*CacheEntry)
}

// Size returns the number of items in cache (including expired)
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.data)
}

// GetOrSet gets a value or sets it if not found
func (c *Cache) GetOrSet(key string, fn func() interface{}) interface{} {
	// Try to get first
	if val, ok := c.Get(key); ok {
		return val
	}

	// Not found, compute and set
	val := fn()
	c.Set(key, val)
	return val
}

// Cleanup removes expired items from cache
func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if now.After(entry.ExpiresAt) {
			delete(c.data, key)
		}
	}
}

// Stats returns cache statistics
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		Size:  len(c.data),
		Items: int64(len(c.data)),
	}
}

// CacheKey generates a cache key from components
// Format: {dataType}:{chainId}:{contract}:{identifier}:{address}
func CacheKey(dataType, chainID, contract, identifier string) string {
	return fmt.Sprintf("%s:%s:%s:%s", dataType, chainID, contract, identifier)
}
