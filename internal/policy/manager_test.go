package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestManager_NewManager creates a policy manager
func TestManager_NewManager(t *testing.T) {
	manager := NewPolicyManager()

	require.NotNil(t, manager)
	assert.NotNil(t, manager.policies)
}

// TestManager_AddPolicy adds a policy to the manager
func TestManager_AddPolicy(t *testing.T) {
	manager := NewPolicyManager()
	rule := NewHasScopeRule("auth")
	policy := NewPolicy("GET", "/api/data", "AND", []Rule{rule})

	manager.AddPolicy(policy)

	assert.Equal(t, 1, len(manager.policies))
}

// TestManager_AddMultiplePolicies adds multiple policies
func TestManager_AddMultiplePolicies(t *testing.T) {
	manager := NewPolicyManager()
	rule := NewHasScopeRule("auth")

	policy1 := NewPolicy("GET", "/api/data", "AND", []Rule{rule})
	policy2 := NewPolicy("POST", "/api/data", "AND", []Rule{rule})
	policy3 := NewPolicy("GET", "/api/users", "AND", []Rule{rule})

	manager.AddPolicy(policy1)
	manager.AddPolicy(policy2)
	manager.AddPolicy(policy3)

	assert.Equal(t, 3, len(manager.policies))
}

// TestManager_GetPoliciesForRoute returns matching policies
func TestManager_GetPoliciesForRoute(t *testing.T) {
	manager := NewPolicyManager()
	rule := NewHasScopeRule("auth")

	policy := NewPolicy("GET", "/api/data", "AND", []Rule{rule})
	manager.AddPolicy(policy)

	policies := manager.GetPoliciesForRoute("/api/data", "GET")

	assert.Len(t, policies, 1)
	assert.Equal(t, "/api/data", policies[0].Path)
}

// TestManager_GetPoliciesForRoute_NoMatch returns empty for no match
func TestManager_GetPoliciesForRoute_NoMatch(t *testing.T) {
	manager := NewPolicyManager()
	rule := NewHasScopeRule("auth")

	policy := NewPolicy("GET", "/api/data", "AND", []Rule{rule})
	manager.AddPolicy(policy)

	policies := manager.GetPoliciesForRoute("/api/users", "GET")

	assert.Len(t, policies, 0)
}

// TestManager_GetPoliciesForRoute_DifferentMethod returns empty for different method
func TestManager_GetPoliciesForRoute_DifferentMethod(t *testing.T) {
	manager := NewPolicyManager()
	rule := NewHasScopeRule("auth")

	policy := NewPolicy("GET", "/api/data", "AND", []Rule{rule})
	manager.AddPolicy(policy)

	policies := manager.GetPoliciesForRoute("/api/data", "POST")

	assert.Len(t, policies, 0)
}

// TestManager_GetPoliciesForRoute_MultipleMatches returns all matching policies
func TestManager_GetPoliciesForRoute_MultipleMatches(t *testing.T) {
	manager := NewPolicyManager()
	rule1 := NewHasScopeRule("auth")
	rule2 := NewHasScopeRule("premium")

	policy1 := NewPolicy("GET", "/api/data", "AND", []Rule{rule1})
	policy2 := NewPolicy("GET", "/api/data", "OR", []Rule{rule2})
	policy3 := NewPolicy("POST", "/api/data", "AND", []Rule{rule1})

	manager.AddPolicy(policy1)
	manager.AddPolicy(policy2)
	manager.AddPolicy(policy3)

	policies := manager.GetPoliciesForRoute("/api/data", "GET")

	assert.Len(t, policies, 2)
}

// TestManager_Clear removes all policies
func TestManager_Clear(t *testing.T) {
	manager := NewPolicyManager()
	rule := NewHasScopeRule("auth")

	manager.AddPolicy(NewPolicy("GET", "/api/data", "AND", []Rule{rule}))
	manager.AddPolicy(NewPolicy("POST", "/api/users", "AND", []Rule{rule}))

	assert.Equal(t, 2, len(manager.policies))

	manager.Clear()

	assert.Equal(t, 0, len(manager.policies))
}

// TestManager_HasPolicy checks if policy exists for route
func TestManager_HasPolicy(t *testing.T) {
	manager := NewPolicyManager()
	rule := NewHasScopeRule("auth")

	policy := NewPolicy("GET", "/api/data", "AND", []Rule{rule})
	manager.AddPolicy(policy)

	assert.True(t, manager.HasPolicy("/api/data", "GET"))
	assert.False(t, manager.HasPolicy("/api/data", "POST"))
	assert.False(t, manager.HasPolicy("/api/users", "GET"))
}

// TestManager_LoadFromJSON loads policies from JSON
func TestManager_LoadFromJSON(t *testing.T) {
	configJSON := `[
		{
			"path": "/api/data",
			"method": "GET",
			"logic": "AND",
			"rules": [{"type": "has_scope", "scope": "read:data"}]
		}
	]`

	manager := NewPolicyManager()
	err := manager.LoadFromJSON([]byte(configJSON))

	require.NoError(t, err)
	assert.Equal(t, 1, len(manager.policies))
	assert.True(t, manager.HasPolicy("/api/data", "GET"))
}

// TestManager_LoadFromJSON_InvalidConfig returns error
func TestManager_LoadFromJSON_InvalidConfig(t *testing.T) {
	configJSON := `invalid`

	manager := NewPolicyManager()
	err := manager.LoadFromJSON([]byte(configJSON))

	assert.Error(t, err)
}

// TestManager_LoadFromJSON_InvalidPolicy returns error
func TestManager_LoadFromJSON_InvalidPolicy(t *testing.T) {
	configJSON := `[{"path": "/api"}]`

	manager := NewPolicyManager()
	err := manager.LoadFromJSON([]byte(configJSON))

	assert.Error(t, err)
}

// TestManager_GetPoliciesCount returns policy count
func TestManager_GetPoliciesCount(t *testing.T) {
	manager := NewPolicyManager()
	rule := NewHasScopeRule("auth")

	assert.Equal(t, 0, manager.GetPoliciesCount())

	manager.AddPolicy(NewPolicy("GET", "/api/data", "AND", []Rule{rule}))
	assert.Equal(t, 1, manager.GetPoliciesCount())

	manager.AddPolicy(NewPolicy("POST", "/api/users", "AND", []Rule{rule}))
	assert.Equal(t, 2, manager.GetPoliciesCount())
}

// TestManager_GetAllPolicies returns all policies
func TestManager_GetAllPolicies(t *testing.T) {
	manager := NewPolicyManager()
	rule := NewHasScopeRule("auth")

	policy1 := NewPolicy("GET", "/api/data", "AND", []Rule{rule})
	policy2 := NewPolicy("POST", "/api/users", "AND", []Rule{rule})

	manager.AddPolicy(policy1)
	manager.AddPolicy(policy2)

	allPolicies := manager.GetAllPolicies()

	assert.Len(t, allPolicies, 2)
}

// TestManager_ReloadPolicies replaces all policies
func TestManager_ReloadPolicies(t *testing.T) {
	manager := NewPolicyManager()
	rule1 := NewHasScopeRule("auth")

	manager.AddPolicy(NewPolicy("GET", "/api/data", "AND", []Rule{rule1}))
	assert.Equal(t, 1, manager.GetPoliciesCount())

	// Create new policies
	rule2 := NewHasScopeRule("admin")
	newPolicies := []*Policy{
		NewPolicy("GET", "/api/admin", "AND", []Rule{rule2}),
		NewPolicy("DELETE", "/api/users", "AND", []Rule{rule2}),
	}

	manager.ReloadPolicies(newPolicies)

	assert.Equal(t, 2, manager.GetPoliciesCount())
	assert.True(t, manager.HasPolicy("/api/admin", "GET"))
	assert.False(t, manager.HasPolicy("/api/data", "GET"))
}
