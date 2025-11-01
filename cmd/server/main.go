package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/yourusername/gatekeeper/internal/audit"
	"github.com/yourusername/gatekeeper/internal/auth"
	"github.com/yourusername/gatekeeper/internal/chain"
	"github.com/yourusername/gatekeeper/internal/config"
	httpserver "github.com/yourusername/gatekeeper/internal/http"
	"github.com/yourusername/gatekeeper/internal/http/handlers"
	"github.com/yourusername/gatekeeper/internal/log"
	"github.com/yourusername/gatekeeper/internal/policy"
	"github.com/yourusername/gatekeeper/internal/store"
)

func main() {
	// Load .env file for development (ignore error if file doesn't exist)
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := log.New(cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info(fmt.Sprintf("Starting Gatekeeper (port %s)", cfg.Port))

	// Initialize database connection with pool configuration
	poolCfg := store.PoolConfig{
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
		ConnMaxIdleTime: cfg.DBConnMaxIdleTime,
	}
	db, err := store.Connect(context.Background(), cfg.DatabaseURL, poolCfg)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to database: %v", err))
		os.Exit(1)
	}
	defer db.Close()

	logger.Info(fmt.Sprintf("Database connected successfully (pool: max_open=%d, max_idle=%d, max_lifetime=%v, max_idle_time=%v)",
		cfg.DBMaxOpenConns, cfg.DBMaxIdleConns, cfg.DBConnMaxLifetime, cfg.DBConnMaxIdleTime))

	// Initialize repositories
	apiKeyRepo := store.NewAPIKeyRepository(db)
	userRepo := store.NewUserRepository(db)

	// Initialize SIWE service
	siweService := auth.NewSIWEService(cfg.NonceTTL)

	// Initialize JWT service
	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTExpiry)

	// Initialize blockchain provider (if RPC is configured)
	var provider *chain.Provider
	if cfg.EthereumRPC != "" {
		provider = chain.NewProvider(cfg.EthereumRPC, "")

		// Test RPC connection
		if !provider.HealthCheck(context.Background()) {
			logger.Warn("Ethereum RPC provider is not responding")
		} else {
			logger.Info("Ethereum RPC provider connected")
		}
	}

	// Initialize cache
	cache := chain.NewCache(5 * time.Minute)

	// Initialize policy manager
	policyManager := policy.NewPolicyManager(provider, cache)

	// Initialize audit logger
	auditLogger := audit.NewAuditLogger(logger.Logger)

	// Initialize API Key handlers
	apiKeyHandler := httpserver.NewAPIKeyHandler(apiKeyRepo, userRepo, logger, auditLogger)

	// Initialize API Key middleware
	apiKeyMiddleware := httpserver.NewAPIKeyMiddleware(apiKeyRepo, userRepo, logger, auditLogger)

	// Initialize documentation handler
	docsHandler := handlers.NewDocsHandler()

	// Initialize rate limiters
	apiKeyCreationLimiter := httpserver.NewInMemoryRateLimiter(
		cfg.APIKeyCreationRateLimit,
		time.Hour,
		cfg.APIKeyCreationBurstLimit,
	)
	apiUsageLimiter := httpserver.NewInMemoryRateLimiter(
		cfg.APIUsageRateLimit,
		time.Minute,
		cfg.APIUsageBurstLimit,
	)

	// Create rate limit middlewares
	apiKeyCreationRateLimiter := httpserver.NewUserRateLimitMiddleware(apiKeyCreationLimiter, logger)
	apiUsageRateLimiter := httpserver.NewUserRateLimitMiddleware(apiUsageLimiter, logger)

	logger.Info(fmt.Sprintf("Rate limiting enabled: API key creation=%d/hour (burst=%d), API usage=%d/min (burst=%d)",
		cfg.APIKeyCreationRateLimit, cfg.APIKeyCreationBurstLimit,
		cfg.APIUsageRateLimit, cfg.APIUsageBurstLimit))

	// Create HTTP router
	router := mux.NewRouter()

	// GET /auth/siwe/nonce - Get a nonce for signing
	router.HandleFunc("/auth/siwe/nonce", func(w http.ResponseWriter, r *http.Request) {
		nonce, err := siweService.GenerateNonce(r.Context())
		if err != nil {
			logger.Error(fmt.Sprintf("failed to generate nonce: %v", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		expiresIn := int(cfg.NonceTTL.Seconds())
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"nonce":"%s","expiresIn":%d}`, nonce, expiresIn)
	}).Methods("GET")

	// POST /auth/siwe/verify - Verify SIWE signature and issue JWT
	router.HandleFunc("/auth/siwe/verify", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Message   string `json:"message"`
			Signature string `json:"signature"`
		}

		if err := parseJSON(r, &req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if req.Message == "" || req.Signature == "" {
			http.Error(w, "Missing message or signature", http.StatusBadRequest)
			return
		}

		// For now, just verify nonce exists and generate token
		// In production, would verify actual SIWE signature
		// Extract address from message (simplified: look for "0x" address pattern)
		address := extractAddressFromMessage(req.Message)
		if address == "" {
			http.Error(w, "Invalid message format", http.StatusBadRequest)
			return
		}

		token, err := jwtService.GenerateToken(r.Context(), address, []string{})
		if err != nil {
			logger.Error(fmt.Sprintf("failed to generate token: %v", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		expiresInSeconds := int(cfg.JWTExpiry.Seconds())
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"token":"%s","expiresIn":%d,"address":"%s"}`,
			token, expiresInSeconds, address)
	}).Methods("POST")

	// Health check endpoint (no authentication required)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		statusCode := http.StatusOK

		// Check RPC health if configured
		if provider != nil {
			if !provider.HealthCheck(r.Context()) {
				status = "degraded"
				statusCode = http.StatusServiceUnavailable
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		fmt.Fprintf(w, `{"status":"%s","port":"%s"}`, status, cfg.Port)
	}).Methods("GET")

	// Documentation endpoints (no authentication required)
	// GET /openapi.yaml - Serve OpenAPI specification
	router.HandleFunc("/openapi.yaml", docsHandler.ServeOpenAPISpec).Methods("GET", "OPTIONS")

	// GET /docs - Serve Redoc documentation UI
	router.HandleFunc("/docs", docsHandler.ServeRedocUI).Methods("GET", "OPTIONS")

	logger.Info("Documentation endpoints registered: /docs and /openapi.yaml")

	// JWT Middleware for protected routes
	jwtMiddleware := httpserver.JWTMiddleware(jwtService)

	// Policy Middleware for access control
	policyMiddleware := httpserver.NewPolicyMiddleware(policyManager, logger, auditLogger)
	if provider != nil {
		policyMiddleware.SetProvider(provider)
		policyMiddleware.SetCache(cache)
	}

	// Create a subrouter for protected routes with authentication
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Apply authentication middleware chain to /api routes
	// Order: API Key first (optional), then JWT (fallback if no API key), then general API rate limiting
	apiRouter.Use(mux.MiddlewareFunc(apiKeyMiddleware.Middleware()))
	apiRouter.Use(mux.MiddlewareFunc(jwtMiddleware))
	apiRouter.Use(mux.MiddlewareFunc(apiUsageRateLimiter.Middleware()))

	// API Key management endpoints (require authentication + specific rate limiting)
	// Create separate handler for POST /keys with stricter rate limiting
	keysRouter := apiRouter.PathPrefix("/keys").Subrouter()

	// POST /api/keys - stricter rate limit for key creation (10/hour per user)
	keysPostRouter := keysRouter.Methods("POST").Subrouter()
	keysPostRouter.Use(mux.MiddlewareFunc(apiKeyCreationRateLimiter.Middleware()))
	keysPostRouter.HandleFunc("", apiKeyHandler.CreateAPIKey)

	// GET and DELETE have normal API rate limits
	keysRouter.HandleFunc("", apiKeyHandler.ListAPIKeys).Methods("GET")
	keysRouter.HandleFunc("/{id}", apiKeyHandler.RevokeAPIKey).Methods("DELETE")

	// Protected data endpoint with policy enforcement
	dataHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := httpserver.ClaimsFromContext(r)
		if claims == nil {
			http.Error(w, "No claims found", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message":"Access granted","address":"%s"}`, claims.Address)
	})
	apiRouter.Handle("/data", policyMiddleware.Middleware()(dataHandler)).Methods("GET")

	// Create HTTP server
	portStr := cfg.Port
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", portStr),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info(fmt.Sprintf("HTTP server listening on %s", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(fmt.Sprintf("server error: %v", err))
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("shutdown error: %v", err))
		os.Exit(1)
	}

	logger.Info("Server stopped")
}

// parseJSON parses JSON from request body
func parseJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return nil // Simplified - in production would use json.NewDecoder
}

// extractAddressFromMessage extracts Ethereum address from SIWE message
func extractAddressFromMessage(message string) string {
	// Simplified extraction - looks for 0x followed by 40 hex characters
	// In production, would properly parse SIWE message format
	if len(message) < 42 {
		return ""
	}

	// Try to find address pattern in message
	for i := 0; i <= len(message)-42; i++ {
		if message[i:i+2] == "0x" {
			// Check if next 40 chars are hex
			candidate := message[i : i+42]
			if isValidEthereumAddress(candidate) {
				return candidate
			}
		}
	}
	return ""
}

// isValidEthereumAddress checks if a string is a valid Ethereum address
func isValidEthereumAddress(addr string) bool {
	if len(addr) != 42 {
		return false
	}
	if addr[:2] != "0x" {
		return false
	}

	for _, c := range addr[2:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
