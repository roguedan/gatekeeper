package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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

// SIWEMessage represents a parsed SIWE message
type SIWEMessage struct {
	Domain    string
	Address   string
	Statement string
	URI       string
	Version   string
	ChainID   string
	Nonce     string
	IssuedAt  string
}

// ParseSIWEMessage parses a SIWE message string into its components
func ParseSIWEMessage(message string) (*SIWEMessage, error) {
	lines := strings.Split(strings.TrimSpace(message), "\n")
	if len(lines) < 4 {
		return nil, fmt.Errorf("invalid SIWE message format: too few lines")
	}

	// First line should be the domain
	domain := strings.TrimSpace(lines[0])
	if domain == "" {
		return nil, fmt.Errorf("invalid SIWE message: missing domain")
	}

	// Find the address line (should contain "wants you to sign in")
	var address string
	var startIdx int
	for i := 1; i < len(lines); i++ {
		if strings.Contains(lines[i], "wants you to sign in") {
			// Extract address from line before this
			if i > 0 {
				address = strings.TrimSpace(lines[i-1])
				startIdx = i + 1
				break
			}
		}
	}

	if address == "" {
		return nil, fmt.Errorf("invalid SIWE message: missing address")
	}

	// Validate address format
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("invalid ethereum address: %s", address)
	}

	msg := &SIWEMessage{
		Domain:  domain,
		Address: strings.ToLower(address),
	}

	// Parse remaining fields
	for i := startIdx; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Check for field: value format
		if strings.HasPrefix(line, "URI:") {
			msg.URI = strings.TrimSpace(strings.TrimPrefix(line, "URI:"))
		} else if strings.HasPrefix(line, "Version:") {
			msg.Version = strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
		} else if strings.HasPrefix(line, "Chain ID:") {
			msg.ChainID = strings.TrimSpace(strings.TrimPrefix(line, "Chain ID:"))
		} else if strings.HasPrefix(line, "Nonce:") {
			msg.Nonce = strings.TrimSpace(strings.TrimPrefix(line, "Nonce:"))
		} else if strings.HasPrefix(line, "Issued At:") {
			msg.IssuedAt = strings.TrimSpace(strings.TrimPrefix(line, "Issued At:"))
		} else if !strings.Contains(line, ":") {
			// This might be a statement line
			if msg.Statement == "" {
				msg.Statement = line
			}
		}
	}

	return msg, nil
}

// ExtractNonceFromMessage extracts the nonce from a SIWE message
func ExtractNonceFromMessage(message string) (string, error) {
	nonceRegex := regexp.MustCompile(`(?m)^Nonce:\s*(.+)$`)
	matches := nonceRegex.FindStringSubmatch(message)
	if len(matches) < 2 {
		return "", fmt.Errorf("nonce not found in message")
	}
	return strings.TrimSpace(matches[1]), nil
}

// ExtractAddressFromMessage extracts the Ethereum address from a SIWE message
func ExtractAddressFromMessage(message string) (string, error) {
	// Look for hex address pattern (0x followed by 40 hex chars)
	addressRegex := regexp.MustCompile(`(?i)(0x[a-fA-F0-9]{40})`)
	matches := addressRegex.FindStringSubmatch(message)
	if len(matches) < 1 {
		return "", fmt.Errorf("address not found in message")
	}
	return strings.ToLower(matches[0]), nil
}

// VerifySignature verifies that the signature was created by signing the message with the private key corresponding to the address
func VerifySignature(message, signature, address string) (bool, error) {
	// Normalize address
	if !common.IsHexAddress(address) {
		return false, fmt.Errorf("invalid ethereum address: %s", address)
	}
	expectedAddr := common.HexToAddress(address)

	// Decode signature
	sigBytes, err := hexutil.Decode(signature)
	if err != nil {
		return false, fmt.Errorf("invalid signature format: %w", err)
	}

	// Signature should be 65 bytes
	if len(sigBytes) != 65 {
		return false, fmt.Errorf("invalid signature length: expected 65 bytes, got %d", len(sigBytes))
	}

	// Ethereum uses v = 27 or 28, but go-ethereum expects v = 0 or 1
	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27
	}

	// Hash the message using Ethereum's personal_sign format
	messageHash := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))

	// Recover the public key from the signature
	pubKey, err := crypto.SigToPub(messageHash.Bytes(), sigBytes)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key: %w", err)
	}

	// Get the address from the public key
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	// Compare addresses
	if recoveredAddr != expectedAddr {
		return false, fmt.Errorf("signature verification failed: expected %s, got %s", expectedAddr.Hex(), recoveredAddr.Hex())
	}

	return true, nil
}
