package policy

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  string          `json:"result"`
	Error   *JSONRPCError   `json:"error,omitempty"`
	ID      interface{}     `json:"id"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ERC20Selectors for standard ERC20 methods
const (
	ERC20BalanceOfSelector = "0x70a08231"
)

// ERC721Selectors for standard ERC721 methods
const (
	ERC721OwnerOfSelector = "0x6352211e"
)

// parseJSONRPCResponse parses a JSON-RPC response and extracts the result hex value
func parseJSONRPCResponse(data []byte) (string, error) {
	var resp JSONRPCResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse JSON-RPC response: %w", err)
	}

	// Check for RPC error
	if resp.Error != nil {
		return "", fmt.Errorf("RPC error: %s", resp.Error.Message)
	}

	// Result should be a hex string
	if resp.Result == "" {
		return "", fmt.Errorf("empty result from RPC call")
	}

	return resp.Result, nil
}

// encodeAddress encodes an Ethereum address to 32-byte hex string for contract call
func encodeAddress(address string) string {
	// Remove 0x prefix if present
	addr := strings.TrimPrefix(address, "0x")
	// Pad to 64 hex characters (32 bytes)
	return "0x" + strings.Repeat("0", 64-len(addr)) + addr
}

// encodeUint256 encodes a big.Int as 32-byte hex string for contract call
func encodeUint256(value *big.Int) string {
	// Convert to hex and pad to 64 characters (32 bytes)
	hexStr := fmt.Sprintf("%x", value)
	return "0x" + strings.Repeat("0", 64-len(hexStr)) + hexStr
}

// decodeAddress decodes a 32-byte hex string to an Ethereum address
func decodeAddress(hexValue string) (string, error) {
	// Remove 0x prefix if present
	hexValue = strings.TrimPrefix(hexValue, "0x")

	// Should be 64 hex characters (32 bytes)
	if len(hexValue) != 64 {
		return "", fmt.Errorf("invalid hex length: expected 64, got %d", len(hexValue))
	}

	// The last 20 bytes (40 hex chars) are the address
	addrHex := hexValue[24:] // Skip first 24 chars (12 bytes of padding)
	return "0x" + addrHex, nil
}

// decodeUint256 decodes a 32-byte hex string to a big.Int
func decodeUint256(hexValue string) (*big.Int, error) {
	// Remove 0x prefix if present
	hexValue = strings.TrimPrefix(hexValue, "0x")

	// Should be 64 hex characters (32 bytes)
	if len(hexValue) != 64 {
		return nil, fmt.Errorf("invalid hex length: expected 64, got %d", len(hexValue))
	}

	// Parse as big.Int from hex
	value := new(big.Int)
	if _, ok := value.SetString(hexValue, 16); !ok {
		return nil, fmt.Errorf("failed to parse hex value: %s", hexValue)
	}

	return value, nil
}

// encodeERC20BalanceOfCall encodes a call to ERC20 balanceOf(address)
// Returns the calldata hex string
func encodeERC20BalanceOfCall(tokenAddress, userAddress string) string {
	// balanceOf selector
	calldata := ERC20BalanceOfSelector
	// Add encoded address parameter
	calldata += strings.TrimPrefix(encodeAddress(userAddress), "0x")
	return calldata
}

// encodeERC721OwnerOfCall encodes a call to ERC721 ownerOf(uint256)
// Returns the calldata hex string
func encodeERC721OwnerOfCall(nftAddress string, tokenID *big.Int) string {
	// ownerOf selector
	calldata := ERC721OwnerOfSelector
	// Add encoded tokenId parameter
	calldata += strings.TrimPrefix(encodeUint256(tokenID), "0x")
	return calldata
}

// normalizeCacheKey generates a consistent cache key for blockchain results
func normalizeCacheKey(dataType, chainID string, contract, identifier string) string {
	return fmt.Sprintf("%s:%s:%s:%s", dataType, chainID, strings.ToLower(contract), strings.ToLower(identifier))
}
