package policy

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/yourusername/gatekeeper/internal/auth"
	"github.com/yourusername/gatekeeper/internal/chain"
	"go.uber.org/zap"
)

// ERC721OwnerRule checks if user owns a specific NFT
type ERC721OwnerRule struct {
	ContractAddress string
	TokenID         *big.Int
	ChainID         uint64
	// cache and provider will be set by manager
	cache    CacheProvider
	provider BlockchainProvider
	logger   *zap.Logger
}

// NewERC721OwnerRule creates a new NFT ownership rule
func NewERC721OwnerRule(contractAddress string, tokenID *big.Int, chainID uint64) *ERC721OwnerRule {
	logger, _ := zap.NewProduction()
	return &ERC721OwnerRule{
		ContractAddress: contractAddress,
		TokenID:         tokenID,
		ChainID:         chainID,
		logger:          logger,
	}
}

// Type returns the rule type
func (r *ERC721OwnerRule) Type() RuleType {
	return ERC721OwnerRuleType
}

// Validate checks if the rule parameters are valid
func (r *ERC721OwnerRule) Validate() error {
	// Validate contract address
	if !isValidAddress(r.ContractAddress) {
		return fmt.Errorf("invalid contract address: %s", r.ContractAddress)
	}

	// Validate token ID
	if r.TokenID == nil {
		return fmt.Errorf("token ID cannot be nil")
	}
	if r.TokenID.Sign() < 0 {
		return fmt.Errorf("token ID cannot be negative")
	}

	// Validate chain ID
	if r.ChainID == 0 {
		return fmt.Errorf("chain ID cannot be zero")
	}

	return nil
}

// Evaluate checks NFT ownership (requires provider and cache to be set)
// This implementation follows fail-closed security: on any error, return false
func (r *ERC721OwnerRule) Evaluate(ctx context.Context, address string, claims *auth.Claims) (bool, error) {
	// Validate inputs
	if !isValidAddress(address) {
		r.logger.Error("invalid address format",
			zap.String("address", address),
			zap.String("rule", "ERC721Owner"))
		return false, nil // Fail-closed
	}

	if !isValidAddress(r.ContractAddress) {
		r.logger.Error("invalid token address format",
			zap.String("token", r.ContractAddress),
			zap.String("rule", "ERC721Owner"))
		return false, nil // Fail-closed
	}

	// If no provider configured, evaluate to false (fail-closed)
	if r.provider == nil {
		r.logger.Warn("no blockchain provider configured",
			zap.String("rule", "ERC721Owner"))
		return false, nil
	}

	// Normalize addresses for cache key
	normalizedToken, err := checksumAddress(r.ContractAddress)
	if err != nil {
		normalizedToken = normalizeAddress(r.ContractAddress)
	}

	// Generate cache key: "erc721_owner:{chainID}:{token}:{tokenID}"
	// Note: Cache by token ID, not user address (owner can be shared)
	chainIDStr := strconv.FormatUint(r.ChainID, 10)
	tokenIDStr := r.TokenID.String()
	cacheKey := chain.CacheKey("erc721_owner", chainIDStr, normalizedToken, tokenIDStr)

	// Try to get from cache first
	if r.cache != nil {
		if cachedOwner, ok := r.cache.Get(cacheKey); ok {
			if ownerAddr, ok := cachedOwner.(string); ok {
				// Compare cached owner with requested address (case-insensitive)
				isOwner := strings.EqualFold(normalizeAddress(ownerAddr), normalizeAddress(address))
				r.logger.Debug("cache hit for ERC721 owner",
					zap.String("cacheKey", cacheKey),
					zap.String("cachedOwner", ownerAddr),
					zap.String("requestedAddress", address),
					zap.Bool("isOwner", isOwner))
				return isOwner, nil
			}
		}
	}

	// Not in cache, make RPC call
	// Encode ERC721 ownerOf(uint256) call
	calldata := encodeERC721OwnerOfCall(r.ContractAddress, r.TokenID)

	// Call provider with timeout handling
	response, err := r.provider.Call(ctx, "eth_call", []interface{}{
		map[string]interface{}{
			"to":   r.ContractAddress,
			"data": calldata,
		},
		"latest",
	})
	if err != nil {
		// Fail closed on RPC error
		// This could be because the token doesn't exist (burned) or network error
		r.logger.Error("RPC call failed for ERC721 owner",
			zap.Error(err),
			zap.String("token", r.ContractAddress),
			zap.String("tokenID", r.TokenID.String()),
			zap.Uint64("chainID", r.ChainID))
		return false, nil
	}

	// Parse JSON-RPC response
	resultHex, err := parseJSONRPCResponse(response)
	if err != nil {
		r.logger.Error("failed to parse RPC response",
			zap.Error(err),
			zap.String("token", r.ContractAddress),
			zap.String("tokenID", r.TokenID.String()))
		return false, nil
	}

	// Decode the owner address from hex
	ownerAddress, err := decodeAddress(resultHex)
	if err != nil {
		r.logger.Error("failed to decode owner address",
			zap.Error(err),
			zap.String("resultHex", resultHex))
		return false, nil
	}

	// Check for zero address (burned token)
	if isZeroAddress(ownerAddress) {
		r.logger.Info("token is burned (zero address owner)",
			zap.String("token", r.ContractAddress),
			zap.String("tokenID", r.TokenID.String()))
		return false, nil
	}

	// Compare addresses (case-insensitive)
	isOwner := strings.EqualFold(normalizeAddress(ownerAddress), normalizeAddress(address))

	r.logger.Info("ERC721 ownership check completed",
		zap.String("address", address),
		zap.String("token", r.ContractAddress),
		zap.String("tokenID", r.TokenID.String()),
		zap.String("owner", ownerAddress),
		zap.Bool("isOwner", isOwner))

	// Cache the owner address with TTL
	// Cache the owner, not the boolean result, so other users can check the same token
	if r.cache != nil {
		r.cache.Set(cacheKey, ownerAddress)
		r.logger.Debug("cached ERC721 owner",
			zap.String("cacheKey", cacheKey),
			zap.String("owner", ownerAddress))
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

// SetLogger sets the logger for the rule
func (r *ERC721OwnerRule) SetLogger(logger *zap.Logger) {
	r.logger = logger
}

// isZeroAddress checks if an address is the zero address (0x0000...0000)
func isZeroAddress(address string) bool {
	normalized := strings.ToLower(strings.TrimPrefix(address, "0x"))
	// Check if all characters are zeros
	for _, c := range normalized {
		if c != '0' {
			return false
		}
	}
	return len(normalized) == 40 // Must be exactly 40 hex chars
}
