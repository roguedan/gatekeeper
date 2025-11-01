package chain

import (
	"context"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestERC20BalanceOf_Success retrieves ERC20 balance correctly
func TestERC20BalanceOf_Success(t *testing.T) {
	// Balance of 100 tokens (18 decimals) = 100000000000000000000
	balanceHex := "0x56bc75e2d63100000" // 100e18 in hex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"` + balanceHex + `","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	checker := NewERC20Checker(provider, nil)
	ctx := context.Background()

	balance, err := checker.BalanceOf(ctx, "0x1234567890123456789012345678901234567890", "0x6B175474E89094C44Da98b954EedeAC495271d0F")

	require.NoError(t, err)
	assert.NotNil(t, balance)
	expected := new(big.Int)
	expected.SetString("100000000000000000000", 10)
	assert.Equal(t, expected, balance)
}

// TestERC20BalanceOf_WithCache uses cached balance
func TestERC20BalanceOf_WithCache(t *testing.T) {
	callCount := 0
	balanceHex := "0x56bc75e2d63100000" // 100e18
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"` + balanceHex + `","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	cache := NewCache(testCacheTTL)
	checker := NewERC20Checker(provider, cache)
	ctx := context.Background()

	// First call
	_, err := checker.BalanceOf(ctx, "0x1234567890123456789012345678901234567890", "0x6B175474E89094C44Da98b954EedeAC495271d0F")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Second call should use cache
	_, err = checker.BalanceOf(ctx, "0x1234567890123456789012345678901234567890", "0x6B175474E89094C44Da98b954EedeAC495271d0F")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount) // No additional call
}

// TestERC20BalanceOf_ZeroBalance returns zero balance
func TestERC20BalanceOf_ZeroBalance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"0x0","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	checker := NewERC20Checker(provider, nil)
	ctx := context.Background()

	balance, err := checker.BalanceOf(ctx, "0x1234567890123456789012345678901234567890", "0x6B175474E89094C44Da98b954EedeAC495271d0F")

	require.NoError(t, err)
	assert.Equal(t, 0, balance.Cmp(big.NewInt(0)))
}

// TestERC20BalanceOf_InvalidAddress returns error
func TestERC20BalanceOf_InvalidAddress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid address"},"id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	checker := NewERC20Checker(provider, nil)
	ctx := context.Background()

	_, err := checker.BalanceOf(ctx, "invalid", "0x6B175474E89094C44Da98b954EedeAC495271d0F")

	assert.Error(t, err)
}

// TestERC20HasMinBalance succeeds when balance >= min
func TestERC20HasMinBalance_Success(t *testing.T) {
	balanceHex := "0x56bc75e2d63100000" // 100e18
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"` + balanceHex + `","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	checker := NewERC20Checker(provider, nil)
	ctx := context.Background()

	minBalance := new(big.Int)
	minBalance.SetString("50000000000000000000", 10) // 50e18
	hasBalance, err := checker.HasMinBalance(ctx, "0x1234567890123456789012345678901234567890", "0x6B175474E89094C44Da98b954EedeAC495271d0F", minBalance)

	require.NoError(t, err)
	assert.True(t, hasBalance)
}

// TestERC20HasMinBalance_InsufficientBalance fails when balance < min
func TestERC20HasMinBalance_InsufficientBalance(t *testing.T) {
	balanceHex := "0x000000000000000000000000000000000000000000000000016345785d8a0000" // 0.1e18
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"` + balanceHex + `","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	checker := NewERC20Checker(provider, nil)
	ctx := context.Background()

	minBalance := new(big.Int)
	minBalance.SetString("50000000000000000000", 10) // 50e18
	hasBalance, err := checker.HasMinBalance(ctx, "0x1234567890123456789012345678901234567890", "0x6B175474E89094C44Da98b954EedeAC495271d0F", minBalance)

	require.NoError(t, err)
	assert.False(t, hasBalance)
}

// TestERC20Decimals retrieves token decimals
func TestERC20Decimals_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// 18 as uint256
		w.Write([]byte(`{"jsonrpc":"2.0","result":"0x12","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	checker := NewERC20Checker(provider, nil)
	ctx := context.Background()

	decimals, err := checker.Decimals(ctx, "0x6B175474E89094C44Da98b954EedeAC495271d0F")

	require.NoError(t, err)
	assert.Equal(t, uint8(18), decimals)
}

// TestERC20ContextCancellation respects context cancellation
func TestERC20ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"0x0","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	checker := NewERC20Checker(provider, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := checker.BalanceOf(ctx, "0x1234567890123456789012345678901234567890", "0x6B175474E89094C44Da98b954EedeAC495271d0F")

	assert.Error(t, err)
}
