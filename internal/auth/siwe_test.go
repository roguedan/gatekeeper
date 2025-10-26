package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED: Test for nonce generation creates unique nonces
func TestSIWEService_GenerateNonce_CreatesUniqueNonce(t *testing.T) {
	service := NewSIWEService(5 * time.Minute)
	ctx := context.Background()

	nonce1, err := service.GenerateNonce(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, nonce1)
	assert.Len(t, nonce1, 32) // 16 bytes hex encoded = 32 chars

	nonce2, err := service.GenerateNonce(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, nonce2)
	assert.NotEqual(t, nonce1, nonce2)
}

// RED: Test for verify nonce with valid nonce
func TestSIWEService_VerifyNonce_WithValidNonce(t *testing.T) {
	service := NewSIWEService(5 * time.Minute)
	ctx := context.Background()

	nonce, _ := service.GenerateNonce(ctx)

	valid, err := service.VerifyNonce(ctx, nonce)
	require.NoError(t, err)
	assert.True(t, valid)
}

// RED: Test for verify nonce with invalid nonce
func TestSIWEService_VerifyNonce_WithInvalidNonce(t *testing.T) {
	service := NewSIWEService(5 * time.Minute)
	ctx := context.Background()

	valid, err := service.VerifyNonce(ctx, "invalid-nonce")
	require.NoError(t, err)
	assert.False(t, valid)
}

// RED: Test for verify nonce with expired nonce
func TestSIWEService_VerifyNonce_WithExpiredNonce(t *testing.T) {
	// Create service with very short TTL
	service := NewSIWEService(1 * time.Millisecond)
	ctx := context.Background()

	nonce, _ := service.GenerateNonce(ctx)

	// Wait for nonce to expire
	time.Sleep(10 * time.Millisecond)

	valid, err := service.VerifyNonce(ctx, nonce)
	require.NoError(t, err)
	assert.False(t, valid)
}

// RED: Test for invalidate nonce
func TestSIWEService_InvalidateNonce_MarkAsUsed(t *testing.T) {
	service := NewSIWEService(5 * time.Minute)
	ctx := context.Background()

	nonce, _ := service.GenerateNonce(ctx)

	// First verify succeeds
	valid, _ := service.VerifyNonce(ctx, nonce)
	assert.True(t, valid)

	// Invalidate the nonce
	err := service.InvalidateNonce(ctx, nonce)
	require.NoError(t, err)

	// Second verify should fail
	valid, _ = service.VerifyNonce(ctx, nonce)
	assert.False(t, valid)
}

// RED: Test for invalidate non-existent nonce returns error
func TestSIWEService_InvalidateNonce_NonExistent(t *testing.T) {
	service := NewSIWEService(5 * time.Minute)
	ctx := context.Background()

	err := service.InvalidateNonce(ctx, "non-existent")
	assert.Error(t, err)
}

// RED: Test for verify nonce after invalidation
func TestSIWEService_VerifyNonce_AfterInvalidation(t *testing.T) {
	service := NewSIWEService(5 * time.Minute)
	ctx := context.Background()

	nonce, _ := service.GenerateNonce(ctx)
	service.InvalidateNonce(ctx, nonce)

	valid, err := service.VerifyNonce(ctx, nonce)
	require.NoError(t, err)
	assert.False(t, valid)
}

// RED: Test for cleanup expired nonces
func TestSIWEService_CleanupExpiredNonces_RemovesExpired(t *testing.T) {
	service := NewSIWEService(1 * time.Millisecond)
	ctx := context.Background()

	// Generate nonce
	nonce, _ := service.GenerateNonce(ctx)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Cleanup should not return error
	err := service.CleanupExpiredNonces(ctx)
	require.NoError(t, err)

	// Nonce should no longer be valid
	valid, _ := service.VerifyNonce(ctx, nonce)
	assert.False(t, valid)
}

// RED: Test for get nonce info
func TestSIWEService_GetNonceInfo_ReturnsInfo(t *testing.T) {
	service := NewSIWEService(5 * time.Minute)
	ctx := context.Background()

	nonce, _ := service.GenerateNonce(ctx)

	info, err := service.GetNonceInfo(ctx, nonce)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, nonce, info.Nonce)
	assert.False(t, info.Used)
	assert.True(t, info.ExpiresAt.After(time.Now()))
}

// RED: Test for get nonce info non-existent nonce
func TestSIWEService_GetNonceInfo_NonExistent(t *testing.T) {
	service := NewSIWEService(5 * time.Minute)
	ctx := context.Background()

	_, err := service.GetNonceInfo(ctx, "non-existent")
	assert.Error(t, err)
}

// RED: Test for high entropy randomness
func TestSIWEService_GenerateNonce_HighEntropy(t *testing.T) {
	service := NewSIWEService(5 * time.Minute)
	ctx := context.Background()

	nonces := make(map[string]bool)
	for i := 0; i < 100; i++ {
		nonce, _ := service.GenerateNonce(ctx)
		assert.False(t, nonces[nonce], "Duplicate nonce generated!")
		nonces[nonce] = true
	}

	assert.Equal(t, 100, len(nonces))
}
