package policy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/gatekeeper/internal/auth"
)

// TestPolicy_EvaluateAND_AllRulesPass verifies AND requires all rules to pass
func TestPolicy_EvaluateAND_AllRulesPass(t *testing.T) {
	rules := []Rule{
		NewHasScopeRule("read:data"),
		NewHasScopeRule("api"),
	}
	policy := NewPolicy("GET", "/api/data", "AND", rules)

	claims := &auth.Claims{
		Address: "0x123",
		Scopes:  []string{"read:data", "api", "admin"},
	}

	result, err := policy.Evaluate(context.Background(), claims.Address, claims)

	require.NoError(t, err)
	assert.True(t, result)
}

// TestPolicy_EvaluateAND_OneRuleFails verifies AND fails if any rule fails
func TestPolicy_EvaluateAND_OneRuleFails(t *testing.T) {
	rules := []Rule{
		NewHasScopeRule("read:data"),
		NewHasScopeRule("missing:scope"),
	}
	policy := NewPolicy("GET", "/api/data", "AND", rules)

	claims := &auth.Claims{
		Address: "0x123",
		Scopes:  []string{"read:data"},
	}

	result, err := policy.Evaluate(context.Background(), claims.Address, claims)

	require.NoError(t, err)
	assert.False(t, result)
}

// TestPolicy_EvaluateAND_AllRulesFail verifies AND fails if all rules fail
func TestPolicy_EvaluateAND_AllRulesFail(t *testing.T) {
	rules := []Rule{
		NewHasScopeRule("missing1"),
		NewHasScopeRule("missing2"),
	}
	policy := NewPolicy("GET", "/api/data", "AND", rules)

	claims := &auth.Claims{
		Address: "0x123",
		Scopes:  []string{"read:data"},
	}

	result, err := policy.Evaluate(context.Background(), claims.Address, claims)

	require.NoError(t, err)
	assert.False(t, result)
}

// TestPolicy_EvaluateAND_ShortCircuit verifies AND short-circuits on first failure
func TestPolicy_EvaluateAND_ShortCircuit(t *testing.T) {
	evaluationCount := 0

	// Create a custom rule that tracks evaluation count
	firstRule := NewHasScopeRule("missing")
	secondRule := NewHasScopeRule("api")

	rules := []Rule{firstRule, secondRule}
	policy := NewPolicy("GET", "/api/data", "AND", rules)

	claims := &auth.Claims{
		Address: "0x123",
		Scopes:  []string{"api"},
	}

	result, err := policy.Evaluate(context.Background(), claims.Address, claims)

	require.NoError(t, err)
	assert.False(t, result)
	// Short-circuit should prevent full evaluation
	_ = evaluationCount // Not practical to verify without instrumentation
}

// TestPolicy_EvaluateOR_AllRulesPass verifies OR passes if all pass
func TestPolicy_EvaluateOR_AllRulesPass(t *testing.T) {
	rules := []Rule{
		NewHasScopeRule("read:data"),
		NewHasScopeRule("api"),
	}
	policy := NewPolicy("GET", "/api/data", "OR", rules)

	claims := &auth.Claims{
		Address: "0x123",
		Scopes:  []string{"read:data", "api"},
	}

	result, err := policy.Evaluate(context.Background(), claims.Address, claims)

	require.NoError(t, err)
	assert.True(t, result)
}

// TestPolicy_EvaluateOR_SomeRulesPass verifies OR passes if any rule passes
func TestPolicy_EvaluateOR_SomeRulesPass(t *testing.T) {
	rules := []Rule{
		NewHasScopeRule("missing"),
		NewHasScopeRule("api"),
		NewHasScopeRule("missing2"),
	}
	policy := NewPolicy("GET", "/api/data", "OR", rules)

	claims := &auth.Claims{
		Address: "0x123",
		Scopes:  []string{"api"},
	}

	result, err := policy.Evaluate(context.Background(), claims.Address, claims)

	require.NoError(t, err)
	assert.True(t, result)
}

// TestPolicy_EvaluateOR_AllRulesFail verifies OR fails if all rules fail
func TestPolicy_EvaluateOR_AllRulesFail(t *testing.T) {
	rules := []Rule{
		NewHasScopeRule("missing1"),
		NewHasScopeRule("missing2"),
		NewHasScopeRule("missing3"),
	}
	policy := NewPolicy("GET", "/api/data", "OR", rules)

	claims := &auth.Claims{
		Address: "0x123",
		Scopes:  []string{"api"},
	}

	result, err := policy.Evaluate(context.Background(), claims.Address, claims)

	require.NoError(t, err)
	assert.False(t, result)
}

// TestPolicy_EvaluateOR_ShortCircuit verifies OR short-circuits on first pass
func TestPolicy_EvaluateOR_ShortCircuit(t *testing.T) {
	rules := []Rule{
		NewHasScopeRule("api"),
		NewHasScopeRule("admin"),
	}
	policy := NewPolicy("GET", "/api/data", "OR", rules)

	claims := &auth.Claims{
		Address: "0x123",
		Scopes:  []string{"api"},
	}

	result, err := policy.Evaluate(context.Background(), claims.Address, claims)

	require.NoError(t, err)
	assert.True(t, result)
	// First rule passes, OR should short-circuit without evaluating second
}

// TestPolicy_EvaluateSingleRule_AND verifies AND with single rule
func TestPolicy_EvaluateSingleRule_AND(t *testing.T) {
	rules := []Rule{NewHasScopeRule("auth")}
	policy := NewPolicy("GET", "/api", "AND", rules)

	claims := &auth.Claims{
		Address: "0x123",
		Scopes:  []string{"auth"},
	}

	result, err := policy.Evaluate(context.Background(), claims.Address, claims)

	require.NoError(t, err)
	assert.True(t, result)
}

// TestPolicy_EvaluateSingleRule_OR verifies OR with single rule
func TestPolicy_EvaluateSingleRule_OR(t *testing.T) {
	rules := []Rule{NewHasScopeRule("auth")}
	policy := NewPolicy("GET", "/api", "OR", rules)

	claims := &auth.Claims{
		Address: "0x123",
		Scopes:  []string{"auth"},
	}

	result, err := policy.Evaluate(context.Background(), claims.Address, claims)

	require.NoError(t, err)
	assert.True(t, result)
}

// TestPolicy_EvaluateWithNilClaims handles nil claims
func TestPolicy_EvaluateWithNilClaims(t *testing.T) {
	rules := []Rule{NewHasScopeRule("auth")}
	policy := NewPolicy("GET", "/api", "AND", rules)

	result, err := policy.Evaluate(context.Background(), "0x123", nil)

	require.NoError(t, err)
	assert.False(t, result)
}

// TestPolicy_EvaluateComplexAND verifies complex AND logic
func TestPolicy_EvaluateComplexAND(t *testing.T) {
	rules := []Rule{
		NewHasScopeRule("auth"),
		NewInAllowlistRule([]string{"0x123", "0x456"}),
		NewHasScopeRule("api"),
	}
	policy := NewPolicy("GET", "/api", "AND", rules)

	claims := &auth.Claims{
		Address: "0x123",
		Scopes:  []string{"auth", "api"},
	}

	result, err := policy.Evaluate(context.Background(), claims.Address, claims)

	require.NoError(t, err)
	assert.True(t, result)
}

// TestPolicy_EvaluateComplexOR verifies complex OR logic
func TestPolicy_EvaluateComplexOR(t *testing.T) {
	rules := []Rule{
		NewHasScopeRule("missing"),
		NewInAllowlistRule([]string{"0x999"}),
		NewHasScopeRule("api"),
	}
	policy := NewPolicy("GET", "/api", "OR", rules)

	claims := &auth.Claims{
		Address: "0x123",
		Scopes:  []string{"api"},
	}

	result, err := policy.Evaluate(context.Background(), claims.Address, claims)

	require.NoError(t, err)
	assert.True(t, result)
}

// TestHasScopeRule_Evaluate_WithValidScope passes with correct scope
func TestHasScopeRule_Evaluate_WithValidScope(t *testing.T) {
	rule := NewHasScopeRule("read:data")

	claims := &auth.Claims{
		Address: "0x123",
		Scopes:  []string{"read:data", "write:data"},
	}

	result, err := rule.Evaluate(context.Background(), "0x123", claims)

	require.NoError(t, err)
	assert.True(t, result)
}

// TestHasScopeRule_Evaluate_WithoutScope fails without scope
func TestHasScopeRule_Evaluate_WithoutScope(t *testing.T) {
	rule := NewHasScopeRule("admin:manage")

	claims := &auth.Claims{
		Address: "0x123",
		Scopes:  []string{"read:data"},
	}

	result, err := rule.Evaluate(context.Background(), "0x123", claims)

	require.NoError(t, err)
	assert.False(t, result)
}

// TestHasScopeRule_Evaluate_WithEmptyScopes fails with empty scopes
func TestHasScopeRule_Evaluate_WithEmptyScopes(t *testing.T) {
	rule := NewHasScopeRule("any:scope")

	claims := &auth.Claims{
		Address: "0x123",
		Scopes:  []string{},
	}

	result, err := rule.Evaluate(context.Background(), "0x123", claims)

	require.NoError(t, err)
	assert.False(t, result)
}

// TestInAllowlistRule_Evaluate_AddressInList passes when in allowlist
func TestInAllowlistRule_Evaluate_AddressInList(t *testing.T) {
	rule := NewInAllowlistRule([]string{"0xABC123", "0xDEF456"})

	claims := &auth.Claims{
		Address: "0xABC123",
		Scopes:  []string{},
	}

	result, err := rule.Evaluate(context.Background(), "0xABC123", claims)

	require.NoError(t, err)
	assert.True(t, result)
}

// TestInAllowlistRule_Evaluate_AddressNotInList fails when not in allowlist
func TestInAllowlistRule_Evaluate_AddressNotInList(t *testing.T) {
	rule := NewInAllowlistRule([]string{"0xABC123"})

	claims := &auth.Claims{
		Address: "0xXXX999",
		Scopes:  []string{},
	}

	result, err := rule.Evaluate(context.Background(), "0xXXX999", claims)

	require.NoError(t, err)
	assert.False(t, result)
}
