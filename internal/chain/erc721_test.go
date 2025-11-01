package chain

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestERC721OwnerOf_Success retrieves NFT owner correctly
func TestERC721OwnerOf_Success(t *testing.T) {
	// Address in padded hex format as return value
	ownerHex := "0x0000000000000000000000001234567890123456789012345678901234567890"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"` + ownerHex + `","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	checker := NewERC721Checker(provider, nil)
	ctx := context.Background()

	owner, err := checker.OwnerOf(ctx, "0xNFTContract", "1")

	require.NoError(t, err)
	assert.Equal(t, "0x1234567890123456789012345678901234567890", owner)
}

// TestERC721OwnerOf_WithCache uses cached owner
func TestERC721OwnerOf_WithCache(t *testing.T) {
	callCount := 0
	ownerHex := "0x0000000000000000000000001234567890123456789012345678901234567890"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"` + ownerHex + `","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	cache := NewCache(testCacheTTL)
	checker := NewERC721Checker(provider, cache)
	ctx := context.Background()

	// First call
	_, err := checker.OwnerOf(ctx, "0xNFTContract", "1")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Second call should use cache
	_, err = checker.OwnerOf(ctx, "0xNFTContract", "1")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount) // No additional call
}

// TestERC721IsOwner verifies if address is owner of token
func TestERC721IsOwner_Success(t *testing.T) {
	ownerHex := "0x0000000000000000000000001234567890123456789012345678901234567890"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"` + ownerHex + `","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	checker := NewERC721Checker(provider, nil)
	ctx := context.Background()

	isOwner, err := checker.IsOwner(ctx, "0xNFTContract", "1", "0x1234567890123456789012345678901234567890")

	require.NoError(t, err)
	assert.True(t, isOwner)
}

// TestERC721IsOwner_NotOwner returns false when not owner
func TestERC721IsOwner_NotOwner(t *testing.T) {
	ownerHex := "0x0000000000000000000000001234567890123456789012345678901234567890"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"` + ownerHex + `","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	checker := NewERC721Checker(provider, nil)
	ctx := context.Background()

	isOwner, err := checker.IsOwner(ctx, "0xNFTContract", "1", "0xOtherAddress0000000000000000000000000000")

	require.NoError(t, err)
	assert.False(t, isOwner)
}

// TestERC721IsOwner_CaseInsensitive handles mixed case addresses
func TestERC721IsOwner_CaseInsensitive(t *testing.T) {
	ownerHex := "0x0000000000000000000000001234567890123456789012345678901234567890"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"` + ownerHex + `","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	checker := NewERC721Checker(provider, nil)
	ctx := context.Background()

	// Uppercase version of owner
	isOwner, err := checker.IsOwner(ctx, "0xNFTContract", "1", "0x1234567890123456789012345678901234567890")

	require.NoError(t, err)
	assert.True(t, isOwner)
}

// TestERC721BalanceOf retrieves account NFT balance
func TestERC721BalanceOf_Success(t *testing.T) {
	// Balance of 5 NFTs
	balanceHex := "0x5" // 5 in hex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"` + balanceHex + `","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	checker := NewERC721Checker(provider, nil)
	ctx := context.Background()

	balance, err := checker.BalanceOf(ctx, "0xNFTContract", "0x1234567890123456789012345678901234567890")

	require.NoError(t, err)
	assert.Equal(t, uint64(5), balance)
}

// TestERC721BalanceOf_ZeroBalance returns zero
func TestERC721BalanceOf_ZeroBalance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"0x0","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	checker := NewERC721Checker(provider, nil)
	ctx := context.Background()

	balance, err := checker.BalanceOf(ctx, "0xNFTContract", "0x1234567890123456789012345678901234567890")

	require.NoError(t, err)
	assert.Equal(t, uint64(0), balance)
}

// TestERC721HasTokens checks if account owns at least one NFT
func TestERC721HasTokens_Success(t *testing.T) {
	balanceHex := "0x5"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"` + balanceHex + `","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	checker := NewERC721Checker(provider, nil)
	ctx := context.Background()

	hasTokens, err := checker.HasTokens(ctx, "0xNFTContract", "0x1234567890123456789012345678901234567890")

	require.NoError(t, err)
	assert.True(t, hasTokens)
}

// TestERC721HasTokens_NoTokens returns false for zero balance
func TestERC721HasTokens_NoTokens(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"0x0","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	checker := NewERC721Checker(provider, nil)
	ctx := context.Background()

	hasTokens, err := checker.HasTokens(ctx, "0xNFTContract", "0x1234567890123456789012345678901234567890")

	require.NoError(t, err)
	assert.False(t, hasTokens)
}

// TestERC721ContextCancellation respects context cancellation
func TestERC721ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"0x0","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	checker := NewERC721Checker(provider, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := checker.OwnerOf(ctx, "0xNFTContract", "1")

	assert.Error(t, err)
}
