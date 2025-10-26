package http

import (
	"encoding/json"
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
	Nonce     string `json:"nonce"`
	Message   string `json:"message"`
	Signature string `json:"signature"`
	Address   string `json:"address"`
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
	if req.Nonce == "" || req.Message == "" || req.Signature == "" || req.Address == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	// Verify nonce exists and is valid
	valid, err := h.siweService.VerifyNonce(ctx, req.Nonce)
	if err != nil || !valid {
		http.Error(w, "invalid or expired nonce", http.StatusUnauthorized)
		return
	}

	// TODO: Verify SIWE signature using spruceid/siwe library
	// For now, we accept any signature
	// In production, this should verify the signature against the message and address

	// Invalidate nonce to prevent replay attacks
	h.siweService.InvalidateNonce(ctx, req.Nonce)

	// Generate JWT token with auth scope
	token, err := h.jwtService.GenerateToken(ctx, req.Address, []string{"auth"})
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := VerifyResponse{
		Token:   token,
		Address: req.Address,
	}
	json.NewEncoder(w).Encode(response)
}
