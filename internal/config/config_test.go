package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED: Test for loading all required fields
func TestLoad_AllRequiredFieldsPresent(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/gatekeeper")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")
	t.Setenv("ETHEREUM_RPC", "https://eth.example.com")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "postgres://localhost/gatekeeper", cfg.DatabaseURL)
	assert.Equal(t, []byte("test-secret-key-at-least-32-chars"), cfg.JWTSecret)
	assert.Equal(t, "https://eth.example.com", cfg.EthereumRPC)
}

// RED: Test for missing required field (PORT)
func TestLoad_MissingPort(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gatekeeper")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")
	t.Setenv("ETHEREUM_RPC", "https://eth.example.com")

	_, err := Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PORT")
}

// RED: Test for missing DATABASE_URL
func TestLoad_MissingDatabaseURL(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")
	t.Setenv("ETHEREUM_RPC", "https://eth.example.com")

	_, err := Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL")
}

// RED: Test for missing JWT_SECRET
func TestLoad_MissingJWTSecret(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/gatekeeper")
	t.Setenv("ETHEREUM_RPC", "https://eth.example.com")

	_, err := Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

// RED: Test for missing ETHEREUM_RPC
func TestLoad_MissingEthereumRPC(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/gatekeeper")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")

	_, err := Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ETHEREUM_RPC")
}

// RED: Test for JWT expiry with default value
func TestLoad_JWTExpiryDefaults(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/gatekeeper")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")
	t.Setenv("ETHEREUM_RPC", "https://eth.example.com")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, 24*time.Hour, cfg.JWTExpiry)
}

// RED: Test for JWT expiry custom value
func TestLoad_JWTExpiryCustom(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/gatekeeper")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")
	t.Setenv("ETHEREUM_RPC", "https://eth.example.com")
	t.Setenv("JWT_EXPIRY_HOURS", "48")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, 48*time.Hour, cfg.JWTExpiry)
}

// RED: Test for log level with default value
func TestLoad_LogLevelDefaults(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/gatekeeper")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")
	t.Setenv("ETHEREUM_RPC", "https://eth.example.com")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "info", cfg.LogLevel)
}

// RED: Test for log level custom value
func TestLoad_LogLevelCustom(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/gatekeeper")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")
	t.Setenv("ETHEREUM_RPC", "https://eth.example.com")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "debug", cfg.LogLevel)
}

// RED: Test for nonce TTL with default value
func TestLoad_NonceTTLDefaults(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/gatekeeper")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")
	t.Setenv("ETHEREUM_RPC", "https://eth.example.com")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, 5*time.Minute, cfg.NonceTTL)
}

// RED: Test for nonce TTL custom value
func TestLoad_NonceTTLCustom(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/gatekeeper")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")
	t.Setenv("ETHEREUM_RPC", "https://eth.example.com")
	t.Setenv("NONCE_TTL_MINUTES", "10")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, 10*time.Minute, cfg.NonceTTL)
}

// RED: Test for database pool configuration defaults
func TestLoad_DBPoolConfigDefaults(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/gatekeeper")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")
	t.Setenv("ETHEREUM_RPC", "https://eth.example.com")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, 25, cfg.DBMaxOpenConns)
	assert.Equal(t, 5, cfg.DBMaxIdleConns)
	assert.Equal(t, 5*time.Minute, cfg.DBConnMaxLifetime)
	assert.Equal(t, 1*time.Minute, cfg.DBConnMaxIdleTime)
}

// RED: Test for database pool configuration custom values
func TestLoad_DBPoolConfigCustom(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/gatekeeper")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")
	t.Setenv("ETHEREUM_RPC", "https://eth.example.com")
	t.Setenv("DB_MAX_OPEN_CONNS", "50")
	t.Setenv("DB_MAX_IDLE_CONNS", "10")
	t.Setenv("DB_CONN_MAX_LIFETIME_MINUTES", "10")
	t.Setenv("DB_CONN_MAX_IDLE_TIME_MINUTES", "2")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, 50, cfg.DBMaxOpenConns)
	assert.Equal(t, 10, cfg.DBMaxIdleConns)
	assert.Equal(t, 10*time.Minute, cfg.DBConnMaxLifetime)
	assert.Equal(t, 2*time.Minute, cfg.DBConnMaxIdleTime)
}

// RED: Test for invalid DB_MAX_OPEN_CONNS
func TestLoad_InvalidDBMaxOpenConns(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/gatekeeper")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")
	t.Setenv("ETHEREUM_RPC", "https://eth.example.com")
	t.Setenv("DB_MAX_OPEN_CONNS", "not-a-number")

	_, err := Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "DB_MAX_OPEN_CONNS")
}

// RED: Test for invalid DB_CONN_MAX_LIFETIME_MINUTES
func TestLoad_InvalidDBConnMaxLifetime(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/gatekeeper")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")
	t.Setenv("ETHEREUM_RPC", "https://eth.example.com")
	t.Setenv("DB_CONN_MAX_LIFETIME_MINUTES", "invalid")

	_, err := Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "DB_CONN_MAX_LIFETIME_MINUTES")
}
