package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/gatekeeper/internal/auth"
	"github.com/yourusername/gatekeeper/internal/log"
	"github.com/yourusername/gatekeeper/internal/store"
)

// MockAPIKeyRepository is a mock implementation of APIKeyRepository
type MockAPIKeyRepository struct {
	mock.Mock
}

func (m *MockAPIKeyRepository) CreateAPIKey(ctx context.Context, req store.APIKeyCreateRequest) (string, *store.APIKeyResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(1) == nil {
		return "", nil, args.Error(2)
	}
	return args.String(0), args.Get(1).(*store.APIKeyResponse), args.Error(2)
}

func (m *MockAPIKeyRepository) ValidateAPIKey(ctx context.Context, rawKey string) (*store.APIKey, error) {
	args := m.Called(ctx, rawKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) ListAPIKeys(ctx context.Context, userID int64) ([]store.APIKey, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) GetAPIKeyByID(ctx context.Context, id int64) (*store.APIKey, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) DeleteAPIKey(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAPIKeyRepository) UpdateLastUsed(ctx context.Context, keyHash string) error {
	args := m.Called(ctx, keyHash)
	return args.Error(0)
}

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetOrCreateUserByAddress(ctx context.Context, address string) (*store.User, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByAddress(ctx context.Context, address string) (*store.User, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id int64) (*store.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, address string) (*store.User, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.User), args.Error(1)
}

func TestCreateAPIKey_Success(t *testing.T) {
	// Setup
	apiKeyRepo := new(MockAPIKeyRepository)
	userRepo := new(MockUserRepository)
	logger, _ := log.New("info")
	handler := NewAPIKeyHandler(apiKeyRepo, userRepo, logger, nil)

	// Mock user
	testUser := &store.User{
		ID:      1,
		Address: "0x1234567890123456789012345678901234567890",
	}

	// Mock API key response
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	testAPIKeyResp := &store.APIKeyResponse{
		ID:        1,
		KeyHash:   "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		Name:      "Test Key",
		Scopes:    []string{"read", "write"},
		ExpiresAt: &expiresAt,
		CreatedAt: now,
	}
	rawKey := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	// Setup expectations
	userRepo.On("GetOrCreateUserByAddress", mock.Anything, testUser.Address).Return(testUser, nil)
	apiKeyRepo.On("CreateAPIKey", mock.Anything, mock.MatchedBy(func(req store.APIKeyCreateRequest) bool {
		return req.UserID == testUser.ID && req.Name == "Test Key" && len(req.Scopes) == 2
	})).Return(rawKey, testAPIKeyResp, nil)

	// Create request
	reqBody := CreateAPIKeyRequest{
		Name:   "Test Key",
		Scopes: []string{"read", "write"},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/keys", bytes.NewReader(bodyBytes))

	// Add claims to context
	claims := &auth.Claims{Address: testUser.Address}
	ctx := context.WithValue(req.Context(), ClaimsContextKey, claims)
	req = req.WithContext(ctx)

	// Execute
	rr := httptest.NewRecorder()
	handler.CreateAPIKey(rr, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Contains(t, rr.Header().Get("Cache-Control"), "no-store")

	var response CreateAPIKeyResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, rawKey, response.Key)
	assert.Equal(t, "abcdef12", response.KeyHash)
	assert.Equal(t, "Test Key", response.Name)
	assert.Contains(t, response.Message, "Save this key securely")

	apiKeyRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestCreateAPIKey_MissingName(t *testing.T) {
	// Setup
	apiKeyRepo := new(MockAPIKeyRepository)
	userRepo := new(MockUserRepository)
	logger, _ := log.New("info")
	handler := NewAPIKeyHandler(apiKeyRepo, userRepo, logger, nil)

	// Create request with missing name
	reqBody := CreateAPIKeyRequest{
		Name:   "",
		Scopes: []string{"read"},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/keys", bytes.NewReader(bodyBytes))

	// Add claims to context
	claims := &auth.Claims{Address: "0x1234567890123456789012345678901234567890"}
	ctx := context.WithValue(req.Context(), ClaimsContextKey, claims)
	req = req.WithContext(ctx)

	// Execute
	rr := httptest.NewRecorder()
	handler.CreateAPIKey(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Contains(t, response.Details, "Name is required")
}

func TestCreateAPIKey_MissingScopes(t *testing.T) {
	// Setup
	apiKeyRepo := new(MockAPIKeyRepository)
	userRepo := new(MockUserRepository)
	logger, _ := log.New("info")
	handler := NewAPIKeyHandler(apiKeyRepo, userRepo, logger, nil)

	// Create request with missing scopes
	reqBody := CreateAPIKeyRequest{
		Name:   "Test Key",
		Scopes: []string{},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/keys", bytes.NewReader(bodyBytes))

	// Add claims to context
	claims := &auth.Claims{Address: "0x1234567890123456789012345678901234567890"}
	ctx := context.WithValue(req.Context(), ClaimsContextKey, claims)
	req = req.WithContext(ctx)

	// Execute
	rr := httptest.NewRecorder()
	handler.CreateAPIKey(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Contains(t, response.Details, "At least one scope is required")
}

func TestCreateAPIKey_Unauthorized(t *testing.T) {
	// Setup
	apiKeyRepo := new(MockAPIKeyRepository)
	userRepo := new(MockUserRepository)
	logger, _ := log.New("info")
	handler := NewAPIKeyHandler(apiKeyRepo, userRepo, logger, nil)

	// Create request without claims in context
	reqBody := CreateAPIKeyRequest{
		Name:   "Test Key",
		Scopes: []string{"read"},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/keys", bytes.NewReader(bodyBytes))

	// Execute
	rr := httptest.NewRecorder()
	handler.CreateAPIKey(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestListAPIKeys_Success(t *testing.T) {
	// Setup
	apiKeyRepo := new(MockAPIKeyRepository)
	userRepo := new(MockUserRepository)
	logger, _ := log.New("info")
	handler := NewAPIKeyHandler(apiKeyRepo, userRepo, logger, nil)

	// Mock user
	testUser := &store.User{
		ID:      1,
		Address: "0x1234567890123456789012345678901234567890",
	}

	// Mock API keys
	now := time.Now()
	lastUsed := now.Add(-1 * time.Hour)
	keys := []store.APIKey{
		{
			ID:         1,
			UserID:     testUser.ID,
			KeyHash:    "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			Name:       "Key 1",
			Scopes:     []string{"read"},
			LastUsedAt: &lastUsed,
			CreatedAt:  now,
		},
		{
			ID:        2,
			UserID:    testUser.ID,
			KeyHash:   "fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
			Name:      "Key 2",
			Scopes:    []string{"read", "write"},
			CreatedAt: now,
		},
	}

	// Setup expectations
	userRepo.On("GetUserByAddress", mock.Anything, testUser.Address).Return(testUser, nil)
	apiKeyRepo.On("ListAPIKeys", mock.Anything, testUser.ID).Return(keys, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/keys", nil)

	// Add claims to context
	claims := &auth.Claims{Address: testUser.Address}
	ctx := context.WithValue(req.Context(), ClaimsContextKey, claims)
	req = req.WithContext(ctx)

	// Execute
	rr := httptest.NewRecorder()
	handler.ListAPIKeys(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response ListAPIKeysResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response.Keys, 2)
	assert.Equal(t, "abcdef12", response.Keys[0].KeyHash)
	assert.Equal(t, "fedcba09", response.Keys[1].KeyHash)

	apiKeyRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestRevokeAPIKey_Success(t *testing.T) {
	// Setup
	apiKeyRepo := new(MockAPIKeyRepository)
	userRepo := new(MockUserRepository)
	logger, _ := log.New("info")
	handler := NewAPIKeyHandler(apiKeyRepo, userRepo, logger, nil)

	// Mock user
	testUser := &store.User{
		ID:      1,
		Address: "0x1234567890123456789012345678901234567890",
	}

	// Mock API key
	testAPIKey := &store.APIKey{
		ID:     1,
		UserID: testUser.ID,
		Name:   "Test Key",
	}

	// Setup expectations
	userRepo.On("GetUserByAddress", mock.Anything, testUser.Address).Return(testUser, nil)
	apiKeyRepo.On("GetAPIKeyByID", mock.Anything, int64(1)).Return(testAPIKey, nil)
	apiKeyRepo.On("DeleteAPIKey", mock.Anything, int64(1)).Return(nil)

	// Create request with URL parameter
	req := httptest.NewRequest("DELETE", "/api/keys/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	// Add claims to context
	claims := &auth.Claims{Address: testUser.Address}
	ctx := context.WithValue(req.Context(), ClaimsContextKey, claims)
	req = req.WithContext(ctx)

	// Execute
	rr := httptest.NewRecorder()
	handler.RevokeAPIKey(rr, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, rr.Code)

	apiKeyRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestRevokeAPIKey_Forbidden(t *testing.T) {
	// Setup
	apiKeyRepo := new(MockAPIKeyRepository)
	userRepo := new(MockUserRepository)
	logger, _ := log.New("info")
	handler := NewAPIKeyHandler(apiKeyRepo, userRepo, logger, nil)

	// Mock user trying to revoke
	testUser := &store.User{
		ID:      1,
		Address: "0x1234567890123456789012345678901234567890",
	}

	// Mock API key owned by different user
	testAPIKey := &store.APIKey{
		ID:     1,
		UserID: 999, // Different user
		Name:   "Test Key",
	}

	// Setup expectations
	userRepo.On("GetUserByAddress", mock.Anything, testUser.Address).Return(testUser, nil)
	apiKeyRepo.On("GetAPIKeyByID", mock.Anything, int64(1)).Return(testAPIKey, nil)

	// Create request with URL parameter
	req := httptest.NewRequest("DELETE", "/api/keys/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	// Add claims to context
	claims := &auth.Claims{Address: testUser.Address}
	ctx := context.WithValue(req.Context(), ClaimsContextKey, claims)
	req = req.WithContext(ctx)

	// Execute
	rr := httptest.NewRecorder()
	handler.RevokeAPIKey(rr, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, rr.Code)

	var response ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Contains(t, response.Error, "Forbidden")

	apiKeyRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}
