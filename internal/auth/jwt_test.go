package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED: Test JWT generation with valid parameters
func TestJWTService_GenerateToken_CreatesValidToken(t *testing.T) {
	secret := "test-secret-key-at-least-32-chars"
	service := NewJWTService([]byte(secret), 24*time.Hour)
	ctx := context.Background()

	address := "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c"
	scopes := []string{"auth", "api"}

	token, err := service.GenerateToken(ctx, address, scopes)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Contains(t, token, ".") // JWT has three parts separated by dots
}

// RED: Test JWT verification with valid token
func TestJWTService_VerifyToken_WithValidToken(t *testing.T) {
	secret := "test-secret-key-at-least-32-chars"
	service := NewJWTService([]byte(secret), 24*time.Hour)
	ctx := context.Background()

	address := "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c"
	scopes := []string{"auth", "api"}
	token, _ := service.GenerateToken(ctx, address, scopes)

	claims, err := service.VerifyToken(ctx, token)

	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, address, claims.Address)
	assert.Equal(t, scopes, claims.Scopes)
}

// RED: Test JWT verification with invalid signature
func TestJWTService_VerifyToken_WithInvalidSignature(t *testing.T) {
	secret := "test-secret-key-at-least-32-chars"
	service := NewJWTService([]byte(secret), 24*time.Hour)
	ctx := context.Background()

	address := "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c"
	scopes := []string{"auth"}
	token, _ := service.GenerateToken(ctx, address, scopes)

	// Tamper with token
	tamperedToken := token[:len(token)-5] + "XXXXX"

	_, err := service.VerifyToken(ctx, tamperedToken)
	assert.Error(t, err)
}

// RED: Test JWT verification with different secret
func TestJWTService_VerifyToken_WithDifferentSecret(t *testing.T) {
	secret1 := "test-secret-key-at-least-32-chars"
	secret2 := "different-secret-key-at-least-32"

	service1 := NewJWTService([]byte(secret1), 24*time.Hour)
	service2 := NewJWTService([]byte(secret2), 24*time.Hour)
	ctx := context.Background()

	address := "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c"
	scopes := []string{"auth"}
	token, _ := service1.GenerateToken(ctx, address, scopes)

	_, err := service2.VerifyToken(ctx, token)
	assert.Error(t, err)
}

// RED: Test JWT with custom expiry
func TestJWTService_VerifyToken_WithCustomExpiry(t *testing.T) {
	secret := "test-secret-key-at-least-32-chars"
	service := NewJWTService([]byte(secret), 1*time.Millisecond)
	ctx := context.Background()

	address := "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c"
	scopes := []string{"auth"}
	token, _ := service.GenerateToken(ctx, address, scopes)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	_, err := service.VerifyToken(ctx, token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

// RED: Test JWT contains correct claims
func TestJWTService_GenerateToken_ContainsCorrectClaims(t *testing.T) {
	secret := "test-secret-key-at-least-32-chars"
	service := NewJWTService([]byte(secret), 24*time.Hour)
	ctx := context.Background()

	address := "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c"
	scopes := []string{"auth", "api", "admin"}
	token, _ := service.GenerateToken(ctx, address, scopes)

	claims, err := service.VerifyToken(ctx, token)

	require.NoError(t, err)
	assert.Equal(t, address, claims.Address)
	assert.Equal(t, scopes, claims.Scopes)
	assert.NotZero(t, claims.IssuedAt)
	assert.NotZero(t, claims.ExpiresAt)
}

// RED: Test JWT round-trip with multiple tokens
func TestJWTService_RoundTrip_MultipleTokens(t *testing.T) {
	secret := "test-secret-key-at-least-32-chars"
	service := NewJWTService([]byte(secret), 24*time.Hour)
	ctx := context.Background()

	addresses := []string{
		"0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c",
		"0x0000000000000000000000000000000000000001",
		"0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	}

	for _, address := range addresses {
		scopes := []string{"auth"}
		token, err := service.GenerateToken(ctx, address, scopes)
		require.NoError(t, err)

		claims, err := service.VerifyToken(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, address, claims.Address)
	}
}

// RED: Test JWT with empty scopes
func TestJWTService_GenerateToken_WithEmptyScopes(t *testing.T) {
	secret := "test-secret-key-at-least-32-chars"
	service := NewJWTService([]byte(secret), 24*time.Hour)
	ctx := context.Background()

	address := "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c"
	scopes := []string{}

	token, err := service.GenerateToken(ctx, address, scopes)
	require.NoError(t, err)

	claims, err := service.VerifyToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, address, claims.Address)
	assert.Empty(t, claims.Scopes)
}

// RED: Test JWT with malformed token
func TestJWTService_VerifyToken_WithMalformedToken(t *testing.T) {
	secret := "test-secret-key-at-least-32-chars"
	service := NewJWTService([]byte(secret), 24*time.Hour)
	ctx := context.Background()

	_, err := service.VerifyToken(ctx, "not.a.valid.jwt.token")
	assert.Error(t, err)
}

// RED: Test JWT with empty token
func TestJWTService_VerifyToken_WithEmptyToken(t *testing.T) {
	secret := "test-secret-key-at-least-32-chars"
	service := NewJWTService([]byte(secret), 24*time.Hour)
	ctx := context.Background()

	_, err := service.VerifyToken(ctx, "")
	assert.Error(t, err)
}
