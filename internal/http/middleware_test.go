package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/gatekeeper/internal/auth"
)

// TestJWTMiddleware_WithValidToken verifies middleware accepts valid JWT
func TestJWTMiddleware_WithValidToken(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-chars")
	jwtService := auth.NewJWTService(secret, 24*time.Hour)

	// Generate valid token
	token, _ := jwtService.GenerateToken(context.Background(), "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c", []string{"auth"})

	// Create middleware
	middleware := JWTMiddleware(jwtService)

	// Create test handler that captures context
	var capturedClaims *auth.Claims
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(claimsContextKey).(*auth.Claims)
		capturedClaims = claims
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(testHandler)

	// Create request with valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, capturedClaims)
	assert.Equal(t, "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c", capturedClaims.Address)
}

// TestJWTMiddleware_WithMissingToken returns 401 for missing token
func TestJWTMiddleware_WithMissingToken(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-chars")
	jwtService := auth.NewJWTService(secret, 24*time.Hour)
	middleware := JWTMiddleware(jwtService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(testHandler)

	// Create request without token
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestJWTMiddleware_WithInvalidToken returns 401
func TestJWTMiddleware_WithInvalidToken(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-chars")
	jwtService := auth.NewJWTService(secret, 24*time.Hour)
	middleware := JWTMiddleware(jwtService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(testHandler)

	// Create request with invalid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestJWTMiddleware_WithMalformedHeader returns 401
func TestJWTMiddleware_WithMalformedHeader(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-chars")
	jwtService := auth.NewJWTService(secret, 24*time.Hour)
	middleware := JWTMiddleware(jwtService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(testHandler)

	// Create request with malformed Authorization header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidToken")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestJWTMiddleware_WithExpiredToken returns 401
func TestJWTMiddleware_WithExpiredToken(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-chars")
	jwtService := auth.NewJWTService(secret, 1*time.Millisecond)

	// Generate token that expires quickly
	token, _ := jwtService.GenerateToken(context.Background(), "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c", []string{"auth"})

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Now verify with fresh service that will check expiration
	verifierService := auth.NewJWTService(secret, 24*time.Hour)
	middleware := JWTMiddleware(verifierService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(testHandler)

	// Create request with expired token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestJWTMiddleware_PreservesContext verifies claims are in request context
func TestJWTMiddleware_PreservesContext(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-chars")
	jwtService := auth.NewJWTService(secret, 24*time.Hour)
	address := "0x742d35Cc6634C0532925a3b844Bc390e38f3dF8c"
	scopes := []string{"auth", "api"}

	token, _ := jwtService.GenerateToken(context.Background(), address, scopes)
	middleware := JWTMiddleware(jwtService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(claimsContextKey).(*auth.Claims)
		assert.Equal(t, address, claims.Address)
		assert.Equal(t, scopes, claims.Scopes)
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
