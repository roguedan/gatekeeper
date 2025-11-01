package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/gatekeeper/internal/chain"
	"github.com/yourusername/gatekeeper/internal/store"
	"go.uber.org/zap"
)

// HealthHandler manages health check endpoints
type HealthHandler struct {
	db       *store.DB
	provider *chain.Provider
	logger   *zap.Logger
	startTime time.Time
	version  string
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(db *store.DB, provider *chain.Provider, logger *zap.Logger, version string) *HealthHandler {
	return &HealthHandler{
		db:       db,
		provider: provider,
		logger:   logger,
		startTime: time.Now(),
		version:  version,
	}
}

// HealthStatus represents the overall health status
type HealthStatus string

const (
	StatusOK       HealthStatus = "ok"
	StatusDegraded HealthStatus = "degraded"
	StatusDown     HealthStatus = "down"
)

// HealthResponse represents the detailed health check response
type HealthResponse struct {
	Status    HealthStatus       `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Checks    HealthChecks      `json:"checks"`
}

// HealthChecks contains individual component health checks
type HealthChecks struct {
	Database *ComponentHealth `json:"database"`
	Ethereum *ComponentHealth `json:"ethereum,omitempty"`
	Uptime   int64           `json:"uptime"`
}

// ComponentHealth represents health of a single component
type ComponentHealth struct {
	Status       HealthStatus `json:"status"`
	ResponseTime int64       `json:"responseTime"` // milliseconds
	Message      string      `json:"message"`
	ChainID      string      `json:"chainId,omitempty"`
}

// Health implements comprehensive health checks
// GET /health - Detailed health check with all dependencies
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Prepare response
	response := HealthResponse{
		Status:    StatusOK,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   h.version,
		Checks: HealthChecks{
			Uptime: int64(time.Since(h.startTime).Seconds()),
		},
	}

	// Check database health
	dbHealth := h.checkDatabase(ctx)
	response.Checks.Database = dbHealth

	// Check Ethereum RPC health if provider is configured
	if h.provider != nil {
		ethHealth := h.checkEthereum(ctx)
		response.Checks.Ethereum = ethHealth

		// Determine overall status based on component health
		if ethHealth.Status == StatusDown {
			response.Status = StatusDegraded
		}
	}

	if dbHealth.Status == StatusDown {
		response.Status = StatusDown
	}

	// Set HTTP status code based on health status
	statusCode := http.StatusOK
	if response.Status == StatusDown {
		statusCode = http.StatusServiceUnavailable
	} else if response.Status == StatusDegraded {
		statusCode = http.StatusOK // Still serving traffic, just degraded
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode health response", zap.Error(err))
	}
}

// Live implements Kubernetes liveness probe
// GET /health/live - Simple check: is service running?
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	// Liveness probe only checks if the service process is running
	// No external dependencies checked
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok"}`)
}

// Ready implements Kubernetes readiness probe
// GET /health/ready - Check if ready to serve traffic (all dependencies healthy)
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check database - critical dependency
	dbHealth := h.checkDatabase(ctx)
	if dbHealth.Status == StatusDown {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"not_ready","reason":"database_down"}`)
		return
	}

	// Check Ethereum RPC if configured - critical dependency
	if h.provider != nil {
		ethHealth := h.checkEthereum(ctx)
		if ethHealth.Status == StatusDown {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"not_ready","reason":"ethereum_rpc_down"}`)
			return
		}
	}

	// All checks passed
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ready"}`)
}

// checkDatabase performs database health check
func (h *HealthHandler) checkDatabase(ctx context.Context) *ComponentHealth {
	start := time.Now()

	health := &ComponentHealth{
		Status: StatusOK,
	}

	// Create timeout context for health check
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Perform health check
	var result int
	err := h.db.QueryRowContext(checkCtx, "SELECT 1").Scan(&result)

	health.ResponseTime = time.Since(start).Milliseconds()

	if err != nil {
		health.Status = StatusDown
		health.Message = fmt.Sprintf("Database connection failed: %v", err)
		h.logger.Error("database health check failed",
			zap.Error(err),
			zap.Int64("response_time_ms", health.ResponseTime))
		return health
	}

	// Get database version for additional info
	var version string
	err = h.db.QueryRowContext(checkCtx, "SELECT version()").Scan(&version)
	if err != nil {
		health.Message = "PostgreSQL connected"
	} else {
		// Extract just the PostgreSQL version number
		health.Message = fmt.Sprintf("PostgreSQL connected")
	}

	return health
}

// checkEthereum performs Ethereum RPC health check
func (h *HealthHandler) checkEthereum(ctx context.Context) *ComponentHealth {
	start := time.Now()

	health := &ComponentHealth{
		Status: StatusOK,
	}

	// Create timeout context for health check
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Call eth_chainId to check RPC health
	response, err := h.provider.Call(checkCtx, "eth_chainId", []interface{}{})

	health.ResponseTime = time.Since(start).Milliseconds()

	if err != nil {
		health.Status = StatusDown
		health.Message = fmt.Sprintf("Ethereum RPC failed: %v", err)
		h.logger.Error("ethereum health check failed",
			zap.Error(err),
			zap.Int64("response_time_ms", health.ResponseTime))
		return health
	}

	// Parse the chain ID from response
	var rpcResp struct {
		Result string `json:"result"`
	}

	if err := json.Unmarshal(response, &rpcResp); err == nil && rpcResp.Result != "" {
		health.ChainID = rpcResp.Result
		health.Message = "Ethereum RPC responding"
	} else {
		health.Message = "Ethereum RPC responding"
	}

	return health
}

// GetStats returns current database pool statistics
func (h *HealthHandler) GetStats() store.PoolStats {
	if h.db != nil {
		return h.db.Stats()
	}
	return store.PoolStats{}
}
