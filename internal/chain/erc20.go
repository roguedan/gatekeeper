package chain

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

// ERC20Checker handles ERC20 token balance checks
type ERC20Checker struct {
	provider *Provider
	cache    *Cache
}

// NewERC20Checker creates a new ERC20 checker
func NewERC20Checker(provider *Provider, cache *Cache) *ERC20Checker {
	return &ERC20Checker{
		provider: provider,
		cache:    cache,
	}
}

// BalanceOf retrieves the balance of an address for an ERC20 token
// Contract: balanceOf(address account) returns (uint256)
func (e *ERC20Checker) BalanceOf(ctx context.Context, account string, tokenContract string) (*big.Int, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("erc20:balance:%s:%s", tokenContract, account)
	if e.cache != nil {
		if cached, ok := e.cache.Get(cacheKey); ok {
			if balance, ok := cached.(*big.Int); ok {
				return balance, nil
			}
		}
	}

	// Encode function call: balanceOf(address)
	// Function signature: balanceOf(address) = 0x70a08231
	// Encoding: 0x70a08231 + left-padded address (32 bytes)
	paddedAddr := account
	if len(account) > 2 && account[:2] == "0x" {
		paddedAddr = account[2:]
	}
	// Left-pad address to 64 hex characters (32 bytes)
	for len(paddedAddr) < 64 {
		paddedAddr = "0" + paddedAddr
	}
	data := "0x70a08231" + paddedAddr

	// Make eth_call
	callObj := []interface{}{
		map[string]interface{}{
			"to":   tokenContract,
			"data": data,
		},
		"latest",
	}

	response, err := e.provider.Call(ctx, "eth_call", callObj)
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf: %w", err)
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
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
	}

	// Convert hex string to big.Int
	balance := new(big.Int)
	result := rpcResp.Result
	if result == "0x" || result == "" {
		return big.NewInt(0), nil
	}

	// Remove 0x prefix and parse
	hexStr := result[2:]
	_, ok := balance.SetString(hexStr, 16)
	if !ok {
		return nil, fmt.Errorf("failed to parse balance: %s", result)
	}

	// Cache the result
	if e.cache != nil {
		e.cache.Set(cacheKey, balance)
	}

	return balance, nil
}

// HasMinBalance checks if account has at least minBalance of token
func (e *ERC20Checker) HasMinBalance(ctx context.Context, account string, tokenContract string, minBalance *big.Int) (bool, error) {
	balance, err := e.BalanceOf(ctx, account, tokenContract)
	if err != nil {
		return false, err
	}

	return balance.Cmp(minBalance) >= 0, nil
}

// Decimals retrieves the decimals for an ERC20 token
// Contract: decimals() returns (uint8)
func (e *ERC20Checker) Decimals(ctx context.Context, tokenContract string) (uint8, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("erc20:decimals:%s", tokenContract)
	if e.cache != nil {
		if cached, ok := e.cache.Get(cacheKey); ok {
			if decimals, ok := cached.(uint8); ok {
				return decimals, nil
			}
		}
	}

	// Function signature: decimals() = 0x313ce567
	data := "0x313ce567"

	callObj := []interface{}{
		map[string]interface{}{
			"to":   tokenContract,
			"data": data,
		},
		"latest",
	}

	response, err := e.provider.Call(ctx, "eth_call", callObj)
	if err != nil {
		return 0, fmt.Errorf("failed to call decimals: %w", err)
	}

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

	// Parse result as uint8
	result := rpcResp.Result
	if result == "0x" || result == "" {
		return 0, nil
	}

	hexStr := result[2:]
	// Remove leading zeros and parse
	val := new(big.Int)
	_, ok := val.SetString(hexStr, 16)
	if !ok {
		return 0, fmt.Errorf("failed to parse decimals: %s", result)
	}

	decimals := uint8(val.Uint64())

	// Cache the result
	if e.cache != nil {
		e.cache.Set(cacheKey, decimals)
	}

	return decimals, nil
}

// Name retrieves the name of an ERC20 token
// Contract: name() returns (string)
func (e *ERC20Checker) Name(ctx context.Context, tokenContract string) (string, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("erc20:name:%s", tokenContract)
	if e.cache != nil {
		if cached, ok := e.cache.Get(cacheKey); ok {
			if name, ok := cached.(string); ok {
				return name, nil
			}
		}
	}

	// Function signature: name() = 0x06fdde03
	data := "0x06fdde03"

	callObj := []interface{}{
		map[string]interface{}{
			"to":   tokenContract,
			"data": data,
		},
		"latest",
	}

	response, err := e.provider.Call(ctx, "eth_call", callObj)
	if err != nil {
		return "", fmt.Errorf("failed to call name: %w", err)
	}

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

	// Decode string from ABI encoding
	name, err := decodeString(rpcResp.Result)
	if err != nil {
		return "", fmt.Errorf("failed to decode name: %w", err)
	}

	// Cache the result
	if e.cache != nil {
		e.cache.Set(cacheKey, name)
	}

	return name, nil
}

// Symbol retrieves the symbol of an ERC20 token
// Contract: symbol() returns (string)
func (e *ERC20Checker) Symbol(ctx context.Context, tokenContract string) (string, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("erc20:symbol:%s", tokenContract)
	if e.cache != nil {
		if cached, ok := e.cache.Get(cacheKey); ok {
			if symbol, ok := cached.(string); ok {
				return symbol, nil
			}
		}
	}

	// Function signature: symbol() = 0x95d89b41
	data := "0x95d89b41"

	callObj := []interface{}{
		map[string]interface{}{
			"to":   tokenContract,
			"data": data,
		},
		"latest",
	}

	response, err := e.provider.Call(ctx, "eth_call", callObj)
	if err != nil {
		return "", fmt.Errorf("failed to call symbol: %w", err)
	}

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

	// Decode string from ABI encoding
	symbol, err := decodeString(rpcResp.Result)
	if err != nil {
		return "", fmt.Errorf("failed to decode symbol: %w", err)
	}

	// Cache the result
	if e.cache != nil {
		e.cache.Set(cacheKey, symbol)
	}

	return symbol, nil
}

// decodeString decodes an ABI-encoded string
func decodeString(encoded string) (string, error) {
	if encoded == "0x" || encoded == "" {
		return "", nil
	}

	// Remove 0x prefix
	if len(encoded) > 2 && encoded[:2] == "0x" {
		encoded = encoded[2:]
	}

	// First 64 chars = offset (usually 0x20 = 32)
	// Next 64 chars = length
	// Following chars = actual string data (in chunks of 64)

	if len(encoded) < 128 {
		return "", fmt.Errorf("encoded string too short")
	}

	// Get length
	lengthHex := encoded[64:128]
	length := new(big.Int)
	length.SetString(lengthHex, 16)

	// Get string data
	dataHex := encoded[128:]
	dataBytes, err := hex.DecodeString(dataHex[:length.Uint64()*2])
	if err != nil {
		return "", fmt.Errorf("failed to decode string data: %w", err)
	}

	return string(dataBytes), nil
}

// testCacheTTL is the test cache TTL (exposed for tests)
const testCacheTTL = 1 * time.Minute
