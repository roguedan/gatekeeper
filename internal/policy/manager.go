package policy

import (
	"sync"

	"go.uber.org/zap"
)

// PolicyManager manages a collection of policies and provides route matching
type PolicyManager struct {
	policies []*Policy
	mu       sync.RWMutex
	loader   *PolicyLoader
	provider BlockchainProvider // For blockchain rules
	cache    CacheProvider      // For caching blockchain results
	logger   *zap.Logger
}

// NewPolicyManager creates a new policy manager
// provider and cache can be nil if blockchain rules are not used
func NewPolicyManager(provider BlockchainProvider, cache CacheProvider) *PolicyManager {
	logger, _ := zap.NewProduction()
	return &PolicyManager{
		policies: make([]*Policy, 0),
		loader:   NewPolicyLoader(),
		provider: provider,
		cache:    cache,
		logger:   logger,
	}
}

// AddPolicy adds a policy to the manager
// Automatically wires up blockchain rules with provider and cache
func (pm *PolicyManager) AddPolicy(policy *Policy) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Wire up blockchain rules with provider and cache
	pm.wireBlockchainRules(policy)

	pm.policies = append(pm.policies, policy)
}

// wireBlockchainRules sets provider and cache on blockchain rules
func (pm *PolicyManager) wireBlockchainRules(policy *Policy) {
	if policy == nil || policy.Rules == nil {
		return
	}

	for _, rule := range policy.Rules {
		switch r := rule.(type) {
		case *ERC20MinBalanceRule:
			r.SetProvider(pm.provider)
			r.SetCache(pm.cache)
			if pm.logger != nil {
				r.SetLogger(pm.logger)
			}
		case *ERC721OwnerRule:
			r.SetProvider(pm.provider)
			r.SetCache(pm.cache)
			if pm.logger != nil {
				r.SetLogger(pm.logger)
			}
		}
	}
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

	// Wire up blockchain rules for all loaded policies
	for _, policy := range policies {
		pm.wireBlockchainRules(policy)
	}

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

	// Wire up blockchain rules for all policies
	for _, policy := range policies {
		pm.wireBlockchainRules(policy)
	}

	pm.policies = make([]*Policy, len(policies))
	copy(pm.policies, policies)
}

// SetLogger sets the logger for the policy manager
func (pm *PolicyManager) SetLogger(logger *zap.Logger) {
	pm.logger = logger
}
