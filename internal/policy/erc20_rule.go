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

// ERC20MinBalanceRule checks if user has minimum ERC20 token balance
type ERC20MinBalanceRule struct {
	ContractAddress string
	MinimumBalance  *big.Int
	ChainID         uint64
	// cache and provider will be set by manager
	cache    CacheProvider
	provider BlockchainProvider
	logger   *zap.Logger
}

// NewERC20MinBalanceRule creates a new ERC20 balance rule
func NewERC20MinBalanceRule(contractAddress string, minimumBalance *big.Int, chainID uint64) *ERC20MinBalanceRule {
	logger, _ := zap.NewProduction()
	return &ERC20MinBalanceRule{
		ContractAddress: contractAddress,
		MinimumBalance:  minimumBalance,
		ChainID:         chainID,
		logger:          logger,
	}
}

// Type returns the rule type
func (r *ERC20MinBalanceRule) Type() RuleType {
	return ERC20MinBalanceRuleType
}

// Validate checks if the rule parameters are valid
func (r *ERC20MinBalanceRule) Validate() error {
	// Validate contract address
	if !isValidAddress(r.ContractAddress) {
		return fmt.Errorf("invalid contract address: %s", r.ContractAddress)
	}

	// Validate minimum balance
	if r.MinimumBalance == nil {
		return fmt.Errorf("minimum balance cannot be nil")
	}
	if r.MinimumBalance.Sign() < 0 {
		return fmt.Errorf("minimum balance cannot be negative")
	}

	// Validate chain ID
	if r.ChainID == 0 {
		return fmt.Errorf("chain ID cannot be zero")
	}

	return nil
}

// Evaluate checks ERC20 balance (requires provider and cache to be set)
// This implementation follows fail-closed security: on any error, return false
func (r *ERC20MinBalanceRule) Evaluate(ctx context.Context, address string, claims *auth.Claims) (bool, error) {
	// Validate inputs
	if !isValidAddress(address) {
		r.logger.Error("invalid address format",
			zap.String("address", address),
			zap.String("rule", "ERC20MinBalance"))
		return false, nil // Fail-closed
	}

	if !isValidAddress(r.ContractAddress) {
		r.logger.Error("invalid token address format",
			zap.String("token", r.ContractAddress),
			zap.String("rule", "ERC20MinBalance"))
		return false, nil // Fail-closed
	}

	// If no provider configured, evaluate to false (fail-closed)
	if r.provider == nil {
		r.logger.Warn("no blockchain provider configured",
			zap.String("rule", "ERC20MinBalance"))
		return false, nil
	}

	// Normalize addresses for cache key (use lowercase for consistency)
	normalizedToken := strings.ToLower(r.ContractAddress)
	normalizedAddr := strings.ToLower(address)

	// Generate cache key: "erc20_balance:{chainID}:{token}:{address}"
	chainIDStr := strconv.FormatUint(r.ChainID, 10)
	cacheKey := chain.CacheKey("erc20_balance", chainIDStr, normalizedToken, normalizedAddr)

	// Try to get from cache first
	if r.cache != nil {
		if cachedResult, ok := r.cache.Get(cacheKey); ok {
			if hasBalance, ok := cachedResult.(bool); ok {
				r.logger.Debug("cache hit for ERC20 balance",
					zap.String("cacheKey", cacheKey),
					zap.Bool("result", hasBalance))
				return hasBalance, nil
			}
		}
	}

	// Not in cache, make RPC call
	// Encode ERC20 balanceOf(address) call
	calldata := encodeERC20BalanceOfCall(r.ContractAddress, address)

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
		r.logger.Error("RPC call failed for ERC20 balance",
			zap.Error(err),
			zap.String("token", r.ContractAddress),
			zap.String("address", address),
			zap.Uint64("chainID", r.ChainID))
		return false, nil
	}

	// Parse JSON-RPC response
	resultHex, err := parseJSONRPCResponse(response)
	if err != nil {
		r.logger.Error("failed to parse RPC response",
			zap.Error(err),
			zap.String("token", r.ContractAddress),
			zap.String("address", address))
		return false, nil
	}

	// Decode the balance from hex using chain utilities
	balance, err := decodeUint256(resultHex)
	if err != nil {
		r.logger.Error("failed to decode balance",
			zap.Error(err),
			zap.String("resultHex", resultHex))
		return false, nil
	}

	// Compare with minimum balance
	hasBalance := balance.Cmp(r.MinimumBalance) >= 0

	r.logger.Info("ERC20 balance check completed",
		zap.String("address", address),
		zap.String("token", r.ContractAddress),
		zap.String("balance", balance.String()),
		zap.String("minimum", r.MinimumBalance.String()),
		zap.Bool("hasBalance", hasBalance))

	// Cache the result with TTL
	if r.cache != nil {
		r.cache.Set(cacheKey, hasBalance)
		r.logger.Debug("cached ERC20 balance result",
			zap.String("cacheKey", cacheKey),
			zap.Bool("result", hasBalance))
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

// SetLogger sets the logger for the rule
func (r *ERC20MinBalanceRule) SetLogger(logger *zap.Logger) {
	r.logger = logger
}
