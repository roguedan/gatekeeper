package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the gatekeeper application.
// All fields are validated during Load() and guaranteed to have valid values.
type Config struct {
	// Server configuration
	Port string

	// Database configuration
	DatabaseURL string

	// JWT configuration
	JWTSecret  []byte
	JWTExpiry  time.Duration

	// Ethereum configuration
	EthereumRPC         string        // Primary RPC endpoint
	EthereumRPCFallback string        // Fallback RPC endpoint (optional)
	ChainID             uint64        // Chain ID (1=mainnet, 5=goerli, 11155111=sepolia)
	CacheTTL            time.Duration // Cache time-to-live for blockchain results
	RPCTimeout          time.Duration // RPC call timeout

	// Logging configuration
	LogLevel string

	// SIWE configuration
	NonceTTL time.Duration
}

// Load loads configuration from environment variables.
// Returns error if required variables are missing or invalid.
func Load() (*Config, error) {
	cfg := &Config{}

	// Load required string field
	if err := loadRequiredString("PORT", &cfg.Port); err != nil {
		return nil, err
	}

	if err := loadRequiredString("DATABASE_URL", &cfg.DatabaseURL); err != nil {
		return nil, err
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}
	cfg.JWTSecret = []byte(jwtSecret)

	if err := loadRequiredString("ETHEREUM_RPC", &cfg.EthereumRPC); err != nil {
		return nil, err
	}

	// Load optional blockchain configuration
	cfg.EthereumRPCFallback = os.Getenv("ETHEREUM_RPC_FALLBACK") // Optional fallback

	// Chain ID - default to mainnet (1)
	if err := loadUint64("CHAIN_ID", 1, &cfg.ChainID); err != nil {
		return nil, err
	}

	// Cache TTL - default 300 seconds (5 minutes)
	if err := loadDurationFromSeconds("CACHE_TTL", 300, &cfg.CacheTTL); err != nil {
		return nil, err
	}

	// RPC timeout - default 5 seconds
	if err := loadDurationFromSeconds("RPC_TIMEOUT", 5, &cfg.RPCTimeout); err != nil {
		return nil, err
	}

	// Load optional fields with defaults
	cfg.LogLevel = os.Getenv("LOG_LEVEL")
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	// JWT expiry - default 24 hours
	if err := loadDurationFromHours("JWT_EXPIRY_HOURS", 24, &cfg.JWTExpiry); err != nil {
		return nil, err
	}

	// Nonce TTL - default 5 minutes
	if err := loadDurationFromMinutes("NONCE_TTL_MINUTES", 5, &cfg.NonceTTL); err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadRequiredString loads a required environment variable.
func loadRequiredString(envVar string, dest *string) error {
	*dest = os.Getenv(envVar)
	if *dest == "" {
		return fmt.Errorf("%s environment variable is required", envVar)
	}
	return nil
}

// loadDurationFromHours loads an optional duration from hours environment variable.
func loadDurationFromHours(envVar string, defaultHours int64, dest *time.Duration) error {
	str := os.Getenv(envVar)
	if str == "" {
		*dest = time.Duration(defaultHours) * time.Hour
		return nil
	}

	hours, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return fmt.Errorf("%s must be a valid integer: %w", envVar, err)
	}
	*dest = time.Duration(hours) * time.Hour
	return nil
}

// loadDurationFromMinutes loads an optional duration from minutes environment variable.
func loadDurationFromMinutes(envVar string, defaultMinutes int64, dest *time.Duration) error {
	str := os.Getenv(envVar)
	if str == "" {
		*dest = time.Duration(defaultMinutes) * time.Minute
		return nil
	}

	minutes, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return fmt.Errorf("%s must be a valid integer: %w", envVar, err)
	}
	*dest = time.Duration(minutes) * time.Minute
	return nil
}

// loadDurationFromSeconds loads an optional duration from seconds environment variable.
func loadDurationFromSeconds(envVar string, defaultSeconds int64, dest *time.Duration) error {
	str := os.Getenv(envVar)
	if str == "" {
		*dest = time.Duration(defaultSeconds) * time.Second
		return nil
	}

	seconds, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return fmt.Errorf("%s must be a valid integer: %w", envVar, err)
	}
	*dest = time.Duration(seconds) * time.Second
	return nil
}

// loadUint64 loads an optional uint64 from environment variable.
func loadUint64(envVar string, defaultValue uint64, dest *uint64) error {
	str := os.Getenv(envVar)
	if str == "" {
		*dest = defaultValue
		return nil
	}

	value, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return fmt.Errorf("%s must be a valid uint64: %w", envVar, err)
	}
	*dest = value
	return nil
}
