package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/gatekeeper/internal/auth"
	"github.com/yourusername/gatekeeper/internal/log"
	"github.com/yourusername/gatekeeper/internal/store"
)

func TestAPIKeyMiddleware_ValidAPIKey_XAPIKeyHeader(t *testing.T) {
	// Setup
	apiKeyRepo := new(MockAPIKeyRepository)
	userRepo := new(MockUserRepository)
	logger, _ := log.New("info")
	middleware := NewAPIKeyMiddleware(apiKeyRepo, userRepo, logger, nil)

	// Mock data
	rawKey := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	testUser := &store.User{
		ID:      1,
		Address: "0x1234567890123456789012345678901234567890",
	}
	testAPIKey := &store.APIKey{
		ID:      1,
		UserID:  testUser.ID,
		KeyHash: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		Name:    "Test Key",
		Scopes:  []string{"read", "write"},
	}

	// Setup expectations
	apiKeyRepo.On("ValidateAPIKey", mock.Anything, rawKey).Return(testAPIKey, nil)
	userRepo.On("GetUserByID", mock.Anything, testUser.ID).Return(testUser, nil)
	apiKeyRepo.On("UpdateLastUsed", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	// Create test handler
	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true

		// Verify claims are in context
		claims := ClaimsFromContext(r)
		assert.NotNil(t, claims)
		assert.Equal(t, testUser.Address, claims.Address)
		assert.Equal(t, testAPIKey.Scopes, claims.Scopes)

		w.WriteHeader(http.StatusOK)
	})

	// Create request with X-API-Key header
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("X-API-Key", rawKey)

	// Execute
	rr := httptest.NewRecorder()
	handler := middleware.Middleware()(nextHandler)
	handler.ServeHTTP(rr, req)

	// Assert
	assert.True(t, nextCalled, "Next handler should be called")
	assert.Equal(t, http.StatusOK, rr.Code)

	// Give time for background goroutine to complete
	time.Sleep(10 * time.Millisecond)

	apiKeyRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestAPIKeyMiddleware_ValidAPIKey_BearerHeader(t *testing.T) {
	// Setup
	apiKeyRepo := new(MockAPIKeyRepository)
	userRepo := new(MockUserRepository)
	logger, _ := log.New("info")
	middleware := NewAPIKeyMiddleware(apiKeyRepo, userRepo, logger, nil)

	// Mock data
	rawKey := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	testUser := &store.User{
		ID:      1,
		Address: "0x1234567890123456789012345678901234567890",
	}
	testAPIKey := &store.APIKey{
		ID:      1,
		UserID:  testUser.ID,
		KeyHash: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		Name:    "Test Key",
		Scopes:  []string{"read"},
	}

	// Setup expectations
	apiKeyRepo.On("ValidateAPIKey", mock.Anything, rawKey).Return(testAPIKey, nil)
	userRepo.On("GetUserByID", mock.Anything, testUser.ID).Return(testUser, nil)
	apiKeyRepo.On("UpdateLastUsed", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	// Create test handler
	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create request with Authorization: Bearer header
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", rawKey))

	// Execute
	rr := httptest.NewRecorder()
	handler := middleware.Middleware()(nextHandler)
	handler.ServeHTTP(rr, req)

	// Assert
	assert.True(t, nextCalled, "Next handler should be called")
	assert.Equal(t, http.StatusOK, rr.Code)

	// Give time for background goroutine to complete
	time.Sleep(10 * time.Millisecond)

	apiKeyRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestAPIKeyMiddleware_InvalidKeyFormat(t *testing.T) {
	// Setup
	apiKeyRepo := new(MockAPIKeyRepository)
	userRepo := new(MockUserRepository)
	logger, _ := log.New("info")
	middleware := NewAPIKeyMiddleware(apiKeyRepo, userRepo, logger, nil)

	// Create test handler
	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create request with invalid key format (too short)
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("X-API-Key", "invalidkey")

	// Execute
	rr := httptest.NewRecorder()
	handler := middleware.Middleware()(nextHandler)
	handler.ServeHTTP(rr, req)

	// Assert
	assert.False(t, nextCalled, "Next handler should not be called")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	// No expectations on repos since we fail before DB lookup
}

func TestAPIKeyMiddleware_InvalidKeyFormat_NotHex(t *testing.T) {
	// Setup
	apiKeyRepo := new(MockAPIKeyRepository)
	userRepo := new(MockUserRepository)
	logger, _ := log.New("info")
	middleware := NewAPIKeyMiddleware(apiKeyRepo, userRepo, logger, nil)

	// Create test handler
	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create request with invalid key format (not hex)
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("X-API-Key", "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")

	// Execute
	rr := httptest.NewRecorder()
	handler := middleware.Middleware()(nextHandler)
	handler.ServeHTTP(rr, req)

	// Assert
	assert.False(t, nextCalled, "Next handler should not be called")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAPIKeyMiddleware_ExpiredKey(t *testing.T) {
	// Setup
	apiKeyRepo := new(MockAPIKeyRepository)
	userRepo := new(MockUserRepository)
	logger, _ := log.New("info")
	middleware := NewAPIKeyMiddleware(apiKeyRepo, userRepo, logger, nil)

	// Mock data
	rawKey := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	// Setup expectations - ValidateAPIKey returns error for expired key
	apiKeyRepo.On("ValidateAPIKey", mock.Anything, rawKey).Return(nil, fmt.Errorf("API key has expired"))

	// Create test handler
	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create request
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("X-API-Key", rawKey)

	// Execute
	rr := httptest.NewRecorder()
	handler := middleware.Middleware()(nextHandler)
	handler.ServeHTTP(rr, req)

	// Assert
	assert.False(t, nextCalled, "Next handler should not be called")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	apiKeyRepo.AssertExpectations(t)
}

func TestAPIKeyMiddleware_NonExistentKey(t *testing.T) {
	// Setup
	apiKeyRepo := new(MockAPIKeyRepository)
	userRepo := new(MockUserRepository)
	logger, _ := log.New("info")
	middleware := NewAPIKeyMiddleware(apiKeyRepo, userRepo, logger, nil)

	// Mock data
	rawKey := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	// Setup expectations - ValidateAPIKey returns error for non-existent key
	apiKeyRepo.On("ValidateAPIKey", mock.Anything, rawKey).Return(nil, fmt.Errorf("API key not found"))

	// Create test handler
	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create request
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("X-API-Key", rawKey)

	// Execute
	rr := httptest.NewRecorder()
	handler := middleware.Middleware()(nextHandler)
	handler.ServeHTTP(rr, req)

	// Assert
	assert.False(t, nextCalled, "Next handler should not be called")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	apiKeyRepo.AssertExpectations(t)
}

func TestAPIKeyMiddleware_NoAPIKey_PassThrough(t *testing.T) {
	// Setup
	apiKeyRepo := new(MockAPIKeyRepository)
	userRepo := new(MockUserRepository)
	logger, _ := log.New("info")
	middleware := NewAPIKeyMiddleware(apiKeyRepo, userRepo, logger, nil)

	// Create test handler
	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true

		// Verify no claims in context
		claims := ClaimsFromContext(r)
		assert.Nil(t, claims)

		w.WriteHeader(http.StatusOK)
	})

	// Create request with no API key
	req := httptest.NewRequest("GET", "/api/data", nil)

	// Execute
	rr := httptest.NewRecorder()
	handler := middleware.Middleware()(nextHandler)
	handler.ServeHTTP(rr, req)

	// Assert
	assert.True(t, nextCalled, "Next handler should be called")
	assert.Equal(t, http.StatusOK, rr.Code)

	// No expectations on repos since we pass through
}

func TestAPIKeyMiddleware_ExistingJWTClaims_SkipValidation(t *testing.T) {
	// Setup
	apiKeyRepo := new(MockAPIKeyRepository)
	userRepo := new(MockUserRepository)
	logger, _ := log.New("info")
	middleware := NewAPIKeyMiddleware(apiKeyRepo, userRepo, logger, nil)

	// Create test handler
	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true

		// Verify existing claims are preserved
		claims := ClaimsFromContext(r)
		assert.NotNil(t, claims)
		assert.Equal(t, "0xjwt", claims.Address)

		w.WriteHeader(http.StatusOK)
	})

	// Create request with API key but existing JWT claims
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("X-API-Key", "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2")

	// Add existing JWT claims to context
	claims := &auth.Claims{Address: "0xjwt", Scopes: []string{"jwt"}}
	ctx := context.WithValue(req.Context(), ClaimsContextKey, claims)
	req = req.WithContext(ctx)

	// Execute
	rr := httptest.NewRecorder()
	handler := middleware.Middleware()(nextHandler)
	handler.ServeHTTP(rr, req)

	// Assert
	assert.True(t, nextCalled, "Next handler should be called")
	assert.Equal(t, http.StatusOK, rr.Code)

	// No expectations on repos since we skip validation when JWT claims exist
}

func TestAPIKeyMiddleware_BearerToken_JWT_IgnoredCorrectly(t *testing.T) {
	// Setup
	apiKeyRepo := new(MockAPIKeyRepository)
	userRepo := new(MockUserRepository)
	logger, _ := log.New("info")
	middleware := NewAPIKeyMiddleware(apiKeyRepo, userRepo, logger, nil)

	// Create test handler
	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create request with JWT token (contains dots, not an API key)
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U")

	// Execute
	rr := httptest.NewRecorder()
	handler := middleware.Middleware()(nextHandler)
	handler.ServeHTTP(rr, req)

	// Assert
	assert.True(t, nextCalled, "Next handler should be called")
	assert.Equal(t, http.StatusOK, rr.Code)

	// No expectations on repos since JWT tokens are ignored by API key middleware
}
