package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

// ERC721Checker handles ERC721 (NFT) ownership checks
type ERC721Checker struct {
	provider *Provider
	cache    *Cache
}

// NewERC721Checker creates a new ERC721 checker
func NewERC721Checker(provider *Provider, cache *Cache) *ERC721Checker {
	return &ERC721Checker{
		provider: provider,
		cache:    cache,
	}
}

// OwnerOf retrieves the owner of an ERC721 token
// Contract: ownerOf(uint256 tokenId) returns (address)
func (e *ERC721Checker) OwnerOf(ctx context.Context, contract string, tokenID string) (string, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("erc721:owner:%s:%s", contract, tokenID)
	if e.cache != nil {
		if cached, ok := e.cache.Get(cacheKey); ok {
			if owner, ok := cached.(string); ok {
				return owner, nil
			}
		}
	}

	// Encode function call: ownerOf(uint256)
	// Function signature: ownerOf(uint256) = 0x6352211e
	// Encoding: 0x6352211e + tokenID as uint256 (left-padded to 64 hex chars)
	tokenIDInt := new(big.Int)
	tokenIDInt.SetString(tokenID, 10)
	paddedTokenID := fmt.Sprintf("%064x", tokenIDInt)
	data := "0x6352211e" + paddedTokenID

	callObj := []interface{}{
		map[string]interface{}{
			"to":   contract,
			"data": data,
		},
		"latest",
	}

	response, err := e.provider.Call(ctx, "eth_call", callObj)
	if err != nil {
		return "", fmt.Errorf("failed to call ownerOf: %w", err)
	}

	// Parse JSON-RPC response
	var rpcResp struct {
		Result string `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(response, &rpcResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if rpcResp.Error != nil {
		return "", fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
	}

	// Extract address from padded result
	// Result is 32 bytes (64 hex chars), address is last 20 bytes (40 hex chars)
	result := rpcResp.Result
	if result == "0x" || result == "" {
		return "", fmt.Errorf("empty owner address")
	}

	// Remove 0x prefix if present
	if len(result) > 2 && result[:2] == "0x" {
		result = result[2:]
	}

	// Extract last 40 hex chars (20 bytes) as address
	if len(result) < 40 {
		return "", fmt.Errorf("invalid owner address format: %s", result)
	}

	owner := "0x" + result[len(result)-40:]

	// Cache the result
	if e.cache != nil {
		e.cache.Set(cacheKey, owner)
	}

	return owner, nil
}

// IsOwner checks if an address owns a specific token
func (e *ERC721Checker) IsOwner(ctx context.Context, contract string, tokenID string, address string) (bool, error) {
	owner, err := e.OwnerOf(ctx, contract, tokenID)
	if err != nil {
		return false, err
	}

	// Normalize addresses for comparison (case-insensitive)
	return normalizeAddress(owner) == normalizeAddress(address), nil
}

// BalanceOf retrieves the number of tokens owned by an address
// Contract: balanceOf(address account) returns (uint256)
func (e *ERC721Checker) BalanceOf(ctx context.Context, contract string, account string) (uint64, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("erc721:balance:%s:%s", contract, account)
	if e.cache != nil {
		if cached, ok := e.cache.Get(cacheKey); ok {
			if balance, ok := cached.(uint64); ok {
				return balance, nil
			}
		}
	}

	// Encode function call: balanceOf(address)
	// Function signature: balanceOf(address) = 0x70a08231
	paddedAddr := account
	if len(account) > 2 && account[:2] == "0x" {
		paddedAddr = account[2:]
	}
	for len(paddedAddr) < 64 {
		paddedAddr = "0" + paddedAddr
	}
	data := "0x70a08231" + paddedAddr

	callObj := []interface{}{
		map[string]interface{}{
			"to":   contract,
			"data": data,
		},
		"latest",
	}

	response, err := e.provider.Call(ctx, "eth_call", callObj)
	if err != nil {
		return 0, fmt.Errorf("failed to call balanceOf: %w", err)
	}

	// Parse JSON-RPC response
	var rpcResp struct {
		Result string `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(response, &rpcResp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	if rpcResp.Error != nil {
		return 0, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
	}

	// Convert hex string to uint64
	result := rpcResp.Result
	if result == "0x" || result == "" {
		return 0, nil
	}

	hexStr := result
	if len(result) > 2 && result[:2] == "0x" {
		hexStr = result[2:]
	}

	balance := new(big.Int)
	_, ok := balance.SetString(hexStr, 16)
	if !ok {
		return 0, fmt.Errorf("failed to parse balance: %s", result)
	}

	uint64Balance := balance.Uint64()

	// Cache the result
	if e.cache != nil {
		e.cache.Set(cacheKey, uint64Balance)
	}

	return uint64Balance, nil
}

// HasTokens checks if an account owns at least one NFT
func (e *ERC721Checker) HasTokens(ctx context.Context, contract string, account string) (bool, error) {
	balance, err := e.BalanceOf(ctx, contract, account)
	if err != nil {
		return false, err
	}

	return balance > 0, nil
}

// normalizeAddress normalizes an Ethereum address for comparison
func normalizeAddress(addr string) string {
	// Remove 0x prefix
	if len(addr) > 2 && addr[:2] == "0x" {
		addr = addr[2:]
	}
	// Convert to lowercase
	return "0x" + strings.ToLower(addr)
}
