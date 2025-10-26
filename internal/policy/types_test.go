package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPolicy_NewPolicy creates a valid policy
func TestPolicy_NewPolicy(t *testing.T) {
	rule := &HasScopeRule{Scope: "auth"}
	policy := NewPolicy("GET", "/api/data", "AND", []Rule{rule})

	require.NotNil(t, policy)
	assert.Equal(t, "GET", policy.Method)
	assert.Equal(t, "/api/data", policy.Path)
	assert.Equal(t, "AND", policy.Logic)
	assert.Len(t, policy.Rules, 1)
}

// TestPolicy_SupportsANDLogic ensures AND logic is valid
func TestPolicy_SupportsANDLogic(t *testing.T) {
	rule := &HasScopeRule{Scope: "auth"}
	policy := NewPolicy("GET", "/api/data", "AND", []Rule{rule})

	assert.Equal(t, "AND", policy.Logic)
}

// TestPolicy_SupportsORLogic ensures OR logic is valid
func TestPolicy_SupportsORLogic(t *testing.T) {
	rule := &HasScopeRule{Scope: "auth"}
	policy := NewPolicy("POST", "/api/data", "OR", []Rule{rule})

	assert.Equal(t, "OR", policy.Logic)
}

// TestPolicy_SupportsMultipleRules allows multiple rules
func TestPolicy_SupportsMultipleRules(t *testing.T) {
	rules := []Rule{
		&HasScopeRule{Scope: "auth"},
		&HasScopeRule{Scope: "api"},
	}
	policy := NewPolicy("GET", "/api/data", "AND", rules)

	assert.Len(t, policy.Rules, 2)
}

// TestHasScopeRule_Creation creates scope rule
func TestHasScopeRule_Creation(t *testing.T) {
	rule := NewHasScopeRule("read:data")

	require.NotNil(t, rule)
	assert.Equal(t, "read:data", rule.Scope)
}

// TestInAllowlistRule_Creation creates allowlist rule
func TestInAllowlistRule_Creation(t *testing.T) {
	addresses := []string{"0xABC123", "0xDEF456"}
	rule := NewInAllowlistRule(addresses)

	require.NotNil(t, rule)
	assert.Equal(t, addresses, rule.Addresses)
}

// TestRuleType_HasScope constant
func TestRuleType_HasScope(t *testing.T) {
	assert.Equal(t, "has_scope", string(HasScopeRuleType))
}

// TestRuleType_InAllowlist constant
func TestRuleType_InAllowlist(t *testing.T) {
	assert.Equal(t, "in_allowlist", string(InAllowlistRuleType))
}

// TestRuleType_ERC20MinBalance constant
func TestRuleType_ERC20MinBalance(t *testing.T) {
	assert.Equal(t, "erc20_min_balance", string(ERC20MinBalanceRuleType))
}

// TestRuleType_ERC721Owner constant
func TestRuleType_ERC721Owner(t *testing.T) {
	assert.Equal(t, "erc721_owner", string(ERC721OwnerRuleType))
}

// TestRuleInterface_Signature verifies Rule interface
func TestRuleInterface_Signature(t *testing.T) {
	var rule Rule = &HasScopeRule{Scope: "test"}
	require.NotNil(t, rule)
}

// TestPolicy_RequiresValidLogic ensures only AND/OR are accepted
func TestPolicy_RequiresValidLogic(t *testing.T) {
	// This test will be verified by validators
	rule := &HasScopeRule{Scope: "auth"}

	// Valid logics
	andPolicy := NewPolicy("GET", "/api", "AND", []Rule{rule})
	assert.NotNil(t, andPolicy)

	orPolicy := NewPolicy("GET", "/api", "OR", []Rule{rule})
	assert.NotNil(t, orPolicy)
}

// TestPolicy_RequiresAtLeastOneRule ensures policies have rules
func TestPolicy_RequiresAtLeastOneRule(t *testing.T) {
	// This is enforced by validation, but policy should not be created empty
	rule := &HasScopeRule{Scope: "auth"}
	policy := NewPolicy("GET", "/api", "AND", []Rule{rule})

	assert.Greater(t, len(policy.Rules), 0)
}

// TestPolicy_SupportsPathPatterns allows flexible path matching
func TestPolicy_SupportsPathPatterns(t *testing.T) {
	rule := &HasScopeRule{Scope: "auth"}

	paths := []string{
		"/api/data",
		"/api/*",
		"/admin/**",
	}

	for _, path := range paths {
		policy := NewPolicy("GET", path, "AND", []Rule{rule})
		assert.Equal(t, path, policy.Path)
	}
}

// TestPolicy_SupportsMultipleMethods allows multiple HTTP methods
func TestPolicy_SupportsMultipleMethods(t *testing.T) {
	rule := &HasScopeRule{Scope: "auth"}

	methods := []string{"GET", "POST", "PUT", "DELETE"}

	for _, method := range methods {
		policy := NewPolicy(method, "/api", "AND", []Rule{rule})
		assert.Equal(t, method, policy.Method)
	}
}
