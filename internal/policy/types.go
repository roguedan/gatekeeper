package policy

import (
	"context"
	"math/big"

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
	normalizedAddress := normalizeAddress(address)

	for _, allowed := range r.Addresses {
		if normalizeAddress(allowed) == normalizedAddress {
			return true, nil
		}
	}

	return false, nil
}

// ERC20MinBalanceRule checks if user has minimum ERC20 token balance
type ERC20MinBalanceRule struct {
	ContractAddress string
	MinimumBalance  *big.Int
	ChainID         uint64
}

// NewERC20MinBalanceRule creates a new ERC20 balance rule
func NewERC20MinBalanceRule(contractAddress string, minimumBalance *big.Int, chainID uint64) *ERC20MinBalanceRule {
	return &ERC20MinBalanceRule{
		ContractAddress: contractAddress,
		MinimumBalance:  minimumBalance,
		ChainID:         chainID,
	}
}

// Type returns the rule type
func (r *ERC20MinBalanceRule) Type() RuleType {
	return ERC20MinBalanceRuleType
}

// Evaluate checks ERC20 balance (implemented later with blockchain integration)
func (r *ERC20MinBalanceRule) Evaluate(ctx context.Context, address string, claims *auth.Claims) (bool, error) {
	// TODO: Implement with blockchain integration
	return false, nil
}

// ERC721OwnerRule checks if user owns a specific NFT
type ERC721OwnerRule struct {
	ContractAddress string
	TokenID         *big.Int
	ChainID         uint64
}

// NewERC721OwnerRule creates a new NFT ownership rule
func NewERC721OwnerRule(contractAddress string, tokenID *big.Int, chainID uint64) *ERC721OwnerRule {
	return &ERC721OwnerRule{
		ContractAddress: contractAddress,
		TokenID:         tokenID,
		ChainID:         chainID,
	}
}

// Type returns the rule type
func (r *ERC721OwnerRule) Type() RuleType {
	return ERC721OwnerRuleType
}

// Evaluate checks NFT ownership (implemented later with blockchain integration)
func (r *ERC721OwnerRule) Evaluate(ctx context.Context, address string, claims *auth.Claims) (bool, error) {
	// TODO: Implement with blockchain integration
	return false, nil
}

// normalizeAddress converts address to lowercase for comparison
// TODO: In production, use Ethereum address checksum validation
func normalizeAddress(address string) string {
	// For now, simple lowercase normalization
	// In production, should validate Ethereum checksum
	return address
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
