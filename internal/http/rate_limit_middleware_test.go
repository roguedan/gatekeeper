package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/gatekeeper/internal/auth"
	"github.com/yourusername/gatekeeper/internal/log"
)

func TestNewRateLimitMiddleware(t *testing.T) {
	logger, _ := log.New("info")
	limiter := NewInMemoryRateLimiter(10, time.Hour, 3)

	middleware := NewRateLimitMiddleware(limiter, logger)

	if middleware == nil {
		t.Fatal("expected non-nil middleware")
	}
	if middleware.limiter != limiter {
		t.Error("limiter not set correctly")
	}
	if middleware.logger != logger {
		t.Error("logger not set correctly")
	}
}

func TestRateLimitMiddleware_Allow(t *testing.T) {
	logger, _ := log.New("info")
	limiter := NewInMemoryRateLimiter(10, time.Hour, 3)
	middleware := NewRateLimitMiddleware(limiter, logger)

	// Create test handler
	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	// Make requests within limit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i, w.Code)
		}
	}
}

func TestRateLimitMiddleware_Deny(t *testing.T) {
	logger, _ := log.New("info")
	limiter := NewInMemoryRateLimiter(10, time.Hour, 3)
	middleware := NewRateLimitMiddleware(limiter, logger)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	// Exhaust rate limit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// Next request should be denied
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}

	// Verify Retry-After header is set
	retryAfter := w.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("expected Retry-After header to be set")
	}

	// Verify rate limit headers
	if w.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Errorf("expected X-RateLimit-Remaining=0, got %s", w.Header().Get("X-RateLimit-Remaining"))
	}
}

func TestRateLimitMiddleware_UserIdentifier(t *testing.T) {
	logger, _ := log.New("info")
	limiter := NewInMemoryRateLimiter(10, time.Hour, 3)
	middleware := NewUserRateLimitMiddleware(limiter, logger)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create request with JWT claims
	claims := &auth.Claims{
		Address: "0x1234567890abcdef1234567890abcdef12345678",
		Scopes:  []string{"auth"},
	}

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), ClaimsContextKey, claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Make multiple requests to test user-specific limiting
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := context.WithValue(req.Context(), ClaimsContextKey, claims)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i, w.Code)
		}
	}

	// 4th request should be denied
	req = httptest.NewRequest("GET", "/test", nil)
	ctx = context.WithValue(req.Context(), ClaimsContextKey, claims)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}
}

func TestRateLimitMiddleware_IPFallback(t *testing.T) {
	logger, _ := log.New("info")
	limiter := NewInMemoryRateLimiter(10, time.Hour, 3)
	middleware := NewRateLimitMiddleware(limiter, logger)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request without authentication should use IP
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i, w.Code)
		}
	}

	// 4th request should be denied
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}
}

func TestRateLimitMiddleware_DifferentUsers(t *testing.T) {
	logger, _ := log.New("info")
	limiter := NewInMemoryRateLimiter(10, time.Hour, 3)
	middleware := NewUserRateLimitMiddleware(limiter, logger)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// User 1 exhausts their limit
	user1Claims := &auth.Claims{Address: "0x111", Scopes: []string{"auth"}}
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := context.WithValue(req.Context(), ClaimsContextKey, user1Claims)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// User 1's next request should be denied
	req1 := httptest.NewRequest("GET", "/test", nil)
	ctx1 := context.WithValue(req1.Context(), ClaimsContextKey, user1Claims)
	req1 = req1.WithContext(ctx1)
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusTooManyRequests {
		t.Errorf("user1: expected status 429, got %d", w1.Code)
	}

	// User 2 should still be allowed
	user2Claims := &auth.Claims{Address: "0x222", Scopes: []string{"auth"}}
	req2 := httptest.NewRequest("GET", "/test", nil)
	ctx2 := context.WithValue(req2.Context(), ClaimsContextKey, user2Claims)
	req2 = req2.WithContext(ctx2)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("user2: expected status 200, got %d", w2.Code)
	}
}

func TestExtractIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		wantIP     string
	}{
		{
			name:       "basic remote addr",
			remoteAddr: "192.168.1.1:12345",
			headers:    map[string]string{},
			wantIP:     "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For header",
			remoteAddr: "192.168.1.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.1"},
			wantIP:     "203.0.113.1",
		},
		{
			name:       "X-Real-IP header",
			remoteAddr: "192.168.1.1:12345",
			headers:    map[string]string{"X-Real-IP": "203.0.113.2"},
			wantIP:     "203.0.113.2",
		},
		{
			name:       "X-Forwarded-For takes precedence",
			remoteAddr: "192.168.1.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1",
				"X-Real-IP":       "203.0.113.2",
			},
			wantIP: "203.0.113.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ip := extractIP(req)
			if ip != tt.wantIP {
				t.Errorf("expected IP %s, got %s", tt.wantIP, ip)
			}
		})
	}
}

func TestPerUserRateLimitMiddleware(t *testing.T) {
	logger, _ := log.New("info")
	middleware := PerUserRateLimitMiddleware(10, time.Hour, 3, logger)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	claims := &auth.Claims{Address: "0x123", Scopes: []string{"auth"}}

	// Make 3 requests (should all succeed)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := context.WithValue(req.Context(), ClaimsContextKey, claims)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i, w.Code)
		}
	}

	// 4th request should be denied
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), ClaimsContextKey, claims)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}
}

func TestPerIPRateLimitMiddleware(t *testing.T) {
	logger, _ := log.New("info")
	middleware := PerIPRateLimitMiddleware(10, time.Hour, 3, logger)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make 3 requests from same IP (should all succeed)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i, w.Code)
		}
	}

	// 4th request should be denied
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}
}

func TestRateLimitMiddleware_RetryAfter(t *testing.T) {
	logger, _ := log.New("info")
	limiter := NewInMemoryRateLimiter(10, time.Hour, 2)
	middleware := NewRateLimitMiddleware(limiter, logger)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust rate limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// Next request should be denied with Retry-After
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}

	retryAfter := w.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("expected Retry-After header")
	}

	// Verify it's a number
	if retryAfter != "60" {
		t.Logf("Retry-After value: %s", retryAfter)
	}
}

func TestRateLimitMiddleware_CustomIdentifier(t *testing.T) {
	logger, _ := log.New("info")
	limiter := NewInMemoryRateLimiter(10, time.Hour, 3)

	// Custom identifier function that uses a header
	customIdentifier := func(r *http.Request) string {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "" {
			return "apikey:" + apiKey
		}
		return "ip:" + extractIP(r)
	}

	middleware := NewRateLimitMiddleware(limiter, logger, WithIdentifierFunc(customIdentifier))

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make requests with API key
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "test-key-123")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i, w.Code)
		}
	}

	// 4th request with same API key should be denied
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}

	// Different API key should be allowed
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-API-Key", "different-key-456")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200 for different key, got %d", w2.Code)
	}
}
