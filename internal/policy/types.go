package policy

import (
	"context"
	"math/big"
	"strconv"
	"strings"

	"github.com/yourusername/gatekeeper/internal/auth"
	"github.com/yourusername/gatekeeper/internal/chain"
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
	// cache and provider will be set by manager
	cache    CacheProvider
	provider BlockchainProvider
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

// Evaluate checks ERC20 balance (requires provider and cache to be set)
func (r *ERC20MinBalanceRule) Evaluate(ctx context.Context, address string, claims *auth.Claims) (bool, error) {
	// If no provider configured, evaluate to false (fail-closed)
	if r.provider == nil {
		return false, nil
	}

	// Generate cache key
	chainIDStr := strconv.FormatUint(r.ChainID, 10)
	cacheKey := chain.CacheKey("erc20_balance", chainIDStr, r.ContractAddress, address)

	// Try to get from cache first
	if r.cache != nil {
		if cachedResult, ok := r.cache.Get(cacheKey); ok {
			if hasBalance, ok := cachedResult.(bool); ok {
				return hasBalance, nil
			}
		}
	}

	// Not in cache, call provider
	calldata := encodeERC20BalanceOfCall(r.ContractAddress, address)
	response, err := r.provider.Call(ctx, "eth_call", []interface{}{
		map[string]interface{}{"to": r.ContractAddress, "data": calldata},
		"latest",
	})
	if err != nil {
		// Fail closed on RPC error
		return false, nil
	}

	// Parse JSON-RPC response
	resultHex, err := parseJSONRPCResponse(response)
	if err != nil {
		return false, nil
	}

	// Decode the balance from hex
	balance, err := decodeUint256(resultHex)
	if err != nil {
		return false, nil
	}

	// Compare with minimum balance
	hasBalance := balance.Cmp(r.MinimumBalance) >= 0

	// Cache the result
	if r.cache != nil {
		r.cache.Set(cacheKey, hasBalance)
	}

	return hasBalance, nil
}

// SetProvider sets the blockchain provider for RPC calls
func (r *ERC20MinBalanceRule) SetProvider(provider BlockchainProvider) {
	r.provider = provider
}

// SetCache sets the cache for storing results
func (r *ERC20MinBalanceRule) SetCache(cache CacheProvider) {
	r.cache = cache
}

// ERC721OwnerRule checks if user owns a specific NFT
type ERC721OwnerRule struct {
	ContractAddress string
	TokenID         *big.Int
	ChainID         uint64
	// cache and provider will be set by manager
	cache    CacheProvider
	provider BlockchainProvider
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

// Evaluate checks NFT ownership (requires provider and cache to be set)
func (r *ERC721OwnerRule) Evaluate(ctx context.Context, address string, claims *auth.Claims) (bool, error) {
	// If no provider configured, evaluate to false (fail-closed)
	if r.provider == nil {
		return false, nil
	}

	// Generate cache key
	chainIDStr := strconv.FormatUint(r.ChainID, 10)
	tokenIDStr := r.TokenID.String()
	cacheKey := chain.CacheKey("erc721_owner", chainIDStr, r.ContractAddress, tokenIDStr)

	// Try to get from cache first
	if r.cache != nil {
		if cachedResult, ok := r.cache.Get(cacheKey); ok {
			if isOwner, ok := cachedResult.(bool); ok {
				return isOwner, nil
			}
		}
	}

	// Not in cache, call provider
	calldata := encodeERC721OwnerOfCall(r.ContractAddress, r.TokenID)
	response, err := r.provider.Call(ctx, "eth_call", []interface{}{
		map[string]interface{}{"to": r.ContractAddress, "data": calldata},
		"latest",
	})
	if err != nil {
		// Fail closed on RPC error
		return false, nil
	}

	// Parse JSON-RPC response
	resultHex, err := parseJSONRPCResponse(response)
	if err != nil {
		return false, nil
	}

	// Decode the owner address from hex
	ownerAddress, err := decodeAddress(resultHex)
	if err != nil {
		return false, nil
	}

	// Compare addresses (case-insensitive)
	isOwner := normalizeAddress(ownerAddress) == normalizeAddress(address)

	// Cache the result
	if r.cache != nil {
		r.cache.Set(cacheKey, isOwner)
	}

	return isOwner, nil
}

// SetProvider sets the blockchain provider for RPC calls
func (r *ERC721OwnerRule) SetProvider(provider BlockchainProvider) {
	r.provider = provider
}

// SetCache sets the cache for storing results
func (r *ERC721OwnerRule) SetCache(cache CacheProvider) {
	r.cache = cache
}

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

// normalizeAddress converts address to lowercase for comparison
// TODO: In production, use Ethereum address checksum validation
func normalizeAddress(address string) string {
	// Simple lowercase normalization for case-insensitive comparison
	// In production, should validate Ethereum checksum
	lowerAddr := strings.ToLower(address)
	// Remove 0x prefix if present for consistent comparison
	lowerAddr = strings.TrimPrefix(lowerAddr, "0x")
	return lowerAddr
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
