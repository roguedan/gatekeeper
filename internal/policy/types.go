package policy

import (
	"context"
	"strings"

	"github.com/yourusername/gatekeeper/internal/auth"
)

// RuleType represents the type of policy rule
type RuleType string

const (
	HasScopeRuleType         RuleType = "has_scope"
	InAllowlistRuleType      RuleType = "in_allowlist"
	ERC20MinBalanceRuleType  RuleType = "erc20_min_balance"
	ERC721OwnerRuleType      RuleType = "erc721_owner"
)

// Rule is the interface for all policy rules
type Rule interface {
	// Evaluate returns true if the rule passes for the given address and claims
	Evaluate(ctx context.Context, address string, claims *auth.Claims) (bool, error)
	// Type returns the rule type
	Type() RuleType
}

// Policy represents an access control policy for a route
type Policy struct {
	Path   string // Route pattern (e.g., "/api/data", "/api/*")
	Method string // HTTP method (GET, POST, etc.)
	Logic  string // "AND" or "OR" - how to combine rules
	Rules  []Rule // List of rules to evaluate
}

// NewPolicy creates a new policy with the given parameters
func NewPolicy(method, path, logic string, rules []Rule) *Policy {
	return &Policy{
		Method: method,
		Path:   path,
		Logic:  logic,
		Rules:  rules,
	}
}

// HasScopeRule checks if user has a specific scope
type HasScopeRule struct {
	Scope string
}

// NewHasScopeRule creates a new scope rule
func NewHasScopeRule(scope string) *HasScopeRule {
	return &HasScopeRule{Scope: scope}
}

// Type returns the rule type
func (r *HasScopeRule) Type() RuleType {
	return HasScopeRuleType
}

// Evaluate checks if the user has the required scope
func (r *HasScopeRule) Evaluate(ctx context.Context, address string, claims *auth.Claims) (bool, error) {
	if claims == nil {
		return false, nil
	}

	for _, scope := range claims.Scopes {
		if scope == r.Scope {
			return true, nil
		}
	}

	return false, nil
}

// InAllowlistRule checks if user address is in an allowlist
type InAllowlistRule struct {
	Addresses []string
}

// NewInAllowlistRule creates a new allowlist rule
func NewInAllowlistRule(addresses []string) *InAllowlistRule {
	return &InAllowlistRule{Addresses: addresses}
}

// Type returns the rule type
func (r *InAllowlistRule) Type() RuleType {
	return InAllowlistRuleType
}

// Evaluate checks if address is in the allowlist (case-insensitive)
func (r *InAllowlistRule) Evaluate(ctx context.Context, address string, claims *auth.Claims) (bool, error) {
	// Normalize address to lowercase for comparison
	normalizedAddress := strings.ToLower(address)

	for _, allowed := range r.Addresses {
		if strings.ToLower(allowed) == normalizedAddress {
			return true, nil
		}
	}

	return false, nil
}

// Note: ERC20MinBalanceRule and ERC721OwnerRule implementations
// are now in separate files (erc20_rule.go and erc721_rule.go)

// BlockchainProvider interface for RPC calls
type BlockchainProvider interface {
	Call(ctx context.Context, method string, params []interface{}) ([]byte, error)
	HealthCheck(ctx context.Context) bool
}

// CacheProvider interface for caching
type CacheProvider interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	GetOrSet(key string, fn func() interface{}) interface{}
}

// Evaluate evaluates the policy for the given address and claims
func (p *Policy) Evaluate(ctx context.Context, address string, claims *auth.Claims) (bool, error) {
	if p.Logic == "AND" {
		return p.evaluateAND(ctx, address, claims)
	} else if p.Logic == "OR" {
		return p.evaluateOR(ctx, address, claims)
	}
	return false, nil
}

// evaluateAND requires all rules to pass
func (p *Policy) evaluateAND(ctx context.Context, address string, claims *auth.Claims) (bool, error) {
	for _, rule := range p.Rules {
		result, err := rule.Evaluate(ctx, address, claims)
		if err != nil {
			return false, err
		}
		// Short-circuit on first failure
		if !result {
			return false, nil
		}
	}
	return true, nil
}

// evaluateOR requires any rule to pass
func (p *Policy) evaluateOR(ctx context.Context, address string, claims *auth.Claims) (bool, error) {
	for _, rule := range p.Rules {
		result, err := rule.Evaluate(ctx, address, claims)
		if err != nil {
			return false, err
		}
		// Short-circuit on first success
		if result {
			return true, nil
		}
	}
	return false, nil
}
