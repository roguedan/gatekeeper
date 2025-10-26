package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/yourusername/gatekeeper/internal/auth"
)

// claimsContextKey is the key used to store JWT claims in the request context
type contextKey string

const claimsContextKey contextKey = "jwt_claims"

// Middleware is a function that wraps an HTTP handler
type Middleware func(http.Handler) http.Handler

// JWTMiddleware creates a middleware that validates JWT tokens
func JWTMiddleware(jwtService *auth.JWTService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			// Parse "Bearer <token>" format
			parts := strings.Fields(authHeader)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			// Verify token
			claims, err := jwtService.VerifyToken(r.Context(), token)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			// Add claims to request context
			ctx := context.WithValue(r.Context(), claimsContextKey, claims)
			r = r.WithContext(ctx)

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// ClaimsFromContext extracts JWT claims from request context
func ClaimsFromContext(r *http.Request) *auth.Claims {
	claims, ok := r.Context().Value(claimsContextKey).(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}
