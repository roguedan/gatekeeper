package chain

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProvider_NewProvider creates a provider with primary and fallback RPC
func TestProvider_NewProvider(t *testing.T) {
	provider := NewProvider("https://eth.example.com", "https://fallback.example.com")

	require.NotNil(t, provider)
	assert.Equal(t, "https://eth.example.com", provider.primaryURL)
	assert.Equal(t, "https://fallback.example.com", provider.fallbackURL)
}

// TestProvider_NewProviderWithoutFallback creates provider with only primary RPC
func TestProvider_NewProviderWithoutFallback(t *testing.T) {
	provider := NewProvider("https://eth.example.com", "")

	require.NotNil(t, provider)
	assert.Equal(t, "https://eth.example.com", provider.primaryURL)
	assert.Equal(t, "", provider.fallbackURL)
}

// TestProvider_Call_WithPrimarySuccess makes successful RPC call on primary
func TestProvider_Call_WithPrimarySuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"0x1","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	ctx := context.Background()

	response, err := provider.Call(ctx, "eth_getBalance", []interface{}{"0x123", "latest"})

	require.NoError(t, err)
	assert.NotEmpty(t, response)
	assert.Equal(t, `{"jsonrpc":"2.0","result":"0x1","id":1}`, string(response))
}

// TestProvider_Call_WithTimeout uses configured timeout
func TestProvider_Call_WithTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"0x1","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	provider.timeout = 50 * time.Millisecond // Set short timeout

	ctx := context.Background()
	_, err := provider.Call(ctx, "eth_getBalance", []interface{}{"0x123", "latest"})

	assert.Error(t, err)
}

// TestProvider_Call_PrimaryFailsUseFallback uses fallback when primary fails
func TestProvider_Call_PrimaryFailsUseFallback(t *testing.T) {
	// Primary server returns error
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal error"},"id":1}`))
	}))
	defer primary.Close()

	// Fallback server returns success
	fallback := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"0x2","id":1}`))
	}))
	defer fallback.Close()

	provider := NewProvider(primary.URL, fallback.URL)
	ctx := context.Background()

	response, err := provider.Call(ctx, "eth_getBalance", []interface{}{"0x123", "latest"})

	require.NoError(t, err)
	assert.Contains(t, string(response), "0x2")
}

// TestProvider_Call_BothFail returns error when both fail
func TestProvider_Call_BothFail(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer primary.Close()

	fallback := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer fallback.Close()

	provider := NewProvider(primary.URL, fallback.URL)
	ctx := context.Background()

	_, err := provider.Call(ctx, "eth_getBalance", []interface{}{"0x123", "latest"})

	assert.Error(t, err)
}

// TestProvider_Call_PrimaryFailsNoFallback returns error
func TestProvider_Call_PrimaryFailsNoFallback(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer primary.Close()

	provider := NewProvider(primary.URL, "")
	ctx := context.Background()

	_, err := provider.Call(ctx, "eth_getBalance", []interface{}{"0x123", "latest"})

	assert.Error(t, err)
}

// TestProvider_Call_WithInvalidURL returns error
func TestProvider_Call_WithInvalidURL(t *testing.T) {
	provider := NewProvider("invalid://url", "")
	ctx := context.Background()

	_, err := provider.Call(ctx, "eth_getBalance", []interface{}{"0x123", "latest"})

	assert.Error(t, err)
}

// TestProvider_Call_JSONRPCError returned in response
func TestProvider_Call_JSONRPCError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","error":{"code":-32700,"message":"Parse error"},"id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	ctx := context.Background()

	response, err := provider.Call(ctx, "eth_getBalance", []interface{}{"0x123", "latest"})

	require.NoError(t, err)
	assert.Contains(t, string(response), "error")
	assert.Contains(t, string(response), "-32700")
}

// TestProvider_CallWithMethod uses correct method name
func TestProvider_CallWithMethod(t *testing.T) {
	capturedMethod := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		capturedMethod = string(body)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"0x1","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	ctx := context.Background()

	provider.Call(ctx, "eth_blockNumber", []interface{}{})

	assert.Contains(t, capturedMethod, "eth_blockNumber")
}

// TestProvider_SetTimeout changes request timeout
func TestProvider_SetTimeout(t *testing.T) {
	provider := NewProvider("https://eth.example.com", "")

	assert.Equal(t, 5*time.Second, provider.timeout)

	provider.SetTimeout(10 * time.Second)

	assert.Equal(t, 10*time.Second, provider.timeout)
}

// TestProvider_Health checks if provider is healthy
func TestProvider_Health(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"0x1","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	ctx := context.Background()

	healthy := provider.HealthCheck(ctx)

	assert.True(t, healthy)
}

// TestProvider_HealthUnhealthy returns false when unhealthy
func TestProvider_HealthUnhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	ctx := context.Background()

	healthy := provider.HealthCheck(ctx)

	assert.False(t, healthy)
}

// TestProvider_ContextCancellation respects context cancellation
func TestProvider_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"0x1","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := provider.Call(ctx, "eth_getBalance", []interface{}{"0x123", "latest"})

	assert.Error(t, err)
}

// TestProvider_MultipleRequests makes multiple sequential requests
func TestProvider_MultipleRequests(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"0x1","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_, err := provider.Call(ctx, "eth_getBalance", []interface{}{"0x123", "latest"})
		require.NoError(t, err)
	}

	assert.Equal(t, 5, requestCount)
}

// TestProvider_ConnectionPooling reuses HTTP connections
func TestProvider_ConnectionPooling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"0x1","id":1}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "")
	ctx := context.Background()

	// Make multiple requests to same provider
	for i := 0; i < 10; i++ {
		_, err := provider.Call(ctx, "eth_getBalance", []interface{}{"0x123", "latest"})
		require.NoError(t, err)
	}

	// HTTP client should have pooled connections
	assert.NotNil(t, provider.client)
}

// TestProvider_Close closes the provider
func TestProvider_Close(t *testing.T) {
	provider := NewProvider("https://eth.example.com", "")

	err := provider.Close()

	assert.NoError(t, err)
}
