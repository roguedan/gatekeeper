package policy

import (
	"sync"
)

// PolicyManager manages a collection of policies and provides route matching
type PolicyManager struct {
	policies []*Policy
	mu       sync.RWMutex
	loader   *PolicyLoader
}

// NewPolicyManager creates a new policy manager
func NewPolicyManager() *PolicyManager {
	return &PolicyManager{
		policies: make([]*Policy, 0),
		loader:   NewPolicyLoader(),
	}
}

// AddPolicy adds a policy to the manager
func (pm *PolicyManager) AddPolicy(policy *Policy) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.policies = append(pm.policies, policy)
}

// GetPoliciesForRoute returns all policies matching the given route and method
func (pm *PolicyManager) GetPoliciesForRoute(path string, method string) []*Policy {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var matching []*Policy
	for _, policy := range pm.policies {
		if policy.Path == path && policy.Method == method {
			matching = append(matching, policy)
		}
	}
	return matching
}

// HasPolicy checks if any policy exists for the given route and method
func (pm *PolicyManager) HasPolicy(path string, method string) bool {
	return len(pm.GetPoliciesForRoute(path, method)) > 0
}

// Clear removes all policies from the manager
func (pm *PolicyManager) Clear() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.policies = make([]*Policy, 0)
}

// LoadFromJSON loads policies from JSON configuration
func (pm *PolicyManager) LoadFromJSON(data []byte) error {
	policies, err := pm.loader.LoadFromJSON(data)
	if err != nil {
		return err
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.policies = policies
	return nil
}

// GetPoliciesCount returns the number of policies
func (pm *PolicyManager) GetPoliciesCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.policies)
}

// GetAllPolicies returns a copy of all policies
func (pm *PolicyManager) GetAllPolicies() []*Policy {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]*Policy, len(pm.policies))
	copy(result, pm.policies)
	return result
}

// ReloadPolicies replaces all policies with new ones
func (pm *PolicyManager) ReloadPolicies(policies []*Policy) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.policies = make([]*Policy, len(policies))
	copy(pm.policies, policies)
}
