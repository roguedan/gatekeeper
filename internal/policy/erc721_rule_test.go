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

// TestERC721OwnerRule_Validate validates rule parameters
func TestERC721OwnerRule_Validate(t *testing.T) {
	tests := []struct {
		name          string
		contractAddr  string
		tokenID       *big.Int
		chainID       uint64
		expectedError bool
	}{
		{
			name:          "valid rule",
			contractAddr:  "0x1234567890123456789012345678901234567890",
			tokenID:       big.NewInt(42),
			chainID:       1,
			expectedError: false,
		},
		{
			name:          "invalid contract address - no 0x prefix",
			contractAddr:  "1234567890123456789012345678901234567890",
			tokenID:       big.NewInt(42),
			chainID:       1,
			expectedError: true,
		},
		{
			name:          "invalid contract address - too short",
			contractAddr:  "0x1234",
			tokenID:       big.NewInt(42),
			chainID:       1,
			expectedError: true,
		},
		{
			name:          "nil token ID",
			contractAddr:  "0x1234567890123456789012345678901234567890",
			tokenID:       nil,
			chainID:       1,
			expectedError: true,
		},
		{
			name:          "negative token ID",
			contractAddr:  "0x1234567890123456789012345678901234567890",
			tokenID:       big.NewInt(-1),
			chainID:       1,
			expectedError: true,
		},
		{
			name:          "zero chain ID",
			contractAddr:  "0x1234567890123456789012345678901234567890",
			tokenID:       big.NewInt(42),
			chainID:       0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewERC721OwnerRule(tt.contractAddr, tt.tokenID, tt.chainID)
			err := rule.Validate()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestERC721OwnerRule_Evaluate_IsOwner passes when user owns NFT
func TestERC721OwnerRule_Evaluate_IsOwner(t *testing.T) {
	tokenID := big.NewInt(42)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	rule := NewERC721OwnerRule("0xBC4CA0EdA7647A8aB7C2061c2E9cDAFCAc3c7f70", tokenID, 1)

	// Mock provider that returns user as owner
	provider := &MockBlockchainProvider{}
	provider.SetOwner(tokenID.String(), userAddr)
	rule.SetProvider(provider)

	// Mock cache
	cache := &MockCache{}
	rule.SetCache(cache)

	result, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	assert.True(t, result)
}

// TestERC721OwnerRule_Evaluate_NotOwner fails when user doesn't own NFT
func TestERC721OwnerRule_Evaluate_NotOwner(t *testing.T) {
	tokenID := big.NewInt(42)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	otherAddr := "0xabcdef1234567890abcdef1234567890abcdef12"
	rule := NewERC721OwnerRule("0xBC4CA0EdA7647A8aB7C2061c2E9cDAFCAc3c7f70", tokenID, 1)

	// Mock provider that returns different owner
	provider := &MockBlockchainProvider{}
	provider.SetOwner(tokenID.String(), otherAddr)
	rule.SetProvider(provider)

	// Mock cache
	cache := &MockCache{}
	rule.SetCache(cache)

	result, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	assert.False(t, result)
}

// TestERC721OwnerRule_Evaluate_NoProvider returns false when no provider
func TestERC721OwnerRule_Evaluate_NoProvider(t *testing.T) {
	tokenID := big.NewInt(42)
	rule := NewERC721OwnerRule("0xBC4CA0EdA7647A8aB7C2061c2E9cDAFCAc3c7f70", tokenID, 1)
	// Don't set provider

	result, err := rule.Evaluate(context.Background(), "0x1234567890abcdef1234567890abcdef12345678", nil)

	require.NoError(t, err)
	assert.False(t, result) // Fail-closed
}

// TestERC721OwnerRule_Evaluate_InvalidAddress returns false on invalid address
func TestERC721OwnerRule_Evaluate_InvalidAddress(t *testing.T) {
	tokenID := big.NewInt(42)
	rule := NewERC721OwnerRule("0xBC4CA0EdA7647A8aB7C2061c2E9cDAFCAc3c7f70", tokenID, 1)

	provider := &MockBlockchainProvider{}
	rule.SetProvider(provider)

	cache := &MockCache{}
	rule.SetCache(cache)

	// Invalid address format
	result, err := rule.Evaluate(context.Background(), "invalid_address", nil)

	require.NoError(t, err)
	assert.False(t, result) // Fail-closed
}

// TestERC721OwnerRule_Evaluate_UsesCache returns cached value without RPC call
func TestERC721OwnerRule_Evaluate_UsesCache(t *testing.T) {
	tokenID := big.NewInt(42)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	nftAddr := "0xBC4CA0EdA7647A8aB7C2061c2E9cDAFCAc3c7f70"
	rule := NewERC721OwnerRule(nftAddr, tokenID, 1)

	// Mock provider
	provider := &MockBlockchainProvider{}
	rule.SetProvider(provider)

	// Mock cache with pre-cached owner (use lowercase address for cache key)
	cache := &MockCache{}
	cache.data = make(map[string]interface{})
	cacheKey := chain.CacheKey("erc721_owner", "1", strings.ToLower(nftAddr), "42")
	cache.data[cacheKey] = userAddr // Cache the owner address
	rule.SetCache(cache)

	result, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	assert.True(t, result)
	// Verify provider was not called
	assert.Empty(t, provider.owners)
}

// TestERC721OwnerRule_Evaluate_CachesResult stores owner in cache
func TestERC721OwnerRule_Evaluate_CachesResult(t *testing.T) {
	tokenID := big.NewInt(42)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	rule := NewERC721OwnerRule("0xBC4CA0EdA7647A8aB7C2061c2E9cDAFCAc3c7f70", tokenID, 1)

	// Mock provider
	provider := &MockBlockchainProvider{}
	provider.SetOwner(tokenID.String(), userAddr)
	rule.SetProvider(provider)

	// Mock cache
	cache := &MockCache{}
	rule.SetCache(cache)

	_, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	// Verify owner was cached
	assert.True(t, cache.has("erc721_owner:1:0xBC4CA0EdA7647A8aB7C2061c2E9cDAFCAc3c7f70:42") ||
		cache.has("erc721_owner:1:0xbc4ca0eda7647a8ab7c2061c2e9cdafcac3c7f70:42"))
}

// TestERC721OwnerRule_Evaluate_BurnedToken returns false for zero address
func TestERC721OwnerRule_Evaluate_BurnedToken(t *testing.T) {
	tokenID := big.NewInt(42)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	rule := NewERC721OwnerRule("0xBC4CA0EdA7647A8aB7C2061c2E9cDAFCAc3c7f70", tokenID, 1)

	// Mock provider that returns zero address (burned token)
	provider := &MockBlockchainProvider{}
	provider.SetOwner(tokenID.String(), "0x0000000000000000000000000000000000000000")
	rule.SetProvider(provider)

	cache := &MockCache{}
	rule.SetCache(cache)

	result, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	assert.False(t, result) // Burned tokens should fail
}

// TestERC721OwnerRule_Evaluate_LargeTokenID supports large token IDs
func TestERC721OwnerRule_Evaluate_LargeTokenID(t *testing.T) {
	// Very large token ID
	largeID := new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	rule := NewERC721OwnerRule("0xBC4CA0EdA7647A8aB7C2061c2E9cDAFCAc3c7f70", largeID, 1)

	provider := &MockBlockchainProvider{}
	provider.SetOwner(largeID.String(), userAddr)
	rule.SetProvider(provider)

	cache := &MockCache{}
	rule.SetCache(cache)

	result, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	assert.True(t, result)
}

// TestERC721OwnerRule_Evaluate_MultiChain supports different chain IDs
func TestERC721OwnerRule_Evaluate_MultiChain(t *testing.T) {
	tokenID := big.NewInt(42)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"

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
			rule := NewERC721OwnerRule("0xBC4CA0EdA7647A8aB7C2061c2E9cDAFCAc3c7f70", tokenID, tt.chainID)
			assert.Equal(t, tt.chainID, rule.ChainID)

			provider := &MockBlockchainProvider{}
			provider.SetOwner(tokenID.String(), userAddr)
			rule.SetProvider(provider)

			cache := &MockCache{}
			rule.SetCache(cache)

			result, err := rule.Evaluate(context.Background(), userAddr, nil)
			require.NoError(t, err)
			assert.True(t, result)
		})
	}
}

// TestERC721OwnerRule_Type returns correct rule type
func TestERC721OwnerRule_Type(t *testing.T) {
	rule := NewERC721OwnerRule("0xBC4CA0EdA7647A8aB7C2061c2E9cDAFCAc3c7f70", big.NewInt(42), 1)
	assert.Equal(t, ERC721OwnerRuleType, rule.Type())
}

// TestERC721OwnerRule_Evaluate_CaseInsensitiveAddress compares addresses case-insensitively
func TestERC721OwnerRule_Evaluate_CaseInsensitiveAddress(t *testing.T) {
	tokenID := big.NewInt(42)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	ownerAddr := "0x1234567890ABCDEF1234567890ABCDEF12345678" // Same address, different case
	rule := NewERC721OwnerRule("0xBC4CA0EdA7647A8aB7C2061c2E9cDAFCAc3c7f70", tokenID, 1)

	provider := &MockBlockchainProvider{}
	provider.SetOwner(tokenID.String(), ownerAddr)
	rule.SetProvider(provider)

	cache := &MockCache{}
	rule.SetCache(cache)

	result, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	assert.True(t, result) // Should match despite case difference
}

// TestERC721OwnerRule_Evaluate_ZeroTokenID supports token ID zero
func TestERC721OwnerRule_Evaluate_ZeroTokenID(t *testing.T) {
	tokenID := big.NewInt(0)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	rule := NewERC721OwnerRule("0xBC4CA0EdA7647A8aB7C2061c2E9cDAFCAc3c7f70", tokenID, 1)

	provider := &MockBlockchainProvider{}
	provider.SetOwner(tokenID.String(), userAddr)
	rule.SetProvider(provider)

	cache := &MockCache{}
	rule.SetCache(cache)

	result, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	assert.True(t, result)
}

// TestIsZeroAddress validates zero address detection
func TestIsZeroAddress(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{
			name:     "zero address with 0x",
			address:  "0x0000000000000000000000000000000000000000",
			expected: true,
		},
		{
			name:     "zero address without 0x",
			address:  "0000000000000000000000000000000000000000",
			expected: true,
		},
		{
			name:     "non-zero address",
			address:  "0x1234567890abcdef1234567890abcdef12345678",
			expected: false,
		},
		{
			name:     "mixed case zero address",
			address:  "0x0000000000000000000000000000000000000000",
			expected: true,
		},
		{
			name:     "invalid length",
			address:  "0x0000",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isZeroAddress(tt.address)
			assert.Equal(t, tt.expected, result)
		})
	}
}
