package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoader_ValidPolicyConfiguration loads and validates valid config
func TestLoader_ValidPolicyConfiguration(t *testing.T) {
	configJSON := `[
		{
			"path": "/api/data",
			"method": "GET",
			"logic": "AND",
			"rules": [
				{
					"type": "has_scope",
					"scope": "read:data"
				}
			]
		}
	]`

	loader := NewPolicyLoader()
	policies, err := loader.LoadFromJSON([]byte(configJSON))

	require.NoError(t, err)
	assert.Len(t, policies, 1)
	assert.Equal(t, "/api/data", policies[0].Path)
	assert.Equal(t, "GET", policies[0].Method)
	assert.Equal(t, "AND", policies[0].Logic)
}

// TestLoader_InvalidJSON returns error on malformed JSON
func TestLoader_InvalidJSON(t *testing.T) {
	configJSON := `invalid json {`

	loader := NewPolicyLoader()
	_, err := loader.LoadFromJSON([]byte(configJSON))

	assert.Error(t, err)
}

// TestLoader_MissingPath validates required field
func TestLoader_MissingPath(t *testing.T) {
	configJSON := `[
		{
			"method": "GET",
			"logic": "AND",
			"rules": [{"type": "has_scope", "scope": "auth"}]
		}
	]`

	loader := NewPolicyLoader()
	_, err := loader.LoadFromJSON([]byte(configJSON))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path")
}

// TestLoader_MissingMethod validates required field
func TestLoader_MissingMethod(t *testing.T) {
	configJSON := `[
		{
			"path": "/api/data",
			"logic": "AND",
			"rules": [{"type": "has_scope", "scope": "auth"}]
		}
	]`

	loader := NewPolicyLoader()
	_, err := loader.LoadFromJSON([]byte(configJSON))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "method")
}

// TestLoader_MissingLogic validates required field
func TestLoader_MissingLogic(t *testing.T) {
	configJSON := `[
		{
			"path": "/api/data",
			"method": "GET",
			"rules": [{"type": "has_scope", "scope": "auth"}]
		}
	]`

	loader := NewPolicyLoader()
	_, err := loader.LoadFromJSON([]byte(configJSON))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logic")
}

// TestLoader_MissingRules validates required field
func TestLoader_MissingRules(t *testing.T) {
	configJSON := `[
		{
			"path": "/api/data",
			"method": "GET",
			"logic": "AND"
		}
	]`

	loader := NewPolicyLoader()
	_, err := loader.LoadFromJSON([]byte(configJSON))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rules")
}

// TestLoader_EmptyRules validates rules array is not empty
func TestLoader_EmptyRules(t *testing.T) {
	configJSON := `[
		{
			"path": "/api/data",
			"method": "GET",
			"logic": "AND",
			"rules": []
		}
	]`

	loader := NewPolicyLoader()
	_, err := loader.LoadFromJSON([]byte(configJSON))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one rule")
}

// TestLoader_InvalidLogic validates logic is AND or OR
func TestLoader_InvalidLogic(t *testing.T) {
	configJSON := `[
		{
			"path": "/api/data",
			"method": "GET",
			"logic": "INVALID",
			"rules": [{"type": "has_scope", "scope": "auth"}]
		}
	]`

	loader := NewPolicyLoader()
	_, err := loader.LoadFromJSON([]byte(configJSON))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logic")
	assert.Contains(t, err.Error(), "AND")
}

// TestLoader_UnknownRuleType validates rule type
func TestLoader_UnknownRuleType(t *testing.T) {
	configJSON := `[
		{
			"path": "/api/data",
			"method": "GET",
			"logic": "AND",
			"rules": [{"type": "unknown_rule_type"}]
		}
	]`

	loader := NewPolicyLoader()
	_, err := loader.LoadFromJSON([]byte(configJSON))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown")
}

// TestLoader_HasScopeRuleMissingScope validates scope field
func TestLoader_HasScopeRuleMissingScope(t *testing.T) {
	configJSON := `[
		{
			"path": "/api/data",
			"method": "GET",
			"logic": "AND",
			"rules": [{"type": "has_scope"}]
		}
	]`

	loader := NewPolicyLoader()
	_, err := loader.LoadFromJSON([]byte(configJSON))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scope")
}

// TestLoader_InAllowlistRuleMissingAddresses validates addresses field
func TestLoader_InAllowlistRuleMissingAddresses(t *testing.T) {
	configJSON := `[
		{
			"path": "/api/data",
			"method": "GET",
			"logic": "AND",
			"rules": [{"type": "in_allowlist"}]
		}
	]`

	loader := NewPolicyLoader()
	_, err := loader.LoadFromJSON([]byte(configJSON))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "addresses")
}

// TestLoader_ERC20RuleMissingRequired validates ERC20 required fields
func TestLoader_ERC20RuleMissingRequired(t *testing.T) {
	configJSON := `[
		{
			"path": "/api/data",
			"method": "GET",
			"logic": "AND",
			"rules": [{"type": "erc20_min_balance"}]
		}
	]`

	loader := NewPolicyLoader()
	_, err := loader.LoadFromJSON([]byte(configJSON))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contract_address")
}

// TestLoader_MultiplePolicies loads multiple policies
func TestLoader_MultiplePolicies(t *testing.T) {
	configJSON := `[
		{
			"path": "/api/users",
			"method": "GET",
			"logic": "AND",
			"rules": [{"type": "has_scope", "scope": "read:users"}]
		},
		{
			"path": "/api/admin",
			"method": "POST",
			"logic": "OR",
			"rules": [{"type": "has_scope", "scope": "admin"}]
		}
	]`

	loader := NewPolicyLoader()
	policies, err := loader.LoadFromJSON([]byte(configJSON))

	require.NoError(t, err)
	assert.Len(t, policies, 2)
	assert.Equal(t, "/api/users", policies[0].Path)
	assert.Equal(t, "/api/admin", policies[1].Path)
}

// TestLoader_ComplexPolicy with multiple rules
func TestLoader_ComplexPolicy(t *testing.T) {
	configJSON := `[
		{
			"path": "/api/premium",
			"method": "GET",
			"logic": "AND",
			"rules": [
				{"type": "has_scope", "scope": "premium"},
				{"type": "in_allowlist", "addresses": ["0xABC123", "0xDEF456"]},
				{"type": "erc20_min_balance", "contract_address": "0x1234", "minimum_balance": "1000000000000000000", "chain_id": 1}
			]
		}
	]`

	loader := NewPolicyLoader()
	policies, err := loader.LoadFromJSON([]byte(configJSON))

	require.NoError(t, err)
	assert.Len(t, policies, 1)
	assert.Len(t, policies[0].Rules, 3)
}

// TestLoader_EmptyConfiguration loads empty policy list
func TestLoader_EmptyConfiguration(t *testing.T) {
	configJSON := `[]`

	loader := NewPolicyLoader()
	policies, err := loader.LoadFromJSON([]byte(configJSON))

	require.NoError(t, err)
	assert.Len(t, policies, 0)
}

// TestLoader_HasScopeRuleCreation verifies has_scope rule is created
func TestLoader_HasScopeRuleCreation(t *testing.T) {
	configJSON := `[
		{
			"path": "/api/data",
			"method": "GET",
			"logic": "AND",
			"rules": [{"type": "has_scope", "scope": "read:data"}]
		}
	]`

	loader := NewPolicyLoader()
	policies, err := loader.LoadFromJSON([]byte(configJSON))

	require.NoError(t, err)
	rule := policies[0].Rules[0]
	_, ok := rule.(*HasScopeRule)
	assert.True(t, ok, "Rule should be HasScopeRule")
}

// TestLoader_InAllowlistRuleCreation verifies in_allowlist rule is created
func TestLoader_InAllowlistRuleCreation(t *testing.T) {
	configJSON := `[
		{
			"path": "/api/data",
			"method": "GET",
			"logic": "AND",
			"rules": [{"type": "in_allowlist", "addresses": ["0x123"]}]
		}
	]`

	loader := NewPolicyLoader()
	policies, err := loader.LoadFromJSON([]byte(configJSON))

	require.NoError(t, err)
	rule := policies[0].Rules[0]
	_, ok := rule.(*InAllowlistRule)
	assert.True(t, ok, "Rule should be InAllowlistRule")
}
