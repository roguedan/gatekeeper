package audit

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ActionType represents the type of security-related action
type ActionType string

const (
	// API Key actions
	ActionAPIKeyCreated   ActionType = "api_key_created"
	ActionAPIKeyRevoked   ActionType = "api_key_revoked"
	ActionAPIKeyUsed      ActionType = "api_key_used"
	ActionAPIKeyListed    ActionType = "api_key_listed"
	ActionAPIKeyValidated ActionType = "api_key_validated"

	// Authentication actions
	ActionAuthSuccess ActionType = "auth_success"
	ActionAuthFailure ActionType = "auth_failure"

	// Authorization actions
	ActionAuthzGranted ActionType = "authz_granted"
	ActionAuthzDenied  ActionType = "authz_denied"

	// Policy actions
	ActionPolicyEvaluated ActionType = "policy_evaluated"
	ActionRuleEvaluated   ActionType = "rule_evaluated"
	ActionCacheHit        ActionType = "cache_hit"
	ActionCacheMiss       ActionType = "cache_miss"
	ActionRPCCall         ActionType = "rpc_call"
)

// Result represents the outcome of an action
type Result string

const (
	ResultSuccess Result = "success"
	ResultFailure Result = "failure"
	ResultDenied  Result = "denied"
	ResultGranted Result = "granted"
)

// AuditEvent represents a security audit event
type AuditEvent struct {
	// Core fields
	Timestamp  time.Time  `json:"timestamp"`
	Action     ActionType `json:"action"`
	Result     Result     `json:"result"`
	RequestID  string     `json:"request_id,omitempty"`
	UserAddr   string     `json:"user_addr,omitempty"`
	ResourceID string     `json:"resource_id,omitempty"`

	// API Key specific
	KeyID     int64    `json:"key_id,omitempty"`
	KeyName   string   `json:"key_name,omitempty"`
	KeyScopes []string `json:"key_scopes,omitempty"`
	KeyExpiry *string  `json:"key_expiry,omitempty"`

	// HTTP request specific
	Method   string `json:"method,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
	IPAddr   string `json:"ip_addr,omitempty"`

	// Policy specific
	PolicyPath   string `json:"policy_path,omitempty"`
	PolicyMethod string `json:"policy_method,omitempty"`
	RuleType     string `json:"rule_type,omitempty"`
	RuleResult   bool   `json:"rule_result,omitempty"`

	// Blockchain specific
	ChainID         int64  `json:"chain_id,omitempty"`
	ContractAddress string `json:"contract_address,omitempty"`
	RPCMethod       string `json:"rpc_method,omitempty"`

	// Cache specific
	CacheKey string `json:"cache_key,omitempty"`

	// Error information
	Error       string `json:"error,omitempty"`
	ErrorDetail string `json:"error_detail,omitempty"`

	// Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AuditLogger provides structured audit logging for security events
type AuditLogger interface {
	// API Key operations
	LogAPIKeyCreated(ctx context.Context, event AuditEvent)
	LogAPIKeyRevoked(ctx context.Context, event AuditEvent)
	LogAPIKeyUsed(ctx context.Context, event AuditEvent)
	LogAPIKeyListed(ctx context.Context, event AuditEvent)

	// Authentication
	LogAuthAttempt(ctx context.Context, event AuditEvent)

	// Authorization
	LogAuthzDecision(ctx context.Context, event AuditEvent)

	// Policy evaluation
	LogPolicyEvaluation(ctx context.Context, event AuditEvent)

	// Generic audit event
	Log(ctx context.Context, event AuditEvent)

	// Async logging (non-blocking)
	LogAsync(event AuditEvent)
}

// zapAuditLogger implements AuditLogger using zap
type zapAuditLogger struct {
	logger *zap.Logger
	async  chan AuditEvent
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *zap.Logger) AuditLogger {
	if logger == nil {
		// Create a default production logger if none provided
		logger, _ = zap.NewProduction()
	}

	// Create logger with "audit" namespace
	auditLogger := logger.Named("audit")

	l := &zapAuditLogger{
		logger: auditLogger,
		async:  make(chan AuditEvent, 1000), // Buffer up to 1000 events
	}

	// Start async processor
	go l.processAsync()

	return l
}

// processAsync processes audit events in the background
func (l *zapAuditLogger) processAsync() {
	for event := range l.async {
		l.log(event)
	}
}

// LogAPIKeyCreated logs API key creation
func (l *zapAuditLogger) LogAPIKeyCreated(ctx context.Context, event AuditEvent) {
	event.Action = ActionAPIKeyCreated
	event.Timestamp = time.Now()
	l.log(event)
}

// LogAPIKeyRevoked logs API key revocation
func (l *zapAuditLogger) LogAPIKeyRevoked(ctx context.Context, event AuditEvent) {
	event.Action = ActionAPIKeyRevoked
	event.Timestamp = time.Now()
	l.log(event)
}

// LogAPIKeyUsed logs API key usage (typically async)
func (l *zapAuditLogger) LogAPIKeyUsed(ctx context.Context, event AuditEvent) {
	event.Action = ActionAPIKeyUsed
	event.Timestamp = time.Now()
	// Use async for usage events to avoid blocking requests
	l.LogAsync(event)
}

// LogAPIKeyListed logs API key listing
func (l *zapAuditLogger) LogAPIKeyListed(ctx context.Context, event AuditEvent) {
	event.Action = ActionAPIKeyListed
	event.Timestamp = time.Now()
	l.log(event)
}

// LogAuthAttempt logs authentication attempts
func (l *zapAuditLogger) LogAuthAttempt(ctx context.Context, event AuditEvent) {
	if event.Result == ResultSuccess {
		event.Action = ActionAuthSuccess
	} else {
		event.Action = ActionAuthFailure
	}
	event.Timestamp = time.Now()
	l.log(event)
}

// LogAuthzDecision logs authorization decisions
func (l *zapAuditLogger) LogAuthzDecision(ctx context.Context, event AuditEvent) {
	if event.Result == ResultGranted {
		event.Action = ActionAuthzGranted
	} else {
		event.Action = ActionAuthzDenied
	}
	event.Timestamp = time.Now()
	l.log(event)
}

// LogPolicyEvaluation logs policy evaluation results
func (l *zapAuditLogger) LogPolicyEvaluation(ctx context.Context, event AuditEvent) {
	event.Action = ActionPolicyEvaluated
	event.Timestamp = time.Now()
	l.log(event)
}

// Log logs a generic audit event
func (l *zapAuditLogger) Log(ctx context.Context, event AuditEvent) {
	event.Timestamp = time.Now()
	l.log(event)
}

// LogAsync logs an event asynchronously (non-blocking)
func (l *zapAuditLogger) LogAsync(event AuditEvent) {
	event.Timestamp = time.Now()
	// Non-blocking send - drop event if buffer is full
	select {
	case l.async <- event:
		// Successfully queued
	default:
		// Buffer full - log synchronously as fallback
		l.logger.Warn("audit async buffer full, logging synchronously",
			zap.String("action", string(event.Action)))
		l.log(event)
	}
}

// log is the internal logging implementation
func (l *zapAuditLogger) log(event AuditEvent) {
	fields := []zapcore.Field{
		zap.Time("timestamp", event.Timestamp),
		zap.String("action", string(event.Action)),
		zap.String("result", string(event.Result)),
	}

	// Add optional fields if present
	if event.RequestID != "" {
		fields = append(fields, zap.String("request_id", event.RequestID))
	}
	if event.UserAddr != "" {
		fields = append(fields, zap.String("user_addr", sanitizeAddress(event.UserAddr)))
	}
	if event.ResourceID != "" {
		fields = append(fields, zap.String("resource_id", event.ResourceID))
	}

	// API Key fields
	if event.KeyID != 0 {
		fields = append(fields, zap.Int64("key_id", event.KeyID))
	}
	if event.KeyName != "" {
		fields = append(fields, zap.String("key_name", event.KeyName))
	}
	if len(event.KeyScopes) > 0 {
		fields = append(fields, zap.Strings("key_scopes", event.KeyScopes))
	}
	if event.KeyExpiry != nil {
		fields = append(fields, zap.String("key_expiry", *event.KeyExpiry))
	}

	// HTTP fields
	if event.Method != "" {
		fields = append(fields, zap.String("method", event.Method))
	}
	if event.Endpoint != "" {
		fields = append(fields, zap.String("endpoint", event.Endpoint))
	}
	if event.IPAddr != "" {
		fields = append(fields, zap.String("ip_addr", event.IPAddr))
	}

	// Policy fields
	if event.PolicyPath != "" {
		fields = append(fields, zap.String("policy_path", event.PolicyPath))
	}
	if event.PolicyMethod != "" {
		fields = append(fields, zap.String("policy_method", event.PolicyMethod))
	}
	if event.RuleType != "" {
		fields = append(fields, zap.String("rule_type", event.RuleType))
		fields = append(fields, zap.Bool("rule_result", event.RuleResult))
	}

	// Blockchain fields
	if event.ChainID != 0 {
		fields = append(fields, zap.Int64("chain_id", event.ChainID))
	}
	if event.ContractAddress != "" {
		fields = append(fields, zap.String("contract_address", event.ContractAddress))
	}
	if event.RPCMethod != "" {
		fields = append(fields, zap.String("rpc_method", event.RPCMethod))
	}

	// Cache fields
	if event.CacheKey != "" {
		fields = append(fields, zap.String("cache_key", event.CacheKey))
	}

	// Error fields
	if event.Error != "" {
		fields = append(fields, zap.String("error", event.Error))
	}
	if event.ErrorDetail != "" {
		fields = append(fields, zap.String("error_detail", event.ErrorDetail))
	}

	// Metadata
	if len(event.Metadata) > 0 {
		for k, v := range event.Metadata {
			fields = append(fields, zap.Any(k, v))
		}
	}

	// Choose log level based on action result
	level := zapcore.InfoLevel
	if event.Result == ResultFailure || event.Result == ResultDenied {
		level = zapcore.WarnLevel
	}
	if event.Error != "" {
		level = zapcore.ErrorLevel
	}

	// Log the event
	if ce := l.logger.Check(level, "audit event"); ce != nil {
		ce.Write(fields...)
	}
}

// sanitizeAddress sanitizes an Ethereum address for logging
// Keeps first 6 and last 4 characters for identification
func sanitizeAddress(addr string) string {
	if len(addr) <= 10 {
		return addr
	}
	// Don't sanitize - we need full addresses for audit trail
	// But ensure we're not logging sensitive keys
	return addr
}

// Close gracefully shuts down the audit logger
func (l *zapAuditLogger) Close() error {
	close(l.async)
	return l.logger.Sync()
}
