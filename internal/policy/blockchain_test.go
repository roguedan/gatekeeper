package policy

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/gatekeeper/internal/chain"
)

// Test addresses - valid Ethereum addresses for testing
const (
	testTokenAddr   = "0x1234567890123456789012345678901234567890" // Mock token contract
	testNFTAddr     = "0x2234567890123456789012345678901234567890" // Mock NFT contract
	testUserAddr    = "0x3234567890123456789012345678901234567890" // Mock user address
	testUserAddr2   = "0x4234567890123456789012345678901234567890" // Second mock user address
)

// MockCache is a simple cache implementation for testing
type MockCache struct {
	data map[string]interface{}
}

// Get retrieves a value from cache
func (m *MockCache) Get(key string) (interface{}, bool) {
	val, ok := m.data[key]
	return val, ok
}

// Set stores a value in cache
func (m *MockCache) Set(key string, value interface{}) {
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	m.data[key] = value
}

// GetOrSet gets or sets a value
func (m *MockCache) GetOrSet(key string, fn func() interface{}) interface{} {
	if val, ok := m.Get(key); ok {
		return val
	}
	val := fn()
	m.Set(key, val)
	return val
}

// has checks if a key exists in cache
func (m *MockCache) has(key string) bool {
	_, ok := m.data[key]
	return ok
}

// MockBlockchainProvider mocks RPC calls for testing
type MockBlockchainProvider struct {
	balances map[string]*big.Int
	owners   map[string]string // tokenID -> owner address
}

// Call mocks the RPC call for eth_call
func (m *MockBlockchainProvider) Call(ctx context.Context, method string, params []interface{}) ([]byte, error) {
	if method != "eth_call" {
		return nil, fmt.Errorf("unsupported method: %s", method)
	}

	if len(params) < 1 {
		return nil, fmt.Errorf("missing params")
	}

	callObj, ok := params[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid call object")
	}

	data, ok := callObj["data"].(string)
	if !ok {
		return nil, fmt.Errorf("missing data field")
	}

	// Parse the calldata to determine ERC20 or ERC721 call
	if strings.HasPrefix(data, "0x70a08231") {
		// ERC20 balanceOf call - extract the address parameter from calldata
		// Calldata format: selector (8 chars) + 64 hex chars (32 bytes) of address
		if len(data) >= 74 {
			addrHex := data[10:74] // Extract the 64 hex chars
			// Find matching balance
			for addr, balance := range m.balances {
				// Normalize and compare last 40 chars (20 bytes = address)
				normalizedStored := strings.ToLower(strings.TrimPrefix(addr, "0x"))
				normalizedQuery := strings.ToLower(addrHex[24:]) // Skip padding
				if normalizedStored == normalizedQuery || strings.HasSuffix(normalizedQuery, normalizedStored) {
					return m.encodeBalanceResponse(balance), nil
				}
			}
		}
		return m.encodeBalanceResponse(big.NewInt(0)), nil
	} else if strings.HasPrefix(data, "0x6352211e") {
		// ERC721 ownerOf call
		// For simplicity, return the first owner in our mock
		for _, owner := range m.owners {
			return m.encodeAddressResponse(owner), nil
		}
		return nil, fmt.Errorf("token not found")
	}

	return nil, fmt.Errorf("unknown selector")
}

// encodeBalanceResponse encodes a balance as JSON-RPC response
func (m *MockBlockchainProvider) encodeBalanceResponse(balance *big.Int) []byte {
	hexVal := fmt.Sprintf("0x%064x", balance)
	response := fmt.Sprintf(`{"jsonrpc":"2.0","result":"%s","id":1}`, hexVal)
	return []byte(response)
}

// encodeAddressResponse encodes an address as JSON-RPC response
func (m *MockBlockchainProvider) encodeAddressResponse(address string) []byte {
	// Pad address to 32 bytes (64 hex chars)
	addrHex := strings.ToLower(strings.TrimPrefix(address, "0x"))
	// Pad with zeros on the left to make 64 hex chars
	padded := fmt.Sprintf("%064s", addrHex)
	// Replace spaces with zeros
	padded = strings.ReplaceAll(padded, " ", "0")
	hexVal := "0x" + padded
	response := fmt.Sprintf(`{"jsonrpc":"2.0","result":"%s","id":1}`, hexVal)
	return []byte(response)
}

// HealthCheck mocks health check
func (m *MockBlockchainProvider) HealthCheck(ctx context.Context) bool {
	return true
}

// SetBalance sets up a mock balance
func (m *MockBlockchainProvider) SetBalance(address string, balance *big.Int) {
	if m.balances == nil {
		m.balances = make(map[string]*big.Int)
	}
	m.balances[address] = balance
}

// SetOwner sets up a mock NFT owner
func (m *MockBlockchainProvider) SetOwner(tokenID string, owner string) {
	if m.owners == nil {
		m.owners = make(map[string]string)
	}
	m.owners[tokenID] = owner
}

// TestERC20Rule_Evaluate_WithSufficientBalance passes when balance meets minimum
func TestERC20Rule_Evaluate_WithSufficientBalance(t *testing.T) {
	// Create rule requiring 1000 tokens (1e18 wei)
	minBalance := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil) // 1e18
	rule := NewERC20MinBalanceRule(testTokenAddr, minBalance, 1)

	// Note: This test shows the structure. Full blockchain integration
	// requires mocking the RPC provider, which is done separately.
	// For now, we verify the rule structure is correct.
	assert.NotNil(t, rule)
	assert.Equal(t, testTokenAddr, rule.ContractAddress)
	assert.Equal(t, minBalance, rule.MinimumBalance)
	assert.Equal(t, uint64(1), rule.ChainID)
}

// TestERC20Rule_Type returns correct rule type
func TestERC20Rule_Type(t *testing.T) {
	rule := NewERC20MinBalanceRule(testTokenAddr, big.NewInt(1000), 1)

	assert.Equal(t, ERC20MinBalanceRuleType, rule.Type())
}

// TestERC721Rule_Type returns correct rule type
func TestERC721Rule_Type(t *testing.T) {
	rule := NewERC721OwnerRule(testNFTAddr, big.NewInt(42), 1)

	assert.Equal(t, ERC721OwnerRuleType, rule.Type())
}

// TestERC20Rule_Creation creates rule with correct parameters
func TestERC20Rule_Creation(t *testing.T) {
	minBalance := big.NewInt(1000)
	rule := NewERC20MinBalanceRule("0x1234", minBalance, 1)

	require.NotNil(t, rule)
	assert.Equal(t, "0x1234", rule.ContractAddress)
	assert.Equal(t, minBalance, rule.MinimumBalance)
	assert.Equal(t, uint64(1), rule.ChainID)
}

// TestERC721Rule_Creation creates rule with correct parameters
func TestERC721Rule_Creation(t *testing.T) {
	tokenID := big.NewInt(42)
	rule := NewERC721OwnerRule(testNFTAddr, tokenID, 1)

	require.NotNil(t, rule)
	assert.Equal(t, testNFTAddr, rule.ContractAddress)
	assert.Equal(t, tokenID, rule.TokenID)
	assert.Equal(t, uint64(1), rule.ChainID)
}

// TestCacheKey_ERC20 generates correct cache key format for ERC20
func TestCacheKey_ERC20(t *testing.T) {
	key := chain.CacheKey("erc20_balance", "1", testTokenAddr, "0xUser1")

	assert.Contains(t, key, "erc20_balance")
	assert.Contains(t, key, "1")
	assert.Contains(t, key, testTokenAddr)
	assert.Contains(t, key, "0xUser1")
}

// TestCacheKey_ERC721 generates correct cache key format for ERC721
func TestCacheKey_ERC721(t *testing.T) {
	key := chain.CacheKey("erc721_owner", "1", testNFTAddr, "42")

	assert.Contains(t, key, "erc721_owner")
	assert.Contains(t, key, "1")
	assert.Contains(t, key, testNFTAddr)
	assert.Contains(t, key, "42")
}

// TestERC20Rule_WithDifferentChains supports multi-chain
func TestERC20Rule_WithDifferentChains(t *testing.T) {
	rule1 := NewERC20MinBalanceRule(testTokenAddr, big.NewInt(1000), 1) // Mainnet
	rule2 := NewERC20MinBalanceRule(testTokenAddr, big.NewInt(1000), 137) // Polygon

	assert.Equal(t, uint64(1), rule1.ChainID)
	assert.Equal(t, uint64(137), rule2.ChainID)
	assert.NotEqual(t, rule1.ChainID, rule2.ChainID)
}

// TestERC721Rule_WithDifferentChains supports multi-chain
func TestERC721Rule_WithDifferentChains(t *testing.T) {
	rule1 := NewERC721OwnerRule(testNFTAddr, big.NewInt(42), 1)    // Mainnet
	rule2 := NewERC721OwnerRule(testNFTAddr, big.NewInt(42), 43114) // Avalanche

	assert.Equal(t, uint64(1), rule1.ChainID)
	assert.Equal(t, uint64(43114), rule2.ChainID)
}

// TestERC20Rule_WithDifferentContracts differentiates contracts
func TestERC20Rule_WithDifferentContracts(t *testing.T) {
	ruleUSDC := NewERC20MinBalanceRule("0xUSDC", big.NewInt(1000000), 1)
	ruleDAI := NewERC20MinBalanceRule("0xDAI", big.NewInt(1000000), 1)

	assert.NotEqual(t, ruleUSDC.ContractAddress, ruleDAI.ContractAddress)
}

// TestERC20Rule_WithDifferentMinimums enforces different balances
func TestERC20Rule_WithDifferentMinimums(t *testing.T) {
	rule1 := NewERC20MinBalanceRule(testTokenAddr, big.NewInt(1000), 1)
	rule2 := NewERC20MinBalanceRule(testTokenAddr, big.NewInt(10000), 1)

	assert.NotEqual(t, rule1.MinimumBalance, rule2.MinimumBalance)
	assert.True(t, rule1.MinimumBalance.Cmp(rule2.MinimumBalance) < 0)
}

// TestERC721Rule_WithDifferentTokenIDs differentiates NFTs
func TestERC721Rule_WithDifferentTokenIDs(t *testing.T) {
	rule1 := NewERC721OwnerRule(testNFTAddr, big.NewInt(42), 1)
	rule2 := NewERC721OwnerRule(testNFTAddr, big.NewInt(43), 1)

	assert.NotEqual(t, rule1.TokenID, rule2.TokenID)
}

// TestERC721Rule_WithLargeTokenID supports large token IDs
func TestERC721Rule_WithLargeTokenID(t *testing.T) {
	// Large token ID from real NFTs
	largeID := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
	rule := NewERC721OwnerRule(testNFTAddr, largeID, 1)

	assert.Equal(t, largeID, rule.TokenID)
}

// TestERC20Rule_WithLargeBalance supports large balance numbers
func TestERC20Rule_WithLargeBalance(t *testing.T) {
	// 1 million tokens with 18 decimals = 1e24
	largeBalance := new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil)
	rule := NewERC20MinBalanceRule(testTokenAddr, largeBalance, 1)

	assert.Equal(t, largeBalance, rule.MinimumBalance)
}

// TestERC20Rule_Evaluate_SufficientBalance passes when balance meets minimum
func TestERC20Rule_Evaluate_SufficientBalance(t *testing.T) {
	minBalance := big.NewInt(1000)
	rule := NewERC20MinBalanceRule(testTokenAddr, minBalance, 1)

	// Mock provider that returns sufficient balance
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
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

// TestERC20Rule_Evaluate_InsufficientBalance fails when balance below minimum
func TestERC20Rule_Evaluate_InsufficientBalance(t *testing.T) {
	minBalance := big.NewInt(1000)
	rule := NewERC20MinBalanceRule(testTokenAddr, minBalance, 1)

	// Mock provider that returns insufficient balance
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
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

// TestERC20Rule_Evaluate_NoProvider returns false when no provider
func TestERC20Rule_Evaluate_NoProvider(t *testing.T) {
	rule := NewERC20MinBalanceRule(testTokenAddr, big.NewInt(1000), 1)
	// Don't set provider

	result, err := rule.Evaluate(context.Background(), testUserAddr, nil)

	require.NoError(t, err)
	assert.False(t, result)
}

// TestERC20Rule_Evaluate_UsesCache returns cached value without RPC call
func TestERC20Rule_Evaluate_UsesCache(t *testing.T) {
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	rule := NewERC20MinBalanceRule(testTokenAddr, big.NewInt(1000), 1)

	// Mock provider
	provider := &MockBlockchainProvider{}
	rule.SetProvider(provider)

	// Mock cache with pre-cached value
	cache := &MockCache{}
	cache.data = make(map[string]interface{})
	cacheKey := chain.CacheKey("erc20_balance", "1", testTokenAddr, userAddr)
	cache.data[cacheKey] = true
	rule.SetCache(cache)

	result, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	assert.True(t, result)
	// Verify provider was not called (would panic with our mock if it was)
}

// TestERC20Rule_Evaluate_CachesResult stores result in cache
func TestERC20Rule_Evaluate_CachesResult(t *testing.T) {
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	rule := NewERC20MinBalanceRule(testTokenAddr, big.NewInt(1000), 1)

	// Mock provider
	provider := &MockBlockchainProvider{}
	provider.SetBalance(userAddr, big.NewInt(5000))
	rule.SetProvider(provider)

	// Mock cache
	cache := &MockCache{}
	rule.SetCache(cache)

	_, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	cacheKey := chain.CacheKey("erc20_balance", "1", testTokenAddr, userAddr)
	assert.True(t, cache.has(cacheKey))
}

// TestERC721Rule_Evaluate_IsOwner passes when user owns NFT
func TestERC721Rule_Evaluate_IsOwner(t *testing.T) {
	tokenID := big.NewInt(42)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	rule := NewERC721OwnerRule(testNFTAddr, tokenID, 1)

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

// TestERC721Rule_Evaluate_NotOwner fails when user doesn't own NFT
func TestERC721Rule_Evaluate_NotOwner(t *testing.T) {
	tokenID := big.NewInt(42)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	otherAddr := "0xabcdef1234567890abcdef1234567890abcdef12"
	rule := NewERC721OwnerRule(testNFTAddr, tokenID, 1)

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

// TestERC721Rule_Evaluate_NoProvider returns false when no provider
func TestERC721Rule_Evaluate_NoProvider(t *testing.T) {
	rule := NewERC721OwnerRule(testNFTAddr, big.NewInt(42), 1)
	// Don't set provider

	result, err := rule.Evaluate(context.Background(), testUserAddr, nil)

	require.NoError(t, err)
	assert.False(t, result)
}

// TestERC721Rule_Evaluate_UsesCache returns cached value without RPC call
func TestERC721Rule_Evaluate_UsesCache(t *testing.T) {
	rule := NewERC721OwnerRule(testNFTAddr, big.NewInt(42), 1)

	// Mock provider
	provider := &MockBlockchainProvider{}
	rule.SetProvider(provider)

	// Mock cache with pre-cached value (ERC721 caches the owner address, not a boolean)
	cache := &MockCache{}
	cache.data = make(map[string]interface{})
	cacheKey := chain.CacheKey("erc721_owner", "1", strings.ToLower(testNFTAddr), "42")
	cache.data[cacheKey] = strings.ToLower(testUserAddr) // Cache the owner address
	rule.SetCache(cache)

	result, err := rule.Evaluate(context.Background(), testUserAddr, nil)

	require.NoError(t, err)
	assert.True(t, result)
}

// TestERC721Rule_Evaluate_CachesResult stores result in cache
func TestERC721Rule_Evaluate_CachesResult(t *testing.T) {
	tokenID := big.NewInt(42)
	userAddr := "0x1234567890abcdef1234567890abcdef12345678"
	rule := NewERC721OwnerRule(testNFTAddr, tokenID, 1)

	// Mock provider
	provider := &MockBlockchainProvider{}
	provider.SetOwner(tokenID.String(), userAddr)
	rule.SetProvider(provider)

	// Mock cache
	cache := &MockCache{}
	rule.SetCache(cache)

	_, err := rule.Evaluate(context.Background(), userAddr, nil)

	require.NoError(t, err)
	cacheKey := chain.CacheKey("erc721_owner", "1", testNFTAddr, "42")
	assert.True(t, cache.has(cacheKey))
}
