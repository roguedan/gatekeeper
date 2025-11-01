package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// DB wraps a sqlx database connection with connection pooling
type DB struct {
	*sqlx.DB
}

// PoolConfig holds database connection pool configuration
type PoolConfig struct {
	MaxOpenConns    int           // Maximum number of open connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum lifetime of a connection
	ConnMaxIdleTime time.Duration // Maximum idle time of a connection
}

// Connect establishes a connection to PostgreSQL with connection pooling
func Connect(ctx context.Context, databaseURL string, poolCfg PoolConfig) (*DB, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(poolCfg.MaxOpenConns)
	db.SetMaxIdleConns(poolCfg.MaxIdleConns)
	db.SetConnMaxLifetime(poolCfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(poolCfg.ConnMaxIdleTime)

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

// Close closes the database connection
func (d *DB) Close() error {
	return d.DB.Close()
}

// Stats returns database connection pool statistics
func (d *DB) Stats() PoolStats {
	stats := d.DB.Stats()
	return PoolStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
	}
}

// PoolStats holds current connection pool statistics
type PoolStats struct {
	MaxOpenConnections int // Maximum number of open connections to the database
	OpenConnections    int // The number of established connections both in use and idle
	InUse              int // The number of connections currently in use
	Idle               int // The number of idle connections
}

// RunMigrations runs all database migrations
func (d *DB) RunMigrations() error {
	return Migrate(context.Background(), d.DB)
}

// CheckDatabaseHealth verifies database connectivity with a simple query.
// Returns an error if the database connection fails or times out after 5 seconds.
func CheckDatabaseHealth(ctx context.Context, db *sql.DB) error {
	// Create a timeout context of 5 seconds
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Execute a simple query to verify database is responding
	var result int
	err := db.QueryRowContext(timeoutCtx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Verify the result is correct
	if result != 1 {
		return fmt.Errorf("database health check returned unexpected value: %d", result)
	}

	return nil
}
