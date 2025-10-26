package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims with custom fields
type Claims struct {
	Address string   `json:"address"`
	Scopes  []string `json:"scopes"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token generation and verification
type JWTService struct {
	secret []byte
	expiry time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secret []byte, expiry time.Duration) *JWTService {
	return &JWTService{
		secret: secret,
		expiry: expiry,
	}
}

// GenerateToken creates a new JWT token for the given address with scopes
func (j *JWTService) GenerateToken(ctx context.Context, address string, scopes []string) (string, error) {
	now := time.Now()
	claims := Claims{
		Address: address,
		Scopes:  scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expiry)),
			Issuer:    "gatekeeper",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// VerifyToken verifies and parses a JWT token
func (j *JWTService) VerifyToken(ctx context.Context, tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token is empty")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && time.Now().Unix() > claims.ExpiresAt.Unix() {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}
