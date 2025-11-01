package http

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yourusername/gatekeeper/internal/audit"
	"github.com/yourusername/gatekeeper/internal/auth"
	"github.com/yourusername/gatekeeper/internal/log"
	"github.com/yourusername/gatekeeper/internal/store"
)

// APIKeyMiddleware creates a middleware that validates API keys
type APIKeyMiddleware struct {
	apiKeyRepo  *store.APIKeyRepository
	userRepo    *store.UserRepository
	logger      *log.Logger
	auditLogger audit.AuditLogger
}

// NewAPIKeyMiddleware creates a new API key middleware
func NewAPIKeyMiddleware(apiKeyRepo *store.APIKeyRepository, userRepo *store.UserRepository, logger *log.Logger, auditLogger audit.AuditLogger) *APIKeyMiddleware {
	return &APIKeyMiddleware{
		apiKeyRepo:  apiKeyRepo,
		userRepo:    userRepo,
		logger:      logger,
		auditLogger: auditLogger,
	}
}

// Middleware returns the HTTP middleware function
func (m *APIKeyMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Check if claims are already in context (from JWT middleware)
			existingClaims := ClaimsFromContext(r)
			if existingClaims != nil {
				// JWT already authenticated, skip API key validation
				next.ServeHTTP(w, r)
				return
			}

			// Try to extract API key from headers
			apiKey := m.extractAPIKey(r)
			if apiKey == "" {
				// No API key found, pass through to next middleware
				// (JWT middleware will handle authentication)
				next.ServeHTTP(w, r)
				return
			}

			// Validate API key format (hex-encoded, 64 characters)
			if !m.isValidAPIKeyFormat(apiKey) {
				m.logger.Warn(fmt.Sprintf("Invalid API key format from %s", r.RemoteAddr))

				// Audit log: Invalid API key format
				if m.auditLogger != nil {
					m.auditLogger.LogAuthAttempt(ctx, audit.AuditEvent{
						Result:   audit.ResultFailure,
						Method:   r.Method,
						Endpoint: r.URL.Path,
						IPAddr:   r.RemoteAddr,
						Error:    "invalid_api_key_format",
					})
				}

				m.writeUnauthorized(w, "invalid_api_key", "API key format is invalid")
				return
			}

			// Validate API key against database
			apiKeyData, err := m.apiKeyRepo.ValidateAPIKey(ctx, apiKey)
			if err != nil {
				m.logger.Warn(fmt.Sprintf("API key validation failed: %v", err))

				// Audit log: API key validation failed
				if m.auditLogger != nil {
					var expiredErr *store.ExpiredError
					var notFoundErr *store.NotFoundError
					errorType := "validation_failed"
					if errors.As(err, &expiredErr) {
						errorType = "key_expired"
					} else if errors.As(err, &notFoundErr) {
						errorType = "key_not_found"
					}

					m.auditLogger.LogAuthAttempt(ctx, audit.AuditEvent{
						Result:      audit.ResultFailure,
						Method:      r.Method,
						Endpoint:    r.URL.Path,
						IPAddr:      r.RemoteAddr,
						Error:       errorType,
						ErrorDetail: err.Error(),
					})
				}

				m.writeUnauthorized(w, "invalid_api_key", "API key not found or expired")
				return
			}

			// Get user information
			user, err := m.userRepo.GetUserByID(ctx, apiKeyData.UserID)
			if err != nil {
				m.logger.Error(fmt.Sprintf("Failed to get user %d for API key %d: %v", apiKeyData.UserID, apiKeyData.ID, err))

				// Audit log: User lookup failed
				if m.auditLogger != nil {
					m.auditLogger.LogAuthAttempt(ctx, audit.AuditEvent{
						Result:      audit.ResultFailure,
						KeyID:       apiKeyData.ID,
						KeyName:     apiKeyData.Name,
						Method:      r.Method,
						Endpoint:    r.URL.Path,
						IPAddr:      r.RemoteAddr,
						Error:       "user_not_found",
						ErrorDetail: err.Error(),
					})
				}

				m.writeUnauthorized(w, "invalid_api_key", "User not found")
				return
			}

			// Create claims from API key data
			claims := &auth.Claims{
				Address: user.Address,
				Scopes:  apiKeyData.Scopes,
			}

			// Inject claims into context
			ctx = context.WithValue(ctx, ClaimsContextKey, claims)
			r = r.WithContext(ctx)

			// Audit log: Successful authentication
			if m.auditLogger != nil {
				m.auditLogger.LogAuthAttempt(ctx, audit.AuditEvent{
					Result:     audit.ResultSuccess,
					UserAddr:   user.Address,
					KeyID:      apiKeyData.ID,
					KeyName:    apiKeyData.Name,
					KeyScopes:  apiKeyData.Scopes,
					Method:     r.Method,
					Endpoint:   r.URL.Path,
					IPAddr:     r.RemoteAddr,
					ResourceID: fmt.Sprintf("key:%d", apiKeyData.ID),
				})
			}

			// Update last_used_at in background (non-blocking)
			go func() {
				// Create a new context for the background operation with timeout
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := m.apiKeyRepo.UpdateLastUsed(ctx, apiKeyData.KeyHash); err != nil {
					// Only log non-context errors to avoid noise from timeout/cancellation
					if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
						m.logger.Error(fmt.Sprintf("Failed to update last_used_at for API key %d: %v", apiKeyData.ID, err))
					}
				}

				// Audit log: API key usage (async)
				if m.auditLogger != nil {
					m.auditLogger.LogAsync(audit.AuditEvent{
						Action:     audit.ActionAPIKeyUsed,
						Result:     audit.ResultSuccess,
						UserAddr:   user.Address,
						KeyID:      apiKeyData.ID,
						KeyName:    apiKeyData.Name,
						Method:     r.Method,
						Endpoint:   r.URL.Path,
						IPAddr:     r.RemoteAddr,
						ResourceID: fmt.Sprintf("key:%d", apiKeyData.ID),
					})
				}
			}()

			// Log successful authentication
			m.logger.Info(fmt.Sprintf("API key authentication successful: user=%s, key_id=%d, key_name=%s",
				user.Address, apiKeyData.ID, apiKeyData.Name))

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// extractAPIKey extracts the API key from the request headers
// Supports both X-API-Key header and Authorization: Bearer format
func (m *APIKeyMiddleware) extractAPIKey(r *http.Request) string {
	// Primary: X-API-Key header
	apiKey := r.Header.Get("X-API-Key")
	if apiKey != "" {
		return strings.TrimSpace(apiKey)
	}

	// Fallback: Authorization: Bearer <key>
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Fields(authHeader)
		if len(parts) == 2 && parts[0] == "Bearer" {
			// Check if it looks like an API key (64 hex chars) vs JWT
			// JWT tokens contain dots, API keys don't
			token := parts[1]
			if !strings.Contains(token, ".") && len(token) == 64 {
				return token
			}
		}
	}

	return ""
}

// isValidAPIKeyFormat validates that the API key is a valid hex string of 64 characters
func (m *APIKeyMiddleware) isValidAPIKeyFormat(apiKey string) bool {
	// Must be exactly 64 characters (32 bytes hex-encoded)
	if len(apiKey) != 64 {
		return false
	}

	// Must be valid hex
	_, err := hex.DecodeString(apiKey)
	return err == nil
}

// writeUnauthorized writes a 401 Unauthorized response
func (m *APIKeyMiddleware) writeUnauthorized(w http.ResponseWriter, errorCode, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	response := ErrorResponse{
		Error:   errorCode,
		Details: details,
	}

	json.NewEncoder(w).Encode(response)
}
