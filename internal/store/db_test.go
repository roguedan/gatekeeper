package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED: Test connection pool initialization with custom configuration
func TestConnect_CustomPoolConfig(t *testing.T) {
	dbURL := getTestDatabaseURL()
	ctx := context.Background()

	poolCfg := PoolConfig{
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: 10 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
	}

	db, err := Connect(ctx, dbURL, poolCfg)
	require.NoError(t, err)
	defer db.Close()

	// Verify connection is working
	err = db.PingContext(ctx)
	require.NoError(t, err)

	// Verify pool stats
	stats := db.Stats()
	assert.Equal(t, 50, stats.MaxOpenConnections)
}

// RED: Test connection pool initialization with default configuration
func TestConnect_DefaultPoolConfig(t *testing.T) {
	dbURL := getTestDatabaseURL()
	ctx := context.Background()

	poolCfg := PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	db, err := Connect(ctx, dbURL, poolCfg)
	require.NoError(t, err)
	defer db.Close()

	// Verify connection is working
	err = db.PingContext(ctx)
	require.NoError(t, err)

	// Verify pool stats
	stats := db.Stats()
	assert.Equal(t, 25, stats.MaxOpenConnections)
}

// RED: Test Stats returns current pool metrics
func TestDB_Stats(t *testing.T) {
	dbURL := getTestDatabaseURL()
	ctx := context.Background()

	poolCfg := PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	db, err := Connect(ctx, dbURL, poolCfg)
	require.NoError(t, err)
	defer db.Close()

	// Get initial stats
	stats := db.Stats()
	assert.Equal(t, 25, stats.MaxOpenConnections)
	assert.GreaterOrEqual(t, stats.OpenConnections, 0)
	assert.GreaterOrEqual(t, stats.Idle, 0)
	assert.GreaterOrEqual(t, stats.InUse, 0)

	// Make a query to open a connection
	err = db.PingContext(ctx)
	require.NoError(t, err)

	// Get stats after query
	stats = db.Stats()
	assert.GreaterOrEqual(t, stats.OpenConnections, 1)
}

// RED: Test connection pool with invalid database URL
func TestConnect_InvalidDatabaseURL(t *testing.T) {
	ctx := context.Background()

	poolCfg := PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	_, err := Connect(ctx, "postgres://invalid:5432/nonexistent", poolCfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

// RED: Test connection pool with zero max connections
func TestConnect_ZeroMaxConnections(t *testing.T) {
	dbURL := getTestDatabaseURL()
	ctx := context.Background()

	poolCfg := PoolConfig{
		MaxOpenConns:    0, // 0 means unlimited in database/sql
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	db, err := Connect(ctx, dbURL, poolCfg)
	require.NoError(t, err)
	defer db.Close()

	// Verify connection is working
	err = db.PingContext(ctx)
	require.NoError(t, err)

	// Stats should show unlimited (0) max connections
	stats := db.Stats()
	assert.Equal(t, 0, stats.MaxOpenConnections)
}

// RED: Test connection pool with context timeout
func TestConnect_ContextTimeout(t *testing.T) {
	dbURL := getTestDatabaseURL()

	// Create a context that's already cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond) // Ensure context is expired

	poolCfg := PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	_, err := Connect(ctx, dbURL, poolCfg)
	require.Error(t, err)
}

// getTestDatabaseURL returns the test database URL from environment or default
func getTestDatabaseURL() string {
	dbURL := "postgres://postgres:postgres@localhost:5432/gatekeeper_test?sslmode=disable"
	return dbURL
}
