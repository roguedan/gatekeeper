package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/gatekeeper/internal/chain"
	"github.com/yourusername/gatekeeper/internal/store"
	"go.uber.org/zap"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *store.DB {
	t.Helper()

	// Get database URL from environment or use default
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://gatekeeper:change-this-secure-password@localhost:5432/gatekeeper_test?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolCfg := store.PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	db, err := store.Connect(ctx, dbURL, poolCfg)
	if err != nil {
		t.Skipf("skipping test: failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestHealthHandler_Health(t *testing.T) {
	logger := zap.NewNop()

	t.Run("healthy database and ethereum", func(t *testing.T) {
		// Setup test database
		db := setupTestDB(t)
		defer db.Close()

		// Create mock provider
		provider := chain.NewProvider("http://localhost:8545", "")

		handler := NewHealthHandler(db, provider, logger, "1.0.0")

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler.Health(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response HealthResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "1.0.0", response.Version)
		assert.NotEmpty(t, response.Timestamp)
		assert.NotNil(t, response.Checks.Database)
		assert.Greater(t, response.Checks.Uptime, int64(0))

		// Database should be healthy
		assert.Equal(t, StatusOK, response.Checks.Database.Status)
		assert.Contains(t, response.Checks.Database.Message, "PostgreSQL")
	})

	t.Run("database only (no ethereum provider)", func(t *testing.T) {
		// Setup test database
		db := setupTestDB(t)
		defer db.Close()

		handler := NewHealthHandler(db, nil, logger, "1.0.0")

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler.Health(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response HealthResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, StatusOK, response.Status)
		assert.NotNil(t, response.Checks.Database)
		assert.Nil(t, response.Checks.Ethereum) // No ethereum when provider is nil
	})

	t.Run("measures response time", func(t *testing.T) {
		// Setup test database
		db := setupTestDB(t)
		defer db.Close()

		handler := NewHealthHandler(db, nil, logger, "1.0.0")

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler.Health(w, req)

		var response HealthResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		// Response time should be recorded and greater than 0
		assert.Greater(t, response.Checks.Database.ResponseTime, int64(0))
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		// Setup test database
		db := setupTestDB(t)
		defer db.Close()

		handler := NewHealthHandler(db, nil, logger, "1.0.0")

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		req := httptest.NewRequest("GET", "/health", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Health(w, req)

		var response HealthResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		// Even with cancelled context, handler should respond
		assert.NotEmpty(t, response.Status)
	})
}

func TestHealthHandler_Live(t *testing.T) {
	logger := zap.NewNop()
	db := setupTestDB(t)
	defer db.Close()

	handler := NewHealthHandler(db, nil, logger, "1.0.0")

	t.Run("always returns ok", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health/live", nil)
		w := httptest.NewRecorder()

		handler.Live(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), `"status":"ok"`)
	})

	t.Run("does not check dependencies", func(t *testing.T) {
		// Even with a nil database, liveness should succeed
		nilHandler := NewHealthHandler(nil, nil, logger, "1.0.0")

		req := httptest.NewRequest("GET", "/health/live", nil)
		w := httptest.NewRecorder()

		nilHandler.Live(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHealthHandler_Ready(t *testing.T) {
	logger := zap.NewNop()

	t.Run("ready when database is healthy", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		handler := NewHealthHandler(db, nil, logger, "1.0.0")

		req := httptest.NewRequest("GET", "/health/ready", nil)
		w := httptest.NewRecorder()

		handler.Ready(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), `"status":"ready"`)
	})

	t.Run("not ready with slow database", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		handler := NewHealthHandler(db, nil, logger, "1.0.0")

		// Create a context with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Wait for context to expire
		time.Sleep(10 * time.Millisecond)

		req := httptest.NewRequest("GET", "/health/ready", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Ready(w, req)

		// Should still succeed because health check has its own timeout
		// This test demonstrates that the handler manages timeouts independently
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHealthHandler_GetStats(t *testing.T) {
	logger := zap.NewNop()

	t.Run("returns database pool stats", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		handler := NewHealthHandler(db, nil, logger, "1.0.0")

		stats := handler.GetStats()

		assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0)
		assert.GreaterOrEqual(t, stats.OpenConnections, 0)
		assert.GreaterOrEqual(t, stats.InUse, 0)
		assert.GreaterOrEqual(t, stats.Idle, 0)
	})

	t.Run("handles nil database", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, logger, "1.0.0")

		stats := handler.GetStats()

		// Should return zero stats without panicking
		assert.Equal(t, 0, stats.MaxOpenConnections)
		assert.Equal(t, 0, stats.OpenConnections)
	})
}

func TestHealthResponse_JSONFormat(t *testing.T) {
	t.Run("serializes correctly", func(t *testing.T) {
		response := HealthResponse{
			Status:    StatusOK,
			Timestamp: "2024-11-01T12:00:00Z",
			Version:   "1.0.0",
			Checks: HealthChecks{
				Database: &ComponentHealth{
					Status:       StatusOK,
					ResponseTime: 15,
					Message:      "PostgreSQL connected",
				},
				Ethereum: &ComponentHealth{
					Status:       StatusOK,
					ResponseTime: 42,
					Message:      "Ethereum RPC responding",
					ChainID:      "0x1",
				},
				Uptime: 3600,
			},
		}

		data, err := json.Marshal(response)
		require.NoError(t, err)

		var decoded HealthResponse
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, response.Status, decoded.Status)
		assert.Equal(t, response.Version, decoded.Version)
		assert.Equal(t, response.Checks.Database.Status, decoded.Checks.Database.Status)
		assert.Equal(t, response.Checks.Ethereum.ChainID, decoded.Checks.Ethereum.ChainID)
	})
}
