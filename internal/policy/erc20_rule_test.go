package policy

import (
	"context"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/gatekeeper/internal/chain"
)

// TestERC20MinBalanceRule_Validate validates rule parameters
func TestERC20MinBalanceRule_Validate(t *testing.T) {
	tests := []struct {
		name          string
		contractAddr  string
		minBalance    *big.Int
		chainID       uint64
		expectedError bool
	}{
		{
			name:          "valid rule",
			contractAddr:  "0x1234567890123456789012345678901234567890",
			minBalance:    big.NewInt(1000),
			chainID:       1,
			expectedError: false,
		},
		{
			name:          "invalid contract address - no 0x prefix",
			contractAddr:  "1234567890123456789012345678901234567890",
			minBalance:    big.NewInt(1000),
			chainID:       1,
			expectedError: true,
		},
		{
			name:          "invalid contract address - too short",
			contractAddr:  "0x1234",
			minBalance:    big.NewInt(1000),
			chainID:       1,
			expectedError: true,
		},
		{
			name:          "nil minimum balance",
			contractAddr:  "0x1234567890123456789012345678901234567890",
			minBalance:    nil,
			chainID:       1,
			expectedError: true,
		},
		{
			name:          "negative minimum balance",
			contractAddr:  "0x1234567890123456789012345678901234567890",
			minBalance:    big.NewInt(-1000),
			chainID:       1,
			expectedError: true,
		},
		{
			name:          "zero chain ID",
			contractAddr:  "0x1234567890123456789012345678901234567890",
			minBalance:    big.NewInt(1000),
			chainID:       0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewERC20MinBalanceRule(tt.contractAddr, tt.minBalance, tt.chainID)
			err := rule.Validate()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestERC20MinBalanceRule_Evaluate_SufficientBalance passes when balance meets minimum
func TestERC20MinBalanceRule_Evaluate_SufficientBalance(t *testing.T) {
	minBalance := big.NewInt(1000)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	rule := NewERC20MinBalanceRule("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", minBalance, 1)

	// Mock provider that returns sufficient balance
	provider := &MockBlockchainProvider{}
	provider.SetBalance(userAddr, big.NewInt(5000))
	rule.SetProvider(provider)

	// Mock cache
	cache := &MockCache{}
	rule.SetCache(cache)

	result, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	assert.True(t, result)
}

// TestERC20MinBalanceRule_Evaluate_InsufficientBalance fails when balance below minimum
func TestERC20MinBalanceRule_Evaluate_InsufficientBalance(t *testing.T) {
	minBalance := big.NewInt(1000)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	rule := NewERC20MinBalanceRule("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", minBalance, 1)

	// Mock provider that returns insufficient balance
	provider := &MockBlockchainProvider{}
	provider.SetBalance(userAddr, big.NewInt(500))
	rule.SetProvider(provider)

	// Mock cache
	cache := &MockCache{}
	rule.SetCache(cache)

	result, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	assert.False(t, result)
}

// TestERC20MinBalanceRule_Evaluate_ZeroBalance fails when balance is zero
func TestERC20MinBalanceRule_Evaluate_ZeroBalance(t *testing.T) {
	minBalance := big.NewInt(1000)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	rule := NewERC20MinBalanceRule("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", minBalance, 1)

	// Mock provider that returns zero balance
	provider := &MockBlockchainProvider{}
	provider.SetBalance(userAddr, big.NewInt(0))
	rule.SetProvider(provider)

	// Mock cache
	cache := &MockCache{}
	rule.SetCache(cache)

	result, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	assert.False(t, result)
}

// TestERC20MinBalanceRule_Evaluate_NoProvider returns false when no provider
func TestERC20MinBalanceRule_Evaluate_NoProvider(t *testing.T) {
	minBalance := big.NewInt(1000)
	rule := NewERC20MinBalanceRule("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", minBalance, 1)
	// Don't set provider

	result, err := rule.Evaluate(context.Background(), "0x1234567890abcdef1234567890abcdef12345678", nil)

	require.NoError(t, err)
	assert.False(t, result) // Fail-closed
}

// TestERC20MinBalanceRule_Evaluate_InvalidAddress returns false on invalid address
func TestERC20MinBalanceRule_Evaluate_InvalidAddress(t *testing.T) {
	minBalance := big.NewInt(1000)
	rule := NewERC20MinBalanceRule("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", minBalance, 1)

	provider := &MockBlockchainProvider{}
	rule.SetProvider(provider)

	cache := &MockCache{}
	rule.SetCache(cache)

	// Invalid address format
	result, err := rule.Evaluate(context.Background(), "invalid_address", nil)

	require.NoError(t, err)
	assert.False(t, result) // Fail-closed
}

// TestERC20MinBalanceRule_Evaluate_UsesCache returns cached value without RPC call
func TestERC20MinBalanceRule_Evaluate_UsesCache(t *testing.T) {
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	tokenAddr := "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
	minBalance := big.NewInt(1000)
	rule := NewERC20MinBalanceRule(tokenAddr, minBalance, 1)

	// Mock provider
	provider := &MockBlockchainProvider{}
	rule.SetProvider(provider)

	// Mock cache with pre-cached value (use lowercase addresses for cache key)
	cache := &MockCache{}
	cache.data = make(map[string]interface{})
	cacheKey := chain.CacheKey("erc20_balance", "1", strings.ToLower(tokenAddr), strings.ToLower(userAddr))
	cache.data[cacheKey] = true
	rule.SetCache(cache)

	result, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	assert.True(t, result)
	// Verify provider was not called (balances map would be empty)
	assert.Empty(t, provider.balances)
}

// TestERC20MinBalanceRule_Evaluate_CachesResult stores result in cache
func TestERC20MinBalanceRule_Evaluate_CachesResult(t *testing.T) {
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	minBalance := big.NewInt(1000)
	rule := NewERC20MinBalanceRule("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", minBalance, 1)

	// Mock provider
	provider := &MockBlockchainProvider{}
	provider.SetBalance(userAddr, big.NewInt(5000))
	rule.SetProvider(provider)

	// Mock cache
	cache := &MockCache{}
	rule.SetCache(cache)

	_, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	// Verify result was cached
	assert.True(t, cache.has("erc20_balance:1:0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48:"+userAddr) ||
		cache.has("erc20_balance:1:0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48:"+userAddr))
}

// TestERC20MinBalanceRule_Evaluate_LargeBalance supports large token amounts
func TestERC20MinBalanceRule_Evaluate_LargeBalance(t *testing.T) {
	// 1 million tokens with 18 decimals = 1e24
	largeBalance := new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	rule := NewERC20MinBalanceRule("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", largeBalance, 1)

	// Mock provider with even larger balance
	provider := &MockBlockchainProvider{}
	hugeBalance := new(big.Int).Exp(big.NewInt(10), big.NewInt(30), nil)
	provider.SetBalance(userAddr, hugeBalance)
	rule.SetProvider(provider)

	cache := &MockCache{}
	rule.SetCache(cache)

	result, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	assert.True(t, result)
}

// TestERC20MinBalanceRule_Evaluate_MultiChain supports different chain IDs
func TestERC20MinBalanceRule_Evaluate_MultiChain(t *testing.T) {
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	minBalance := big.NewInt(1000)

	tests := []struct {
		name    string
		chainID uint64
	}{
		{"Ethereum Mainnet", 1},
		{"Goerli Testnet", 5},
		{"Sepolia Testnet", 11155111},
		{"Polygon", 137},
		{"Arbitrum", 42161},
		{"Optimism", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewERC20MinBalanceRule("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", minBalance, tt.chainID)
			assert.Equal(t, tt.chainID, rule.ChainID)

			provider := &MockBlockchainProvider{}
			provider.SetBalance(userAddr, big.NewInt(5000))
			rule.SetProvider(provider)

			cache := &MockCache{}
			rule.SetCache(cache)

			result, err := rule.Evaluate(context.Background(), userAddr, nil)
			require.NoError(t, err)
			assert.True(t, result)
		})
	}
}

// TestERC20MinBalanceRule_Type returns correct rule type
func TestERC20MinBalanceRule_Type(t *testing.T) {
	rule := NewERC20MinBalanceRule("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", big.NewInt(1000), 1)
	assert.Equal(t, ERC20MinBalanceRuleType, rule.Type())
}

// TestERC20MinBalanceRule_Evaluate_ExactBalance passes when balance equals minimum
func TestERC20MinBalanceRule_Evaluate_ExactBalance(t *testing.T) {
	minBalance := big.NewInt(1000)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	rule := NewERC20MinBalanceRule("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", minBalance, 1)

	// Mock provider that returns exact balance
	provider := &MockBlockchainProvider{}
	provider.SetBalance(userAddr, big.NewInt(1000))
	rule.SetProvider(provider)

	cache := &MockCache{}
	rule.SetCache(cache)

	result, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	assert.True(t, result) // Should pass with equal balance
}
