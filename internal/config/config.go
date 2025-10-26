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
	EthereumRPC string

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
