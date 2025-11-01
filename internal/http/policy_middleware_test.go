package http

import (
	"context"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/gatekeeper/internal/auth"
	"github.com/yourusername/gatekeeper/internal/log"
	"github.com/yourusername/gatekeeper/internal/policy"
)

// TestPolicyMiddleware_NoClaims rejects requests without authentication
func TestPolicyMiddleware_NoClaims(t *testing.T) {
	pm := policy.NewPolicyManager(&mockBlockchainProvider{}, &mockCache{})
	logger, err := log.New("debug")
	require.NoError(t, err)
	defer logger.Close()
	middleware := NewPolicyMiddleware(pm, logger, nil)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("GET", "/api/data", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "Unauthorized\n", w.Body.String())
}

// TestPolicyMiddleware_NoPolicies allows requests when no policies exist
func TestPolicyMiddleware_NoPolicies(t *testing.T) {
	pm := policy.NewPolicyManager(&mockBlockchainProvider{}, &mockCache{})
	logger, err := log.New("debug")
	require.NoError(t, err)
	defer logger.Close()
	middleware := NewPolicyMiddleware(pm, logger, nil)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	claims := &auth.Claims{
		Address: "0x1234567890abcdef1234567890abcdef12345678",
		Scopes:  []string{},
	}

	req := httptest.NewRequest("GET", "/api/data", nil)
	req = req.WithContext(context.WithValue(req.Context(), "claims", claims))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())
}

// TestPolicyMiddleware_PassingPolicy allows requests that pass policy
func TestPolicyMiddleware_PassingPolicy(t *testing.T) {
	pm := policy.NewPolicyManager(&mockBlockchainProvider{}, &mockCache{})
	logger, err := log.New("debug")
	require.NoError(t, err)
	defer logger.Close()
	middleware := NewPolicyMiddleware(pm, logger, nil)

	// Add policy requiring admin scope
	adminPolicy := policy.NewPolicy(
		"GET",
		"/api/admin",
		"AND",
		[]policy.Rule{
			policy.NewHasScopeRule("admin"),
		},
	)
	pm.AddPolicy(adminPolicy)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	claims := &auth.Claims{
		Address: "0x1234567890abcdef1234567890abcdef12345678",
		Scopes:  []string{"admin"},
	}

	req := httptest.NewRequest("GET", "/api/admin", nil)
	req = req.WithContext(context.WithValue(req.Context(), "claims", claims))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())
}

// TestPolicyMiddleware_FailingPolicy denies requests that fail policy
func TestPolicyMiddleware_FailingPolicy(t *testing.T) {
	pm := policy.NewPolicyManager(&mockBlockchainProvider{}, &mockCache{})
	logger, err := log.New("debug")
	require.NoError(t, err)
	defer logger.Close()
	middleware := NewPolicyMiddleware(pm, logger, nil)

	// Add policy requiring admin scope
	adminPolicy := policy.NewPolicy(
		"GET",
		"/api/admin",
		"AND",
		[]policy.Rule{
			policy.NewHasScopeRule("admin"),
		},
	)
	pm.AddPolicy(adminPolicy)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	claims := &auth.Claims{
		Address: "0x1234567890abcdef1234567890abcdef12345678",
		Scopes:  []string{}, // No admin scope
	}

	req := httptest.NewRequest("GET", "/api/admin", nil)
	req = req.WithContext(context.WithValue(req.Context(), "claims", claims))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "Forbidden\n", w.Body.String())
}

// TestPolicyMiddleware_AllowlistPolicy checks address in allowlist
func TestPolicyMiddleware_AllowlistPolicy(t *testing.T) {
	pm := policy.NewPolicyManager(&mockBlockchainProvider{}, &mockCache{})
	logger, err := log.New("debug")
	require.NoError(t, err)
	defer logger.Close()
	middleware := NewPolicyMiddleware(pm, logger, nil)

	allowedAddr := "0x1234567890abcdef1234567890abcdef12345678"
	otherAddr := "0xabcdef1234567890abcdef1234567890abcdef12"

	// Add allowlist policy
	allowlistPolicy := policy.NewPolicy(
		"POST",
		"/api/transfer",
		"AND",
		[]policy.Rule{
			policy.NewInAllowlistRule([]string{allowedAddr}),
		},
	)
	pm.AddPolicy(allowlistPolicy)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("transfer successful"))
	}))

	// Test with allowed address
	allowedClaims := &auth.Claims{
		Address: allowedAddr,
		Scopes:  []string{},
	}

	allowedReq := httptest.NewRequest("POST", "/api/transfer", nil)
	allowedReq = allowedReq.WithContext(context.WithValue(allowedReq.Context(), "claims", allowedClaims))
	allowedW := httptest.NewRecorder()

	handler.ServeHTTP(allowedW, allowedReq)
	assert.Equal(t, http.StatusOK, allowedW.Code)

	// Test with disallowed address
	disallowedClaims := &auth.Claims{
		Address: otherAddr,
		Scopes:  []string{},
	}

	disallowedReq := httptest.NewRequest("POST", "/api/transfer", nil)
	disallowedReq = disallowedReq.WithContext(context.WithValue(disallowedReq.Context(), "claims", disallowedClaims))
	disallowedW := httptest.NewRecorder()

	handler.ServeHTTP(disallowedW, disallowedReq)
	assert.Equal(t, http.StatusForbidden, disallowedW.Code)
}

// TestPolicyMiddleware_MultipleRulesAND requires all rules to pass
func TestPolicyMiddleware_MultipleRulesAND(t *testing.T) {
	pm := policy.NewPolicyManager(&mockBlockchainProvider{}, &mockCache{})
	logger, err := log.New("debug")
	require.NoError(t, err)
	defer logger.Close()
	middleware := NewPolicyMiddleware(pm, logger, nil)

	userAddr := "0x1234567890abcdef1234567890abcdef12345678"

	// Policy requiring both admin scope AND address in allowlist
	restrictivePolicy := policy.NewPolicy(
		"DELETE",
		"/api/resource",
		"AND",
		[]policy.Rule{
			policy.NewHasScopeRule("admin"),
			policy.NewInAllowlistRule([]string{userAddr}),
		},
	)
	pm.AddPolicy(restrictivePolicy)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("deleted"))
	}))

	// Has scope but not in allowlist - should fail
	claims1 := &auth.Claims{
		Address: "0xotheraddress", // Different address
		Scopes:  []string{"admin"},
	}

	req1 := httptest.NewRequest("DELETE", "/api/resource", nil)
	req1 = req1.WithContext(context.WithValue(req1.Context(), "claims", claims1))
	w1 := httptest.NewRecorder()

	handler.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusForbidden, w1.Code)

	// In allowlist but no scope - should fail
	claims2 := &auth.Claims{
		Address: userAddr,
		Scopes:  []string{}, // No admin scope
	}

	req2 := httptest.NewRequest("DELETE", "/api/resource", nil)
	req2 = req2.WithContext(context.WithValue(req2.Context(), "claims", claims2))
	w2 := httptest.NewRecorder()

	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusForbidden, w2.Code)

	// Has both - should pass
	claims3 := &auth.Claims{
		Address: userAddr,
		Scopes:  []string{"admin"},
	}

	req3 := httptest.NewRequest("DELETE", "/api/resource", nil)
	req3 = req3.WithContext(context.WithValue(req3.Context(), "claims", claims3))
	w3 := httptest.NewRecorder()

	handler.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}

// TestPolicyMiddleware_DifferentMethods distinguishes between HTTP methods
func TestPolicyMiddleware_DifferentMethods(t *testing.T) {
	pm := policy.NewPolicyManager(&mockBlockchainProvider{}, &mockCache{})
	logger, err := log.New("debug")
	require.NoError(t, err)
	defer logger.Close()
	middleware := NewPolicyMiddleware(pm, logger, nil)

	userAddr := "0x1234567890abcdef1234567890abcdef12345678"

	// Only GET requires admin
	getPolicy := policy.NewPolicy(
		"GET",
		"/api/data",
		"AND",
		[]policy.Rule{
			policy.NewHasScopeRule("admin"),
		},
	)
	pm.AddPolicy(getPolicy)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	claims := &auth.Claims{
		Address: userAddr,
		Scopes:  []string{}, // No admin scope
	}

	// GET should fail
	getReq := httptest.NewRequest("GET", "/api/data", nil)
	getReq = getReq.WithContext(context.WithValue(getReq.Context(), "claims", claims))
	getW := httptest.NewRecorder()

	handler.ServeHTTP(getW, getReq)
	assert.Equal(t, http.StatusForbidden, getW.Code)

	// POST should pass (no policy)
	postReq := httptest.NewRequest("POST", "/api/data", nil)
	postReq = postReq.WithContext(context.WithValue(postReq.Context(), "claims", claims))
	postW := httptest.NewRecorder()

	handler.ServeHTTP(postW, postReq)
	assert.Equal(t, http.StatusOK, postW.Code)
}

// TestPolicyMiddleware_URLPath extracts path without query parameters
func TestPolicyMiddleware_URLPath(t *testing.T) {
	pm := policy.NewPolicyManager(&mockBlockchainProvider{}, &mockCache{})
	logger, err := log.New("debug")
	require.NoError(t, err)
	defer logger.Close()
	middleware := NewPolicyMiddleware(pm, logger, nil)

	// Policy for exact path
	dataPolicy := policy.NewPolicy(
		"GET",
		"/api/data",
		"AND",
		[]policy.Rule{
			policy.NewHasScopeRule("read"),
		},
	)
	pm.AddPolicy(dataPolicy)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data"))
	}))

	claims := &auth.Claims{
		Address: "0x1234567890abcdef1234567890abcdef12345678",
		Scopes:  []string{"read"},
	}

	// URL with query parameters should still match policy for path
	req := httptest.NewRequest("GET", "/api/data?filter=active&limit=10", nil)
	req = req.WithContext(context.WithValue(req.Context(), "claims", claims))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestPolicyMiddleware_CaseInsensitiveAddresses handles address case variations
func TestPolicyMiddleware_CaseInsensitiveAddresses(t *testing.T) {
	pm := policy.NewPolicyManager(&mockBlockchainProvider{}, &mockCache{})
	logger, err := log.New("debug")
	require.NoError(t, err)
	defer logger.Close()
	middleware := NewPolicyMiddleware(pm, logger, nil)

	// Address in different cases
	lowerAddr := "0x1234567890abcdef1234567890abcdef12345678"
	upperAddr := "0x1234567890ABCDEF1234567890ABCDEF12345678"

	allowlistPolicy := policy.NewPolicy(
		"GET",
		"/api/user",
		"AND",
		[]policy.Rule{
			policy.NewInAllowlistRule([]string{lowerAddr}),
		},
	)
	pm.AddPolicy(allowlistPolicy)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("user data"))
	}))

	// Request with uppercase address should match
	claims := &auth.Claims{
		Address: upperAddr,
		Scopes:  []string{},
	}

	req := httptest.NewRequest("GET", "/api/user", nil)
	req = req.WithContext(context.WithValue(req.Context(), "claims", claims))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestPolicyMiddleware_ContextChaining preserves claims in context
func TestPolicyMiddleware_ContextChaining(t *testing.T) {
	pm := policy.NewPolicyManager(&mockBlockchainProvider{}, &mockCache{})
	logger, err := log.New("debug")
	require.NoError(t, err)
	defer logger.Close()
	middleware := NewPolicyMiddleware(pm, logger, nil)

	userAddr := "0x1234567890abcdef1234567890abcdef12345678"

	// Handler that reads claims from context
	var handlerClaims *auth.Claims
	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if claims, ok := r.Context().Value("claims").(*auth.Claims); ok {
			handlerClaims = claims
		}
		w.WriteHeader(http.StatusOK)
	}))

	claims := &auth.Claims{
		Address: userAddr,
		Scopes:  []string{"read"},
	}

	req := httptest.NewRequest("GET", "/api/data", nil)
	req = req.WithContext(context.WithValue(req.Context(), "claims", claims))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotNil(t, handlerClaims)
	assert.Equal(t, userAddr, handlerClaims.Address)
	assert.Equal(t, []string{"read"}, handlerClaims.Scopes)
}

// TestPolicyMiddleware_MultipleRoutes handles different routes independently
func TestPolicyMiddleware_MultipleRoutes(t *testing.T) {
	pm := policy.NewPolicyManager(&mockBlockchainProvider{}, &mockCache{})
	logger, err := log.New("debug")
	require.NoError(t, err)
	defer logger.Close()
	middleware := NewPolicyMiddleware(pm, logger, nil)

	userAddr := "0x1234567890abcdef1234567890abcdef12345678"

	// Add policies for different routes
	policy1 := policy.NewPolicy(
		"GET",
		"/api/public",
		"AND",
		[]policy.Rule{},
	)
	policy2 := policy.NewPolicy(
		"GET",
		"/api/protected",
		"AND",
		[]policy.Rule{
			policy.NewHasScopeRule("admin"),
		},
	)

	pm.AddPolicy(policy1)
	pm.AddPolicy(policy2)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	claims := &auth.Claims{
		Address: userAddr,
		Scopes:  []string{}, // No admin scope
	}

	// Public route should pass
	req1 := httptest.NewRequest("GET", "/api/public", nil)
	req1 = req1.WithContext(context.WithValue(req1.Context(), "claims", claims))
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Protected route should fail
	req2 := httptest.NewRequest("GET", "/api/protected", nil)
	req2 = req2.WithContext(context.WithValue(req2.Context(), "claims", claims))
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusForbidden, w2.Code)
}

// Mock implementations for testing

type mockBlockchainProvider struct {
	balances map[string]*big.Int
	owners   map[string]string
}

func (m *mockBlockchainProvider) Call(ctx context.Context, method string, params []interface{}) ([]byte, error) {
	return []byte(`{"jsonrpc":"2.0","result":"0x1","id":1}`), nil
}

func (m *mockBlockchainProvider) HealthCheck(ctx context.Context) bool {
	return true
}

func (m *mockBlockchainProvider) SetBalance(address string, balance *big.Int) {
	if m.balances == nil {
		m.balances = make(map[string]*big.Int)
	}
	m.balances[address] = balance
}

type mockCache struct {
	data map[string]interface{}
}

func (m *mockCache) Get(key string) (interface{}, bool) {
	val, ok := m.data[key]
	return val, ok
}

func (m *mockCache) Set(key string, value interface{}) {
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	m.data[key] = value
}

func (m *mockCache) GetOrSet(key string, fn func() interface{}) interface{} {
	if val, ok := m.Get(key); ok {
		return val
	}
	val := fn()
	m.Set(key, val)
	return val
}
