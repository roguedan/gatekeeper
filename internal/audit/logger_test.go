package audit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// TestAuditLogger_LogAPIKeyCreated tests API key creation logging
func TestAuditLogger_LogAPIKeyCreated(t *testing.T) {
	// Create observer to capture logs
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	auditLogger := NewAuditLogger(logger)

	ctx := context.Background()
	expiryStr := "2024-12-31T23:59:59Z"

	// Test successful API key creation
	auditLogger.LogAPIKeyCreated(ctx, AuditEvent{
		Result:     ResultSuccess,
		UserAddr:   "0x1234567890abcdef",
		KeyID:      123,
		KeyName:    "test-key",
		KeyScopes:  []string{"read", "write"},
		KeyExpiry:  &expiryStr,
		ResourceID: "key:123",
	})

	// Wait a bit for async processing
	time.Sleep(10 * time.Millisecond)

	// Verify log was captured
	entries := observed.All()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "audit event", entry.Message)
	assert.Equal(t, string(ActionAPIKeyCreated), entry.ContextMap()["action"])
	assert.Equal(t, string(ResultSuccess), entry.ContextMap()["result"])
	assert.Equal(t, "0x1234567890abcdef", entry.ContextMap()["user_addr"])
	assert.Equal(t, int64(123), entry.ContextMap()["key_id"])
	assert.Equal(t, "test-key", entry.ContextMap()["key_name"])
	assert.Equal(t, "key:123", entry.ContextMap()["resource_id"])
}

// TestAuditLogger_LogAPIKeyRevoked tests API key revocation logging
func TestAuditLogger_LogAPIKeyRevoked(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	auditLogger := NewAuditLogger(logger)

	ctx := context.Background()

	auditLogger.LogAPIKeyRevoked(ctx, AuditEvent{
		Result:     ResultSuccess,
		UserAddr:   "0xabcdef1234567890",
		KeyID:      456,
		KeyName:    "revoked-key",
		ResourceID: "key:456",
		Metadata: map[string]interface{}{
			"reason": "user_requested",
		},
	})

	time.Sleep(10 * time.Millisecond)

	entries := observed.All()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, string(ActionAPIKeyRevoked), entry.ContextMap()["action"])
	assert.Equal(t, string(ResultSuccess), entry.ContextMap()["result"])
	assert.Equal(t, "0xabcdef1234567890", entry.ContextMap()["user_addr"])
	assert.Equal(t, int64(456), entry.ContextMap()["key_id"])
	assert.Equal(t, "revoked-key", entry.ContextMap()["key_name"])
	assert.Equal(t, "user_requested", entry.ContextMap()["reason"])
}

// TestAuditLogger_LogAuthAttempt tests authentication attempt logging
func TestAuditLogger_LogAuthAttempt(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	auditLogger := NewAuditLogger(logger)

	ctx := context.Background()

	t.Run("successful authentication", func(t *testing.T) {
		observed.TakeAll() // Clear previous logs

		auditLogger.LogAuthAttempt(ctx, AuditEvent{
			Result:     ResultSuccess,
			UserAddr:   "0x1234567890abcdef",
			KeyID:      789,
			KeyName:    "auth-key",
			KeyScopes:  []string{"admin"},
			Method:     "POST",
			Endpoint:   "/api/data",
			IPAddr:     "192.168.1.1",
			ResourceID: "key:789",
		})

		time.Sleep(10 * time.Millisecond)

		entries := observed.All()
		require.Len(t, entries, 1)

		entry := entries[0]
		assert.Equal(t, string(ActionAuthSuccess), entry.ContextMap()["action"])
		assert.Equal(t, string(ResultSuccess), entry.ContextMap()["result"])
		assert.Equal(t, "POST", entry.ContextMap()["method"])
		assert.Equal(t, "/api/data", entry.ContextMap()["endpoint"])
		assert.Equal(t, "192.168.1.1", entry.ContextMap()["ip_addr"])
	})

	t.Run("failed authentication", func(t *testing.T) {
		observed.TakeAll() // Clear previous logs

		auditLogger.LogAuthAttempt(ctx, AuditEvent{
			Result:      ResultFailure,
			Method:      "POST",
			Endpoint:    "/api/data",
			IPAddr:      "192.168.1.100",
			Error:       "invalid_api_key",
			ErrorDetail: "API key not found",
		})

		time.Sleep(10 * time.Millisecond)

		entries := observed.All()
		require.Len(t, entries, 1)

		entry := entries[0]
		assert.Equal(t, string(ActionAuthFailure), entry.ContextMap()["action"])
		assert.Equal(t, string(ResultFailure), entry.ContextMap()["result"])
		assert.Equal(t, "invalid_api_key", entry.ContextMap()["error"])
		assert.Equal(t, "API key not found", entry.ContextMap()["error_detail"])
		assert.Equal(t, zapcore.ErrorLevel, entry.Level) // Error level for failures
	})
}

// TestAuditLogger_LogAuthzDecision tests authorization decision logging
func TestAuditLogger_LogAuthzDecision(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	auditLogger := NewAuditLogger(logger)

	ctx := context.Background()

	t.Run("access granted", func(t *testing.T) {
		observed.TakeAll()

		auditLogger.LogAuthzDecision(ctx, AuditEvent{
			Result:       ResultGranted,
			UserAddr:     "0x1234567890abcdef",
			Method:       "GET",
			Endpoint:     "/api/protected",
			IPAddr:       "192.168.1.50",
			PolicyPath:   "/api/protected",
			PolicyMethod: "GET",
			Metadata: map[string]interface{}{
				"policies_count": 2,
				"scopes":         []string{"read"},
			},
		})

		time.Sleep(10 * time.Millisecond)

		entries := observed.All()
		require.Len(t, entries, 1)

		entry := entries[0]
		assert.Equal(t, string(ActionAuthzGranted), entry.ContextMap()["action"])
		assert.Equal(t, string(ResultGranted), entry.ContextMap()["result"])
		assert.Equal(t, "/api/protected", entry.ContextMap()["policy_path"])
		assert.Equal(t, "GET", entry.ContextMap()["policy_method"])
		assert.Equal(t, int64(2), entry.ContextMap()["policies_count"])
	})

	t.Run("access denied", func(t *testing.T) {
		observed.TakeAll()

		auditLogger.LogAuthzDecision(ctx, AuditEvent{
			Result:       ResultDenied,
			UserAddr:     "0xabcdef1234567890",
			Method:       "POST",
			Endpoint:     "/api/admin",
			IPAddr:       "192.168.1.200",
			PolicyPath:   "/api/admin",
			PolicyMethod: "POST",
			Error:        "policy_failed",
			Metadata: map[string]interface{}{
				"policies_count": 1,
			},
		})

		time.Sleep(20 * time.Millisecond)

		entries := observed.All()
		require.GreaterOrEqual(t, len(entries), 1, "Should have at least one log entry")

		// Find the most recent entry (should be the one we just logged)
		entry := entries[len(entries)-1]
		assert.Equal(t, string(ActionAuthzDenied), entry.ContextMap()["action"])
		assert.Equal(t, string(ResultDenied), entry.ContextMap()["result"])
		assert.Equal(t, "policy_failed", entry.ContextMap()["error"])
		// Events with errors use error level, even if denied
		assert.Equal(t, zapcore.ErrorLevel, entry.Level)
	})
}

// TestAuditLogger_LogPolicyEvaluation tests policy evaluation logging
func TestAuditLogger_LogPolicyEvaluation(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	auditLogger := NewAuditLogger(logger)

	ctx := context.Background()

	auditLogger.LogPolicyEvaluation(ctx, AuditEvent{
		Result:       ResultSuccess,
		UserAddr:     "0x1234567890abcdef",
		PolicyPath:   "/api/data",
		PolicyMethod: "GET",
		RuleType:     "has_scope",
		RuleResult:   true,
		Metadata: map[string]interface{}{
			"rule_name": "require_read_scope",
		},
	})

	time.Sleep(10 * time.Millisecond)

	entries := observed.All()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, string(ActionPolicyEvaluated), entry.ContextMap()["action"])
	assert.Equal(t, "has_scope", entry.ContextMap()["rule_type"])
	assert.Equal(t, true, entry.ContextMap()["rule_result"])
}

// TestAuditLogger_LogAsync tests asynchronous logging
func TestAuditLogger_LogAsync(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	auditLogger := NewAuditLogger(logger)

	// Log multiple events asynchronously
	for i := 0; i < 5; i++ {
		auditLogger.LogAsync(AuditEvent{
			Action:     ActionAPIKeyUsed,
			Result:     ResultSuccess,
			UserAddr:   "0x1234567890abcdef",
			KeyID:      int64(i),
			Method:     "GET",
			Endpoint:   "/api/data",
			ResourceID: "key:1",
		})
	}

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	// Verify all events were logged
	entries := observed.All()
	assert.GreaterOrEqual(t, len(entries), 5)

	// Verify events have correct action
	for _, entry := range entries {
		assert.Equal(t, string(ActionAPIKeyUsed), entry.ContextMap()["action"])
	}
}

// TestAuditLogger_NoSensitiveData tests that sensitive data is not logged
func TestAuditLogger_NoSensitiveData(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	auditLogger := NewAuditLogger(logger)

	ctx := context.Background()

	// API keys should never be logged, only metadata
	auditLogger.LogAPIKeyCreated(ctx, AuditEvent{
		Result:     ResultSuccess,
		UserAddr:   "0x1234567890abcdef",
		KeyID:      123,
		KeyName:    "test-key",
		ResourceID: "key:123",
		// Note: Raw API key is NOT included in audit event
	})

	time.Sleep(10 * time.Millisecond)

	entries := observed.All()
	require.Len(t, entries, 1)

	entry := entries[0]
	// Verify that we only log metadata, not the actual key
	assert.NotContains(t, entry.ContextMap(), "raw_key")
	assert.NotContains(t, entry.ContextMap(), "api_key")
	assert.Contains(t, entry.ContextMap(), "key_id")
	assert.Contains(t, entry.ContextMap(), "key_name")
}

// TestAuditLogger_StructuredFormat tests that logs are properly structured
func TestAuditLogger_StructuredFormat(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	auditLogger := NewAuditLogger(logger)

	ctx := context.Background()

	auditLogger.LogAPIKeyCreated(ctx, AuditEvent{
		Result:     ResultSuccess,
		UserAddr:   "0x1234567890abcdef",
		KeyID:      123,
		KeyName:    "test-key",
		ResourceID: "key:123",
		Metadata: map[string]interface{}{
			"custom_field_1": "value1",
			"custom_field_2": 42,
		},
	})

	time.Sleep(10 * time.Millisecond)

	entries := observed.All()
	require.Len(t, entries, 1)

	entry := entries[0]

	// Verify all required fields are present
	requiredFields := []string{
		"timestamp",
		"action",
		"result",
		"user_addr",
		"key_id",
		"key_name",
		"resource_id",
	}

	for _, field := range requiredFields {
		assert.Contains(t, entry.ContextMap(), field, "Missing required field: %s", field)
	}

	// Verify metadata fields are present
	assert.Equal(t, "value1", entry.ContextMap()["custom_field_1"])
	assert.Equal(t, int64(42), entry.ContextMap()["custom_field_2"])
}

// TestAuditLogger_LogLevels tests that correct log levels are used
func TestAuditLogger_LogLevels(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	auditLogger := NewAuditLogger(logger)

	ctx := context.Background()

	tests := []struct {
		name          string
		event         AuditEvent
		expectedLevel zapcore.Level
	}{
		{
			name: "success events use info level",
			event: AuditEvent{
				Result:   ResultSuccess,
				UserAddr: "0x1234567890abcdef",
			},
			expectedLevel: zapcore.InfoLevel,
		},
		{
			name: "failure events use error level",
			event: AuditEvent{
				Result: ResultFailure,
				Error:  "something failed",
			},
			expectedLevel: zapcore.ErrorLevel,
		},
		{
			name: "denied events use warn level",
			event: AuditEvent{
				Result:   ResultDenied,
				UserAddr: "0x1234567890abcdef",
			},
			expectedLevel: zapcore.WarnLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observed.TakeAll() // Clear previous logs

			auditLogger.Log(ctx, tt.event)

			time.Sleep(10 * time.Millisecond)

			entries := observed.All()
			require.Len(t, entries, 1)

			assert.Equal(t, tt.expectedLevel, entries[0].Level)
		})
	}
}

// TestAuditLogger_BufferOverflow tests behavior when async buffer is full
func TestAuditLogger_BufferOverflow(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	auditLogger := NewAuditLogger(logger)

	// Try to overflow the buffer with many rapid events
	// The implementation should handle this gracefully
	for i := 0; i < 1500; i++ {
		auditLogger.LogAsync(AuditEvent{
			Action:   ActionAPIKeyUsed,
			Result:   ResultSuccess,
			UserAddr: "0x1234567890abcdef",
			KeyID:    int64(i),
		})
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Should have logged most events (some may be dropped if buffer overflows)
	entries := observed.All()
	assert.Greater(t, len(entries), 1000, "Should have logged a significant number of events")
}
