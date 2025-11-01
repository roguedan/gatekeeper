package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/yourusername/gatekeeper/internal/audit"
	"github.com/yourusername/gatekeeper/internal/log"
	"github.com/yourusername/gatekeeper/internal/store"
)

// APIKeyHandler handles API key management endpoints
type APIKeyHandler struct {
	apiKeyRepo  *store.APIKeyRepository
	userRepo    *store.UserRepository
	logger      *log.Logger
	auditLogger audit.AuditLogger
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(apiKeyRepo *store.APIKeyRepository, userRepo *store.UserRepository, logger *log.Logger, auditLogger audit.AuditLogger) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyRepo:  apiKeyRepo,
		userRepo:    userRepo,
		logger:      logger,
		auditLogger: auditLogger,
	}
}

// CreateAPIKeyRequest represents the request to create a new API key
type CreateAPIKeyRequest struct {
	Name             string   `json:"name"`
	Scopes           []string `json:"scopes"`
	ExpiresInSeconds *int64   `json:"expiresInSeconds,omitempty"`
}

// CreateAPIKeyResponse represents the response when creating a new API key
type CreateAPIKeyResponse struct {
	Key       string     `json:"key"`     // Raw key - only shown once
	KeyHash   string     `json:"keyHash"` // First 8 chars for reference
	Name      string     `json:"name"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	Message   string     `json:"message"`
}

// APIKeyMetadata represents API key metadata (without the raw key)
type APIKeyMetadata struct {
	ID         int64      `json:"id"`
	KeyHash    string     `json:"keyHash"` // First 8 chars for identification
	Name       string     `json:"name"`
	Scopes     []string   `json:"scopes"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	IsExpired  bool       `json:"isExpired"`
}

// ListAPIKeysResponse represents the response when listing API keys
type ListAPIKeysResponse struct {
	Keys []APIKeyMetadata `json:"keys"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// CreateAPIKey handles POST /api/keys - Generate new API key
func (h *APIKeyHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user from JWT context
	claims := ClaimsFromContext(r)
	if claims == nil {
		h.writeError(w, "Unauthorized", "No authentication claims found", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request body", err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" {
		h.writeError(w, "Validation failed", "Name is required", http.StatusBadRequest)
		return
	}
	if len(req.Name) > 255 {
		h.writeError(w, "Validation failed", "Name must be 255 characters or less", http.StatusBadRequest)
		return
	}
	if len(req.Scopes) == 0 {
		h.writeError(w, "Validation failed", "At least one scope is required", http.StatusBadRequest)
		return
	}
	if req.ExpiresInSeconds != nil && *req.ExpiresInSeconds <= 0 {
		h.writeError(w, "Validation failed", "ExpiresInSeconds must be positive", http.StatusBadRequest)
		return
	}

	// Get or create user by address
	user, err := h.userRepo.GetOrCreateUserByAddress(ctx, claims.Address)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to get/create user for address %s: %v", claims.Address, err))
		h.writeError(w, "Internal server error", "Failed to retrieve user", http.StatusInternalServerError)
		return
	}

	// Convert seconds to duration if provided
	var expiresIn *time.Duration
	if req.ExpiresInSeconds != nil {
		duration := time.Duration(*req.ExpiresInSeconds) * time.Second
		expiresIn = &duration
	}

	// Create repository request
	repoReq := store.APIKeyCreateRequest{
		UserID:    user.ID,
		Name:      req.Name,
		Scopes:    req.Scopes,
		ExpiresIn: expiresIn,
	}

	// Create API key
	rawKey, apiKeyResponse, err := h.apiKeyRepo.CreateAPIKey(ctx, repoReq)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to create API key for user %d: %v", user.ID, err))

		// Audit log: API key creation failed
		if h.auditLogger != nil {
			var expiryStr *string
			if apiKeyResponse != nil && apiKeyResponse.ExpiresAt != nil {
				expStr := apiKeyResponse.ExpiresAt.Format(time.RFC3339)
				expiryStr = &expStr
			}
			h.auditLogger.LogAPIKeyCreated(ctx, audit.AuditEvent{
				Result:    audit.ResultFailure,
				UserAddr:  claims.Address,
				KeyName:   req.Name,
				KeyScopes: req.Scopes,
				KeyExpiry: expiryStr,
				Error:     "failed to create API key",
				ErrorDetail: err.Error(),
			})
		}

		h.writeError(w, "Internal server error", "Failed to create API key", http.StatusInternalServerError)
		return
	}

	// Audit log: API key creation success
	if h.auditLogger != nil {
		var expiryStr *string
		if apiKeyResponse.ExpiresAt != nil {
			expStr := apiKeyResponse.ExpiresAt.Format(time.RFC3339)
			expiryStr = &expStr
		}
		h.auditLogger.LogAPIKeyCreated(ctx, audit.AuditEvent{
			Result:     audit.ResultSuccess,
			UserAddr:   claims.Address,
			KeyID:      apiKeyResponse.ID,
			KeyName:    apiKeyResponse.Name,
			KeyScopes:  apiKeyResponse.Scopes,
			KeyExpiry:  expiryStr,
			ResourceID: fmt.Sprintf("key:%d", apiKeyResponse.ID),
		})
	}

	// Log the creation
	h.logger.Info(fmt.Sprintf("API key created: user=%s, name=%s, id=%d", claims.Address, req.Name, apiKeyResponse.ID))

	// Prepare response
	response := CreateAPIKeyResponse{
		Key:       rawKey,
		KeyHash:   apiKeyResponse.KeyHash[:8], // Show first 8 chars for reference
		Name:      apiKeyResponse.Name,
		Scopes:    apiKeyResponse.Scopes,
		ExpiresAt: apiKeyResponse.ExpiresAt,
		CreatedAt: apiKeyResponse.CreatedAt,
		Message:   "Save this key securely - you won't see it again",
	}

	// Set security headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(response)
}

// ListAPIKeys handles GET /api/keys - List user's API keys
func (h *APIKeyHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user from JWT context
	claims := ClaimsFromContext(r)
	if claims == nil {
		h.writeError(w, "Unauthorized", "No authentication claims found", http.StatusUnauthorized)
		return
	}

	// Get user by address
	user, err := h.userRepo.GetUserByAddress(ctx, claims.Address)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to get user for address %s: %v", claims.Address, err))
		h.writeError(w, "User not found", "User does not exist", http.StatusNotFound)
		return
	}

	// List API keys for user
	keys, err := h.apiKeyRepo.ListAPIKeys(ctx, user.ID)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to list API keys for user %d: %v", user.ID, err))

		// Audit log: API key listing failed
		if h.auditLogger != nil {
			h.auditLogger.LogAPIKeyListed(ctx, audit.AuditEvent{
				Result:      audit.ResultFailure,
				UserAddr:    claims.Address,
				Error:       "failed to list API keys",
				ErrorDetail: err.Error(),
			})
		}

		h.writeError(w, "Internal server error", "Failed to retrieve API keys", http.StatusInternalServerError)
		return
	}

	// Audit log: API key listing success
	if h.auditLogger != nil {
		h.auditLogger.LogAPIKeyListed(ctx, audit.AuditEvent{
			Result:   audit.ResultSuccess,
			UserAddr: claims.Address,
			Metadata: map[string]interface{}{
				"count": len(keys),
			},
		})
	}

	// Convert to metadata format
	metadata := make([]APIKeyMetadata, len(keys))
	now := time.Now()
	for i, key := range keys {
		isExpired := key.ExpiresAt != nil && key.ExpiresAt.Before(now)
		metadata[i] = APIKeyMetadata{
			ID:         key.ID,
			KeyHash:    key.KeyHash[:8], // Show first 8 chars for identification
			Name:       key.Name,
			Scopes:     key.Scopes,
			ExpiresAt:  key.ExpiresAt,
			LastUsedAt: key.LastUsedAt,
			CreatedAt:  key.CreatedAt,
			IsExpired:  isExpired,
		}
	}

	response := ListAPIKeysResponse{
		Keys: metadata,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RevokeAPIKey handles DELETE /api/keys/{id} - Revoke API key
func (h *APIKeyHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user from JWT context
	claims := ClaimsFromContext(r)
	if claims == nil {
		h.writeError(w, "Unauthorized", "No authentication claims found", http.StatusUnauthorized)
		return
	}

	// Extract key ID from URL parameter
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		h.writeError(w, "Invalid request", "Key ID is required", http.StatusBadRequest)
		return
	}

	keyID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.writeError(w, "Invalid request", "Key ID must be a valid integer", http.StatusBadRequest)
		return
	}

	// Get user by address
	user, err := h.userRepo.GetUserByAddress(ctx, claims.Address)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to get user for address %s: %v", claims.Address, err))
		h.writeError(w, "User not found", "User does not exist", http.StatusNotFound)
		return
	}

	// Verify ownership - get the key first
	apiKey, err := h.apiKeyRepo.GetAPIKeyByID(ctx, keyID)
	if err != nil {
		h.writeError(w, "API key not found", "The specified API key does not exist", http.StatusNotFound)
		return
	}

	// Check if the key belongs to the user
	if apiKey.UserID != user.ID {
		h.logger.Warn(fmt.Sprintf("User %s attempted to revoke key %d owned by user %d", claims.Address, keyID, apiKey.UserID))
		h.writeError(w, "Forbidden", "You do not have permission to revoke this API key", http.StatusForbidden)
		return
	}

	// Delete the API key
	err = h.apiKeyRepo.DeleteAPIKey(ctx, keyID)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to delete API key %d: %v", keyID, err))

		// Audit log: API key revocation failed
		if h.auditLogger != nil {
			h.auditLogger.LogAPIKeyRevoked(ctx, audit.AuditEvent{
				Result:      audit.ResultFailure,
				UserAddr:    claims.Address,
				KeyID:       keyID,
				KeyName:     apiKey.Name,
				ResourceID:  fmt.Sprintf("key:%d", keyID),
				Error:       "failed to revoke API key",
				ErrorDetail: err.Error(),
			})
		}

		h.writeError(w, "Internal server error", "Failed to revoke API key", http.StatusInternalServerError)
		return
	}

	// Audit log: API key revocation success
	if h.auditLogger != nil {
		h.auditLogger.LogAPIKeyRevoked(ctx, audit.AuditEvent{
			Result:     audit.ResultSuccess,
			UserAddr:   claims.Address,
			KeyID:      keyID,
			KeyName:    apiKey.Name,
			ResourceID: fmt.Sprintf("key:%d", keyID),
			Metadata: map[string]interface{}{
				"reason": "user_requested",
			},
		})
	}

	// Log the revocation
	h.logger.Info(fmt.Sprintf("API key revoked: user=%s, key_id=%d, name=%s", claims.Address, keyID, apiKey.Name))

	// Return 204 No Content on success
	w.WriteHeader(http.StatusNoContent)
}

// writeError writes a JSON error response
func (h *APIKeyHandler) writeError(w http.ResponseWriter, error, details string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   error,
		Details: details,
	}
	json.NewEncoder(w).Encode(response)
}
