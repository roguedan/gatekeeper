package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// NonceInfo contains metadata about a nonce
type NonceInfo struct {
	Nonce     string
	CreatedAt time.Time
	ExpiresAt time.Time
	Used      bool
	UsedAt    *time.Time
}

// nonceData holds internal nonce state
type nonceData struct {
	createdAt time.Time
	expiresAt time.Time
	used      bool
	usedAt    *time.Time
}

// SIWEService manages SIWE nonce generation and verification
type SIWEService struct {
	nonces  map[string]*nonceData
	ttl     time.Duration
	mu      sync.RWMutex
}

// NewSIWEService creates a new SIWE service with the specified nonce TTL
func NewSIWEService(ttl time.Duration) *SIWEService {
	return &SIWEService{
		nonces: make(map[string]*nonceData),
		ttl:    ttl,
	}
}

// GenerateNonce creates a new cryptographically random nonce
func (s *SIWEService) GenerateNonce(ctx context.Context) (string, error) {
	// Generate 16 random bytes (128 bits of entropy)
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	nonce := hex.EncodeToString(bytes)

	// Store nonce with expiry
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	s.nonces[nonce] = &nonceData{
		createdAt: now,
		expiresAt: now.Add(s.ttl),
		used:      false,
	}

	return nonce, nil
}

// VerifyNonce checks if a nonce is valid (exists and not expired)
func (s *SIWEService) VerifyNonce(ctx context.Context, nonce string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.nonces[nonce]
	if !exists {
		return false, nil
	}

	// Check if nonce has expired
	if time.Now().After(data.expiresAt) {
		return false, nil
	}

	// Check if nonce has been used
	if data.used {
		return false, nil
	}

	return true, nil
}

// InvalidateNonce marks a nonce as used
func (s *SIWEService) InvalidateNonce(ctx context.Context, nonce string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.nonces[nonce]
	if !exists {
		return fmt.Errorf("nonce not found: %s", nonce)
	}

	now := time.Now()
	data.used = true
	data.usedAt = &now

	return nil
}

// CleanupExpiredNonces removes expired nonces from storage
func (s *SIWEService) CleanupExpiredNonces(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for nonce, data := range s.nonces {
		if now.After(data.expiresAt) {
			delete(s.nonces, nonce)
		}
	}

	return nil
}

// GetNonceInfo returns metadata about a nonce
func (s *SIWEService) GetNonceInfo(ctx context.Context, nonce string) (*NonceInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.nonces[nonce]
	if !exists {
		return nil, fmt.Errorf("nonce not found: %s", nonce)
	}

	return &NonceInfo{
		Nonce:     nonce,
		CreatedAt: data.createdAt,
		ExpiresAt: data.expiresAt,
		Used:      data.used,
		UsedAt:    data.usedAt,
	}, nil
}
