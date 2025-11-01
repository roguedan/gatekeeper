package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// setupTestDB creates a test database connection and runs migrations
func setupTestDB(t *testing.T) *DB {
	t.Helper()

	// Get database URL from environment or use default
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/gatekeeper_test?sslmode=disable"
	}

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use default pool configuration for tests
	poolCfg := PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	db, err := Connect(ctx, dbURL, poolCfg)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Clean up all tables before each test
	cleanupTestDB(t, db)

	// Register cleanup function
	t.Cleanup(func() {
		cleanupTestDB(t, db)
		db.Close()
	})

	return db
}

// cleanupTestDB removes all data from test tables
func cleanupTestDB(t *testing.T, db *DB) {
	t.Helper()

	// Truncate tables in reverse dependency order
	tables := []string{
		"allowlist_entries",
		"allowlists",
		"api_keys",
		"nonces",
		"users",
	}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		_, err := db.Exec(query)
		if err != nil {
			// Don't fail if table doesn't exist yet
			t.Logf("warning: failed to truncate table %s: %v", table, err)
		}
	}
}

// createTestUser is a helper to create a test user
func createTestUser(t *testing.T, db *DB, address string) *User {
	t.Helper()

	repo := NewUserRepository(db)
	user, err := repo.CreateUser(context.Background(), address)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return user
}

// createTestAllowlist is a helper to create a test allowlist
func createTestAllowlist(t *testing.T, db *DB, name, description string) *Allowlist {
	t.Helper()

	repo := NewAllowlistRepository(db)
	allowlist, err := repo.CreateAllowlist(context.Background(), name, description)
	if err != nil {
		t.Fatalf("failed to create test allowlist: %v", err)
	}

	return allowlist
}

// addTestAddresses is a helper to add multiple addresses to an allowlist
func addTestAddresses(t *testing.T, db *DB, allowlistID int64, addresses []string) {
	t.Helper()

	repo := NewAllowlistRepository(db)
	err := repo.AddAddresses(context.Background(), allowlistID, addresses)
	if err != nil {
		t.Fatalf("failed to add test addresses: %v", err)
	}
}

// createTestAPIKey is a helper to create a test API key
func createTestAPIKey(t *testing.T, db *DB, userID int64, name string, scopes []string) (string, *APIKeyResponse) {
	t.Helper()

	repo := NewAPIKeyRepository(db)
	req := APIKeyCreateRequest{
		UserID: userID,
		Name:   name,
		Scopes: scopes,
	}

	rawKey, response, err := repo.CreateAPIKey(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to create test API key: %v", err)
	}

	return rawKey, response
}

// withTransaction executes a function within a transaction and rolls it back
func withTransaction(t *testing.T, db *DB, fn func(*sqlx.Tx) error) {
	t.Helper()

	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			t.Logf("warning: failed to rollback transaction: %v", err)
		}
	}()

	if err := fn(tx); err != nil {
		t.Fatalf("transaction function failed: %v", err)
	}
}
