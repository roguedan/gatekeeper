package http

import (
	"testing"
	"time"
)

func TestNewInMemoryRateLimiter(t *testing.T) {
	tests := []struct {
		name              string
		requestsPerWindow int
		window            time.Duration
		burst             int
		wantErr           bool
	}{
		{
			name:              "valid configuration",
			requestsPerWindow: 10,
			window:            time.Hour,
			burst:             3,
			wantErr:           false,
		},
		{
			name:              "high rate limit",
			requestsPerWindow: 1000,
			window:            time.Minute,
			burst:             100,
			wantErr:           false,
		},
		{
			name:              "low rate limit",
			requestsPerWindow: 1,
			window:            time.Minute,
			burst:             1,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewInMemoryRateLimiter(tt.requestsPerWindow, tt.window, tt.burst)
			if limiter == nil {
				t.Fatal("expected non-nil limiter")
			}

			// Verify burst is set correctly
			if limiter.Burst() != tt.burst {
				t.Errorf("expected burst %d, got %d", tt.burst, limiter.Burst())
			}
		})
	}
}

func TestInMemoryRateLimiter_Allow(t *testing.T) {
	tests := []struct {
		name       string
		limit      int
		window     time.Duration
		burst      int
		requests   int
		identifier string
		wantAllow  []bool
	}{
		{
			name:       "all requests within burst",
			limit:      10,
			window:     time.Hour,
			burst:      5,
			requests:   3,
			identifier: "user:0x123",
			wantAllow:  []bool{true, true, true},
		},
		{
			name:       "exceed burst limit",
			limit:      10,
			window:     time.Hour,
			burst:      3,
			requests:   5,
			identifier: "user:0x456",
			wantAllow:  []bool{true, true, true, false, false},
		},
		{
			name:       "single request allowed",
			limit:      1,
			window:     time.Minute,
			burst:      1,
			requests:   1,
			identifier: "user:0x789",
			wantAllow:  []bool{true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewInMemoryRateLimiter(tt.limit, tt.window, tt.burst)

			for i := 0; i < tt.requests; i++ {
				allowed := limiter.Allow(tt.identifier)
				if allowed != tt.wantAllow[i] {
					t.Errorf("request %d: expected Allow()=%v, got %v", i, tt.wantAllow[i], allowed)
				}
			}
		})
	}
}

func TestInMemoryRateLimiter_AllowN(t *testing.T) {
	tests := []struct {
		name       string
		limit      int
		window     time.Duration
		burst      int
		n          int
		identifier string
		wantAllow  bool
	}{
		{
			name:       "allow batch within burst",
			limit:      10,
			window:     time.Hour,
			burst:      5,
			n:          3,
			identifier: "user:0x123",
			wantAllow:  true,
		},
		{
			name:       "deny batch exceeding burst",
			limit:      10,
			window:     time.Hour,
			burst:      3,
			n:          5,
			identifier: "user:0x456",
			wantAllow:  false,
		},
		{
			name:       "allow exact burst amount",
			limit:      10,
			window:     time.Hour,
			burst:      5,
			n:          5,
			identifier: "user:0x789",
			wantAllow:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewInMemoryRateLimiter(tt.limit, tt.window, tt.burst)

			allowed := limiter.AllowN(tt.identifier, tt.n)
			if allowed != tt.wantAllow {
				t.Errorf("expected AllowN()=%v, got %v", tt.wantAllow, allowed)
			}
		})
	}
}

func TestInMemoryRateLimiter_PerUserLimiting(t *testing.T) {
	limiter := NewInMemoryRateLimiter(10, time.Hour, 3)

	// User 1 makes requests
	user1 := "user:0x111"
	for i := 0; i < 3; i++ {
		if !limiter.Allow(user1) {
			t.Errorf("user1 request %d should be allowed", i)
		}
	}
	// 4th request should be denied
	if limiter.Allow(user1) {
		t.Error("user1 request 4 should be denied (exceeded burst)")
	}

	// User 2 should have independent limit
	user2 := "user:0x222"
	for i := 0; i < 3; i++ {
		if !limiter.Allow(user2) {
			t.Errorf("user2 request %d should be allowed", i)
		}
	}
	// 4th request should be denied
	if limiter.Allow(user2) {
		t.Error("user2 request 4 should be denied (exceeded burst)")
	}
}

func TestInMemoryRateLimiter_Reset(t *testing.T) {
	limiter := NewInMemoryRateLimiter(10, time.Hour, 3)

	identifier := "user:0x123"

	// Exhaust the rate limit
	for i := 0; i < 3; i++ {
		limiter.Allow(identifier)
	}

	// Should be rate limited
	if limiter.Allow(identifier) {
		t.Error("should be rate limited before reset")
	}

	// Reset the limiter
	limiter.Reset(identifier)

	// Should be allowed again
	if !limiter.Allow(identifier) {
		t.Error("should be allowed after reset")
	}
}

func TestInMemoryRateLimiter_Clear(t *testing.T) {
	limiter := NewInMemoryRateLimiter(10, time.Hour, 3)

	// Create entries for multiple users
	users := []string{"user:0x111", "user:0x222", "user:0x333"}
	for _, user := range users {
		limiter.Allow(user)
	}

	// Verify entries exist
	limiter.mu.RLock()
	if len(limiter.limiters) != len(users) {
		t.Errorf("expected %d limiters, got %d", len(users), len(limiter.limiters))
	}
	limiter.mu.RUnlock()

	// Clear should not remove recently accessed entries
	limiter.Clear()

	limiter.mu.RLock()
	remaining := len(limiter.limiters)
	limiter.mu.RUnlock()

	if remaining != len(users) {
		t.Errorf("Clear should not remove recent entries, expected %d, got %d", len(users), remaining)
	}
}

func TestRateLimitConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  RateLimitConfig
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultRateLimitConfig(),
			wantErr: false,
		},
		{
			name: "invalid API key creation limit",
			config: RateLimitConfig{
				APIKeyCreationLimit: 0,
				APIKeyCreationBurst: 3,
				APIKeyWindow:        time.Hour,
				APIUsageLimit:       1000,
				APIUsageBurst:       100,
				APIUsageWindow:      time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid burst limit",
			config: RateLimitConfig{
				APIKeyCreationLimit: 10,
				APIKeyCreationBurst: 0,
				APIKeyWindow:        time.Hour,
				APIUsageLimit:       1000,
				APIUsageBurst:       100,
				APIUsageWindow:      time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid window",
			config: RateLimitConfig{
				APIKeyCreationLimit: 10,
				APIKeyCreationBurst: 3,
				APIKeyWindow:        0,
				APIUsageLimit:       1000,
				APIUsageBurst:       100,
				APIUsageWindow:      time.Minute,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInMemoryRateLimiter_BurstBehavior(t *testing.T) {
	// Test that burst allows temporary spikes above the rate
	limiter := NewInMemoryRateLimiter(60, time.Minute, 10) // 60/min = 1/sec, burst 10

	identifier := "user:0x123"

	// Should allow burst of 10 requests immediately
	for i := 0; i < 10; i++ {
		if !limiter.Allow(identifier) {
			t.Errorf("burst request %d should be allowed", i)
		}
	}

	// 11th request should be denied (burst exhausted)
	if limiter.Allow(identifier) {
		t.Error("request after burst should be denied")
	}

	// Wait for token to refill (1 second = 1 token at 60/min rate)
	time.Sleep(1100 * time.Millisecond)

	// Should allow 1 more request after refill
	if !limiter.Allow(identifier) {
		t.Error("request after refill should be allowed")
	}
}

func TestInMemoryRateLimiter_ConcurrentAccess(t *testing.T) {
	limiter := NewInMemoryRateLimiter(100, time.Minute, 50)

	// Test concurrent access from multiple goroutines
	done := make(chan bool)
	const goroutines = 10
	const requestsPerGoroutine = 5

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			identifier := "user:0x" + string(rune('0'+id))
			for j := 0; j < requestsPerGoroutine; j++ {
				limiter.Allow(identifier)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Verify no panics occurred and limiters were created
	limiter.mu.RLock()
	count := len(limiter.limiters)
	limiter.mu.RUnlock()

	if count == 0 {
		t.Error("expected limiters to be created")
	}
}
