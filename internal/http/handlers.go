package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourusername/gatekeeper/internal/auth"
)

// AuthHandler handles authentication-related HTTP endpoints
type AuthHandler struct {
	siweService *auth.SIWEService
	jwtService  *auth.JWTService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(siweService *auth.SIWEService, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		siweService: siweService,
		jwtService:  jwtService,
	}
}

// GetNonce handles GET /auth/siwe/nonce
// Returns a new nonce for SIWE message signing
func (h *AuthHandler) GetNonce(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Generate new nonce
	nonce, err := h.siweService.GenerateNonce(ctx)
	if err != nil {
		http.Error(w, "failed to generate nonce", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Return nonce
	response := map[string]string{
		"nonce": nonce,
	}
	json.NewEncoder(w).Encode(response)
}

// VerifyRequest represents the request body for verification
type VerifyRequest struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

// VerifyResponse represents the response for successful verification
type VerifyResponse struct {
	Token   string `json:"token"`
	Address string `json:"address"`
}

// VerifySIWE handles POST /auth/siwe/verify
// Verifies SIWE message signature and returns JWT token
func (h *AuthHandler) VerifySIWE(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Message == "" || req.Signature == "" {
		http.Error(w, "missing required fields: message and signature", http.StatusBadRequest)
		return
	}

	// Extract nonce from message
	nonce, err := auth.ExtractNonceFromMessage(req.Message)
	if err != nil {
		http.Error(w, "invalid SIWE message: nonce not found", http.StatusBadRequest)
		return
	}

	// Verify nonce exists and is valid
	valid, err := h.siweService.VerifyNonce(ctx, nonce)
	if err != nil || !valid {
		http.Error(w, "invalid or expired nonce", http.StatusUnauthorized)
		return
	}

	// Extract address from message
	address, err := auth.ExtractAddressFromMessage(req.Message)
	if err != nil {
		http.Error(w, "invalid SIWE message: address not found", http.StatusBadRequest)
		return
	}

	// Verify signature matches message + address
	validSig, err := auth.VerifySignature(req.Message, req.Signature, address)
	if err != nil {
		http.Error(w, fmt.Sprintf("signature verification error: %v", err), http.StatusBadRequest)
		return
	}
	if !validSig {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// Invalidate nonce to prevent replay attacks
	if err := h.siweService.InvalidateNonce(ctx, nonce); err != nil {
		// Log error but continue - nonce validation already passed
		// This shouldn't normally happen
	}

	// Generate JWT token with auth scope
	token, err := h.jwtService.GenerateToken(ctx, address, []string{"auth"})
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := VerifyResponse{
		Token:   token,
		Address: address,
	}
	json.NewEncoder(w).Encode(response)
}
