package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/gatekeeper/internal/auth"
)

// Mock SIWEVerifier for testing
type mockSIWEVerifier struct {
	verifyFunc func(message, signature string) (string, error)
}

func (m *mockSIWEVerifier) VerifyMessage(ctx context.Context, message, signature string) (string, error) {
	if m.verifyFunc != nil {
		return m.verifyFunc(message, signature)
	}
	return "", nil
}

// TestGetNonce_ReturnsNonce verifies nonce endpoint returns a nonce
func TestGetNonce_ReturnsNonce(t *testing.T) {
	siweService := auth.NewSIWEService(5 * time.Minute)
	jwtService := auth.NewJWTService([]byte("secret"), 24*time.Hour)
	handler := NewAuthHandler(siweService, jwtService)

	req := httptest.NewRequest("GET", "/auth/siwe/nonce", nil)
	rec := httptest.NewRecorder()

	handler.GetNonce(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response["nonce"])
	assert.Len(t, response["nonce"], 32)
}

// TestGetNonce_ReturnsDifferentNonces verifies each call gets unique nonce
func TestGetNonce_ReturnsDifferentNonces(t *testing.T) {
	siweService := auth.NewSIWEService(5 * time.Minute)
	jwtService := auth.NewJWTService([]byte("secret"), 24*time.Hour)
	handler := NewAuthHandler(siweService, jwtService)

	// Get first nonce
	req1 := httptest.NewRequest("GET", "/auth/siwe/nonce", nil)
	rec1 := httptest.NewRecorder()
	handler.GetNonce(rec1, req1)

	var response1 map[string]string
	json.Unmarshal(rec1.Body.Bytes(), &response1)
	nonce1 := response1["nonce"]

	// Get second nonce
	req2 := httptest.NewRequest("GET", "/auth/siwe/nonce", nil)
	rec2 := httptest.NewRecorder()
	handler.GetNonce(rec2, req2)

	var response2 map[string]string
	json.Unmarshal(rec2.Body.Bytes(), &response2)
	nonce2 := response2["nonce"]

	assert.NotEqual(t, nonce1, nonce2)
}

// TestVerifySIWE_WithValidSignature returns JWT
func TestVerifySIWE_WithValidSignature(t *testing.T) {
	siweService := auth.NewSIWEService(5 * time.Minute)
	jwtService := auth.NewJWTService([]byte("secret"), 24*time.Hour)
	handler := NewAuthHandler(siweService, jwtService)

	// Generate nonce
	nonce, _ := siweService.GenerateNonce(context.Background())
	address := "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c"

	// Create request
	requestBody := map[string]string{
		"nonce":      nonce,
		"message":    "example.com wants you to sign in...",
		"signature":  "0xabcd1234",
		"address":    address,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/auth/siwe/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.VerifySIWE(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response["token"])
	assert.NotEmpty(t, response["address"])
}

// TestVerifySIWE_WithInvalidJSON returns bad request
func TestVerifySIWE_WithInvalidJSON(t *testing.T) {
	siweService := auth.NewSIWEService(5 * time.Minute)
	jwtService := auth.NewJWTService([]byte("secret"), 24*time.Hour)
	handler := NewAuthHandler(siweService, jwtService)

	req := httptest.NewRequest("POST", "/auth/siwe/verify", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.VerifySIWE(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestVerifySIWE_WithMissingNonce returns bad request
func TestVerifySIWE_WithMissingNonce(t *testing.T) {
	siweService := auth.NewSIWEService(5 * time.Minute)
	jwtService := auth.NewJWTService([]byte("secret"), 24*time.Hour)
	handler := NewAuthHandler(siweService, jwtService)

	requestBody := map[string]string{
		"message":   "example.com wants you to sign in...",
		"signature": "0xabcd1234",
		"address":   "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/auth/siwe/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.VerifySIWE(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestVerifySIWE_WithInvalidNonce returns unauthorized
func TestVerifySIWE_WithInvalidNonce(t *testing.T) {
	siweService := auth.NewSIWEService(5 * time.Minute)
	jwtService := auth.NewJWTService([]byte("secret"), 24*time.Hour)
	handler := NewAuthHandler(siweService, jwtService)

	requestBody := map[string]string{
		"nonce":      "invalid-nonce-that-does-not-exist",
		"message":    "example.com wants you to sign in...",
		"signature":  "0xabcd1234",
		"address":    "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/auth/siwe/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.VerifySIWE(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestVerifySIWE_WithExpiredNonce returns unauthorized
func TestVerifySIWE_WithExpiredNonce(t *testing.T) {
	siweService := auth.NewSIWEService(1 * time.Millisecond)
	jwtService := auth.NewJWTService([]byte("secret"), 24*time.Hour)
	handler := NewAuthHandler(siweService, jwtService)

	nonce, _ := siweService.GenerateNonce(context.Background())

	// Wait for nonce to expire
	time.Sleep(10 * time.Millisecond)

	requestBody := map[string]string{
		"nonce":      nonce,
		"message":    "example.com wants you to sign in...",
		"signature":  "0xabcd1234",
		"address":    "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/auth/siwe/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.VerifySIWE(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestVerifySIWE_InvalidatesNonceAfterUse prevents nonce reuse
func TestVerifySIWE_InvalidatesNonceAfterUse(t *testing.T) {
	siweService := auth.NewSIWEService(5 * time.Minute)
	jwtService := auth.NewJWTService([]byte("secret"), 24*time.Hour)
	handler := NewAuthHandler(siweService, jwtService)

	nonce, _ := siweService.GenerateNonce(context.Background())
	address := "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c"

	// First verification succeeds
	requestBody := map[string]string{
		"nonce":      nonce,
		"message":    "example.com wants you to sign in...",
		"signature":  "0xabcd1234",
		"address":    address,
	}
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/auth/siwe/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.VerifySIWE(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Second verification with same nonce should fail
	body2, _ := json.Marshal(requestBody)
	req2 := httptest.NewRequest("POST", "/auth/siwe/verify", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	handler.VerifySIWE(rec2, req2)
	assert.Equal(t, http.StatusUnauthorized, rec2.Code)
}

// TestVerifySIWE_ResponseContainsToken verifies JWT is in response
func TestVerifySIWE_ResponseContainsToken(t *testing.T) {
	siweService := auth.NewSIWEService(5 * time.Minute)
	jwtService := auth.NewJWTService([]byte("secret"), 24*time.Hour)
	handler := NewAuthHandler(siweService, jwtService)

	nonce, _ := siweService.GenerateNonce(context.Background())
	address := "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c"

	requestBody := map[string]string{
		"nonce":      nonce,
		"message":    "example.com wants you to sign in...",
		"signature":  "0xabcd1234",
		"address":    address,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/auth/siwe/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.VerifySIWE(rec, req)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)

	// Token should be a valid JWT
	assert.NotEmpty(t, response["token"])
	parts := bytes.Split([]byte(response["token"]), []byte("."))
	assert.Equal(t, 3, len(parts), "Token should have 3 parts (header.payload.signature)")
}

// TestVerifySIWE_ResponseContainsAddress verifies address is in response
func TestVerifySIWE_ResponseContainsAddress(t *testing.T) {
	siweService := auth.NewSIWEService(5 * time.Minute)
	jwtService := auth.NewJWTService([]byte("secret"), 24*time.Hour)
	handler := NewAuthHandler(siweService, jwtService)

	nonce, _ := siweService.GenerateNonce(context.Background())
	address := "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c"

	requestBody := map[string]string{
		"nonce":      nonce,
		"message":    "example.com wants you to sign in...",
		"signature":  "0xabcd1234",
		"address":    address,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/auth/siwe/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.VerifySIWE(rec, req)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)

	assert.Equal(t, address, response["address"])
}
