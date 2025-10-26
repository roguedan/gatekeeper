package chain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Provider manages Ethereum RPC connections with primary and fallback support
type Provider struct {
	primaryURL  string
	fallbackURL string
	client      *http.Client
	timeout     time.Duration
}

// jsonRPCRequest represents a JSON-RPC 2.0 request
type jsonRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// jsonRPCResponse represents a JSON-RPC 2.0 response
type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
	ID      int             `json:"id"`
}

// jsonRPCError represents a JSON-RPC error
type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewProvider creates a new Ethereum RPC provider with primary and optional fallback
func NewProvider(primaryURL string, fallbackURL string) *Provider {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &Provider{
		primaryURL:  primaryURL,
		fallbackURL: fallbackURL,
		client:      client,
		timeout:     5 * time.Second,
	}
}

// Call makes a JSON-RPC call to the primary provider, with fallback support
func (p *Provider) Call(ctx context.Context, method string, params []interface{}) ([]byte, error) {
	// Try primary provider
	response, err := p.callProvider(ctx, p.primaryURL, method, params)
	if err == nil {
		return response, nil
	}

	// If primary failed and we have a fallback, try it
	if p.fallbackURL != "" {
		response, fallbackErr := p.callProvider(ctx, p.fallbackURL, method, params)
		if fallbackErr == nil {
			return response, nil
		}
		// If fallback also failed, return the original error from primary
	}

	return nil, err
}

// callProvider makes a JSON-RPC call to a specific provider URL
func (p *Provider) callProvider(ctx context.Context, url string, method string, params []interface{}) ([]byte, error) {
	// Create JSON-RPC request
	request := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	// Marshal to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	// Create HTTP request with context and timeout
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	httpReq = httpReq.WithContext(ctx)

	// Make request
	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("RPC call failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if httpResp.StatusCode >= 400 {
		return nil, fmt.Errorf("RPC server returned HTTP %d: %s", httpResp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

// SetTimeout sets the request timeout for RPC calls
func (p *Provider) SetTimeout(timeout time.Duration) {
	p.timeout = timeout
	p.client.Timeout = timeout
}

// HealthCheck verifies the provider is healthy by calling eth_blockNumber
func (p *Provider) HealthCheck(ctx context.Context) bool {
	_, err := p.Call(ctx, "eth_blockNumber", []interface{}{})
	return err == nil
}

// Close closes the provider and its connections
func (p *Provider) Close() error {
	p.client.CloseIdleConnections()
	return nil
}
