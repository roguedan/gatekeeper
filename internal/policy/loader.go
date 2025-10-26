package policy

import (
	"encoding/json"
	"fmt"
	"math/big"
)

// PolicyLoader handles loading and validating policies from JSON
type PolicyLoader struct{}

// NewPolicyLoader creates a new policy loader
func NewPolicyLoader() *PolicyLoader {
	return &PolicyLoader{}
}

// policyConfig represents the JSON structure for a policy
type policyConfig struct {
	Path   string          `json:"path"`
	Method string          `json:"method"`
	Logic  string          `json:"logic"`
	Rules  []json.RawMessage `json:"rules"`
}

// ruleConfig represents the base structure for a rule
type ruleConfig struct {
	Type string `json:"type"`
}

// LoadFromJSON parses policies from JSON bytes
func (l *PolicyLoader) LoadFromJSON(data []byte) ([]*Policy, error) {
	var configs []policyConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	var policies []*Policy
	for i, config := range configs {
		policy, err := l.loadPolicy(config, i)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

// loadPolicy validates and loads a single policy configuration
func (l *PolicyLoader) loadPolicy(config policyConfig, index int) (*Policy, error) {
	// Validate required fields
	if config.Path == "" {
		return nil, fmt.Errorf("policy %d: path is required", index)
	}
	if config.Method == "" {
		return nil, fmt.Errorf("policy %d: method is required", index)
	}
	if config.Logic == "" {
		return nil, fmt.Errorf("policy %d: logic is required", index)
	}

	// Validate logic is AND or OR
	if config.Logic != "AND" && config.Logic != "OR" {
		return nil, fmt.Errorf("policy %d: logic must be 'AND' or 'OR', got '%s'", index, config.Logic)
	}

	// Validate rules exist
	if len(config.Rules) == 0 {
		return nil, fmt.Errorf("policy %d: rules must contain at least one rule", index)
	}

	// Load rules
	rules, err := l.loadRules(config.Rules, index)
	if err != nil {
		return nil, err
	}

	return NewPolicy(config.Method, config.Path, config.Logic, rules), nil
}

// loadRules parses and validates rules
func (l *PolicyLoader) loadRules(rawRules []json.RawMessage, policyIndex int) ([]Rule, error) {
	var rules []Rule

	for i, rawRule := range rawRules {
		rule, err := l.loadRule(rawRule, policyIndex, i)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// loadRule parses and validates a single rule
func (l *PolicyLoader) loadRule(rawRule json.RawMessage, policyIndex, ruleIndex int) (Rule, error) {
	// First, extract rule type
	var baseConfig ruleConfig
	if err := json.Unmarshal(rawRule, &baseConfig); err != nil {
		return nil, fmt.Errorf("policy %d rule %d: invalid rule format: %w", policyIndex, ruleIndex, err)
	}

	if baseConfig.Type == "" {
		return nil, fmt.Errorf("policy %d rule %d: type is required", policyIndex, ruleIndex)
	}

	// Load rule based on type
	switch baseConfig.Type {
	case "has_scope":
		return l.loadHasScopeRule(rawRule, policyIndex, ruleIndex)
	case "in_allowlist":
		return l.loadInAllowlistRule(rawRule, policyIndex, ruleIndex)
	case "erc20_min_balance":
		return l.loadERC20MinBalanceRule(rawRule, policyIndex, ruleIndex)
	case "erc721_owner":
		return l.loadERC721OwnerRule(rawRule, policyIndex, ruleIndex)
	default:
		return nil, fmt.Errorf("policy %d rule %d: unknown rule type '%s'", policyIndex, ruleIndex, baseConfig.Type)
	}
}

// loadHasScopeRule parses a has_scope rule
func (l *PolicyLoader) loadHasScopeRule(rawRule json.RawMessage, policyIndex, ruleIndex int) (*HasScopeRule, error) {
	type hasScopeConfig struct {
		Type  string `json:"type"`
		Scope string `json:"scope"`
	}

	var config hasScopeConfig
	if err := json.Unmarshal(rawRule, &config); err != nil {
		return nil, fmt.Errorf("policy %d rule %d: invalid has_scope rule: %w", policyIndex, ruleIndex, err)
	}

	if config.Scope == "" {
		return nil, fmt.Errorf("policy %d rule %d: scope is required for has_scope rule", policyIndex, ruleIndex)
	}

	return NewHasScopeRule(config.Scope), nil
}

// loadInAllowlistRule parses an in_allowlist rule
func (l *PolicyLoader) loadInAllowlistRule(rawRule json.RawMessage, policyIndex, ruleIndex int) (*InAllowlistRule, error) {
	type allowlistConfig struct {
		Type      string   `json:"type"`
		Addresses []string `json:"addresses"`
	}

	var config allowlistConfig
	if err := json.Unmarshal(rawRule, &config); err != nil {
		return nil, fmt.Errorf("policy %d rule %d: invalid in_allowlist rule: %w", policyIndex, ruleIndex, err)
	}

	if len(config.Addresses) == 0 {
		return nil, fmt.Errorf("policy %d rule %d: addresses is required for in_allowlist rule", policyIndex, ruleIndex)
	}

	return NewInAllowlistRule(config.Addresses), nil
}

// loadERC20MinBalanceRule parses an erc20_min_balance rule
func (l *PolicyLoader) loadERC20MinBalanceRule(rawRule json.RawMessage, policyIndex, ruleIndex int) (*ERC20MinBalanceRule, error) {
	type erc20Config struct {
		Type             string `json:"type"`
		ContractAddress  string `json:"contract_address"`
		MinimumBalance   string `json:"minimum_balance"`
		ChainID          uint64 `json:"chain_id"`
	}

	var config erc20Config
	if err := json.Unmarshal(rawRule, &config); err != nil {
		return nil, fmt.Errorf("policy %d rule %d: invalid erc20_min_balance rule: %w", policyIndex, ruleIndex, err)
	}

	if config.ContractAddress == "" {
		return nil, fmt.Errorf("policy %d rule %d: contract_address is required for erc20_min_balance rule", policyIndex, ruleIndex)
	}

	if config.MinimumBalance == "" {
		return nil, fmt.Errorf("policy %d rule %d: minimum_balance is required for erc20_min_balance rule", policyIndex, ruleIndex)
	}

	// Parse minimum balance as big.Int
	minimumBalance := new(big.Int)
	if _, ok := minimumBalance.SetString(config.MinimumBalance, 10); !ok {
		return nil, fmt.Errorf("policy %d rule %d: invalid minimum_balance format", policyIndex, ruleIndex)
	}

	if config.ChainID == 0 {
		return nil, fmt.Errorf("policy %d rule %d: chain_id is required for erc20_min_balance rule", policyIndex, ruleIndex)
	}

	return NewERC20MinBalanceRule(config.ContractAddress, minimumBalance, config.ChainID), nil
}

// loadERC721OwnerRule parses an erc721_owner rule
func (l *PolicyLoader) loadERC721OwnerRule(rawRule json.RawMessage, policyIndex, ruleIndex int) (*ERC721OwnerRule, error) {
	type erc721Config struct {
		Type            string `json:"type"`
		ContractAddress string `json:"contract_address"`
		TokenID         string `json:"token_id"`
		ChainID         uint64 `json:"chain_id"`
	}

	var config erc721Config
	if err := json.Unmarshal(rawRule, &config); err != nil {
		return nil, fmt.Errorf("policy %d rule %d: invalid erc721_owner rule: %w", policyIndex, ruleIndex, err)
	}

	if config.ContractAddress == "" {
		return nil, fmt.Errorf("policy %d rule %d: contract_address is required for erc721_owner rule", policyIndex, ruleIndex)
	}

	if config.TokenID == "" {
		return nil, fmt.Errorf("policy %d rule %d: token_id is required for erc721_owner rule", policyIndex, ruleIndex)
	}

	// Parse token ID as big.Int
	tokenID := new(big.Int)
	if _, ok := tokenID.SetString(config.TokenID, 10); !ok {
		return nil, fmt.Errorf("policy %d rule %d: invalid token_id format", policyIndex, ruleIndex)
	}

	if config.ChainID == 0 {
		return nil, fmt.Errorf("policy %d rule %d: chain_id is required for erc721_owner rule", policyIndex, ruleIndex)
	}

	return NewERC721OwnerRule(config.ContractAddress, tokenID, config.ChainID), nil
}
