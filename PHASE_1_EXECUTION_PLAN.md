# Phase 1 Execution Plan: Project Setup & Core Authentication

## Overview
Phase 1 focuses on project scaffolding, database setup, and core SIWE authentication with TDD-first approach.

**Duration**: Week 1
**Goal**: Running backend with working SIWE auth, JWT tokens, and documented APIs

---

## Task Breakdown with Skill Mapping & TDD Approach

### Block 1: Project Structure & Configuration

#### Task 1.1: Initialize Go Module and Project Layout
**Skills Used**:
- `go-backend-development` - Standard project structure
- `openspec-spec-driven-development` - Reference specification during setup

**TDD Approach**: Not applicable (setup task)

**Implementation Steps**:
```bash
# 1. Initialize module
go mod init github.com/yourusername/gatekeeper

# 2. Create directory structure
mkdir -p cmd/server internal/{http,auth,store,config,log,chain,policy}
mkdir -p api web/src deployments .github/workflows

# 3. Create core files
touch cmd/server/main.go
touch internal/{config,log}/init.go
touch Makefile README.md .gitignore
```

**Deliverable**: Standard Go project layout following conventions

---

#### Task 1.2: Set up Configuration Management
**Skills Used**:
- `go-backend-development` - Config patterns and best practices
- `openspec-spec-driven-development` - Reference environment requirements

**TDD Approach**: Write config validation tests BEFORE implementation

**Step 1 (🔴 RED): Write failing test**
```go
// internal/config/config_test.go
package config

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestLoad_AllRequiredFieldsPresent(t *testing.T) {
    t.Setenv("PORT", "8080")
    t.Setenv("DATABASE_URL", "postgres://localhost/gatekeeper")
    t.Setenv("JWT_SECRET", "test-secret-key")
    t.Setenv("ETHEREUM_RPC", "https://eth.example.com")

    cfg, err := Load()

    require.NoError(t, err)
    assert.Equal(t, "8080", cfg.Port)
    assert.Equal(t, "postgres://localhost/gatekeeper", cfg.DatabaseURL)
}

func TestLoad_MissingRequiredField(t *testing.T) {
    // Don't set DATABASE_URL
    t.Setenv("JWT_SECRET", "test-secret")
    t.Setenv("ETHEREUM_RPC", "https://eth.example.com")

    cfg, err := Load()

    require.Error(t, err)
    assert.Nil(t, cfg)
}
```

**Step 2 (🟢 GREEN): Implement config**
```go
// internal/config/config.go
package config

import (
    "fmt"
    "os"
    "time"
)

type Config struct {
    Port            string
    DatabaseURL     string
    JWTSecret       []byte
    JWTExpiry       time.Duration
    EthereumRPC     string
    CacheEnabled    bool
    CacheTTL        time.Duration
    Environment     string
}

func Load() (*Config, error) {
    cfg := &Config{
        Port:            os.Getenv("PORT"),
        DatabaseURL:    os.Getenv("DATABASE_URL"),
        JWTSecret:      []byte(os.Getenv("JWT_SECRET")),
        JWTExpiry:      24 * time.Hour,
        EthereumRPC:    os.Getenv("ETHEREUM_RPC"),
        Environment:    os.Getenv("ENVIRONMENT"),
    }

    if err := cfg.Validate(); err != nil {
        return nil, err
    }

    return cfg, nil
}

func (c *Config) Validate() error {
    if c.DatabaseURL == "" {
        return fmt.Errorf("DATABASE_URL is required")
    }
    if len(c.JWTSecret) == 0 {
        return fmt.Errorf("JWT_SECRET is required")
    }
    if c.EthereumRPC == "" {
        return fmt.Errorf("ETHEREUM_RPC is required")
    }
    return nil
}
```

**Test**: `go test ./internal/config -v`

**Deliverable**: Config package with validation, passing tests

---

### Block 2: Database & Logging Setup

#### Task 2.1: Database Connection & Pooling
**Skills Used**:
- `go-backend-development` - Connection pooling patterns
- `test-driven-development` - Mock database in tests

**TDD Approach**: Test connection and health checks

**Implementation**:
```go
// internal/store/db.go
package store

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    _ "github.com/lib/pq"
)

func NewDB(connString string) (*sql.DB, error) {
    db, err := sql.Open("postgres", connString)
    if err != nil {
        return nil, fmt.Errorf("open db: %w", err)
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("ping db: %w", err)
    }

    return db, nil
}
```

**Deliverable**: Database connection with pooling

---

#### Task 2.2: Setup Database Migrations
**Skills Used**:
- `go-backend-development` - Migration patterns
- `technical-documentation` - Document migration process

**TDD Approach**: Not applicable (infrastructure task)

**Implementation**:
```bash
# 1. Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 2. Create migrations directory
mkdir -p internal/store/migrations

# 3. Create first migration
migrate create -ext sql -dir internal/store/migrations -seq init_schema
```

**Create migration file** `internal/store/migrations/000001_init_schema.up.sql`:
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    address TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE nonces (
    nonce TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL
);

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL,
    scopes TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP
);

CREATE INDEX idx_users_address ON users(address);
CREATE INDEX idx_nonces_expires_at ON nonces(expires_at);
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
```

**Create down migration** `internal/store/migrations/000001_init_schema.down.sql`:
```sql
DROP INDEX IF EXISTS idx_api_keys_user_id;
DROP INDEX IF EXISTS idx_nonces_expires_at;
DROP INDEX IF EXISTS idx_users_address;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS nonces;
DROP TABLE IF EXISTS users;
```

**Deliverable**: Working database schema with migrations

---

#### Task 2.3: Structured Logging Setup
**Skills Used**:
- `go-backend-development` - Logging patterns with zap
- `test-driven-development` - Log assertion in tests

**TDD Approach**: Test that logs are properly formatted

**Implementation**:
```go
// internal/log/log.go
package log

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func NewLogger(env string) (*zap.Logger, error) {
    var config zap.Config

    if env == "production" {
        config = zap.NewProductionConfig()
    } else {
        config = zap.NewDevelopmentConfig()
        config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    }

    return config.Build()
}
```

**Deliverable**: Configured zap logger

---

### Block 3: SIWE Authentication (TDD-First)

#### Task 3.1: SIWE Nonce Generation & Management
**Skills Used**:
- `web3-authentication-authorization` - SIWE best practices
- `test-driven-development` - TDD workflow for nonce service
- `go-backend-development` - Go implementation patterns

**TDD Approach**: FULL TDD - Tests drive implementation

**Step 1 (🔴 RED): Write comprehensive tests**
```go
// internal/auth/siwe_test.go
package auth

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestGenerateNonce_CreatesRandomNonce(t *testing.T) {
    service := NewSIWEService()

    nonce1, err := service.GenerateNonce(context.Background())
    require.NoError(t, err)
    assert.NotEmpty(t, nonce1)
    assert.Len(t, nonce1, 32) // 16 bytes = 32 hex chars
}

func TestGenerateNonce_UniqueNonces(t *testing.T) {
    service := NewSIWEService()

    nonce1, _ := service.GenerateNonce(context.Background())
    nonce2, _ := service.GenerateNonce(context.Background())

    assert.NotEqual(t, nonce1, nonce2)
}

func TestVerifyNonce_ValidNonce(t *testing.T) {
    service := NewSIWEService()
    nonce, _ := service.GenerateNonce(context.Background())

    valid := service.VerifyNonce(context.Background(), nonce)

    assert.True(t, valid)
}

func TestVerifyNonce_InvalidNonce(t *testing.T) {
    service := NewSIWEService()

    valid := service.VerifyNonce(context.Background(), "invalid-nonce")

    assert.False(t, valid)
}

func TestVerifyNonce_ConsumeNonce(t *testing.T) {
    service := NewSIWEService()
    nonce, _ := service.GenerateNonce(context.Background())

    // Verify consumes it
    valid := service.VerifyNonce(context.Background(), nonce)
    assert.True(t, valid)

    // Cannot reuse
    validAgain := service.VerifyNonce(context.Background(), nonce)
    assert.False(t, validAgain)
}
```

Run: `go test ./internal/auth -v` → **FAILS**

**Step 2 (🟢 GREEN): Minimal implementation**
```go
// internal/auth/siwe.go
package auth

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "sync"
    "time"
)

type SIWEService struct {
    nonces map[string]time.Time
    mu     sync.RWMutex
}

func NewSIWEService() *SIWEService {
    return &SIWEService{
        nonces: make(map[string]time.Time),
    }
}

func (s *SIWEService) GenerateNonce(ctx context.Context) (string, error) {
    bytes := make([]byte, 16)
    if _, err := rand.Read(bytes); err != nil {
        return "", fmt.Errorf("generate nonce: %w", err)
    }

    nonce := hex.EncodeToString(bytes)

    s.mu.Lock()
    defer s.mu.Unlock()
    s.nonces[nonce] = time.Now().Add(5 * time.Minute)

    return nonce, nil
}

func (s *SIWEService) VerifyNonce(ctx context.Context, nonce string) bool {
    s.mu.Lock()
    defer s.mu.Unlock()

    expiry, exists := s.nonces[nonce]
    if !exists {
        return false
    }

    if time.Now().After(expiry) {
        delete(s.nonces, nonce)
        return false
    }

    // Consume nonce
    delete(s.nonces, nonce)
    return true
}
```

Run: `go test ./internal/auth -v` → **PASSES** ✅

**Step 3 (🔴 RED): Add message verification tests**
```go
func TestVerifyMessage_ValidSignature(t *testing.T) {
    service := NewSIWEService()

    // Create a valid SIWE message and signature (use test fixtures)
    message := "example.com wants you to sign in..."
    nonce, _ := service.GenerateNonce(context.Background())
    message = message + "\nnonce: " + nonce

    // In real test, use actual signed message
    signature := "0x..." // Valid signature

    msg, err := service.VerifyMessage(context.Background(), message, signature)

    require.NoError(t, err)
    assert.NotNil(t, msg)
}

func TestVerifyMessage_InvalidSignature(t *testing.T) {
    service := NewSIWEService()

    message := "example.com wants you to sign in..."
    signature := "0xinvalid"

    msg, err := service.VerifyMessage(context.Background(), message, signature)

    require.Error(t, err)
    assert.Nil(t, msg)
}

func TestVerifyMessage_ExpiredMessage(t *testing.T) {
    service := NewSIWEService()

    // Message with past expiration
    message := "example.com wants you to sign in...\nexpiration_time: 2020-01-01T00:00:00Z"
    signature := "0x..."

    msg, err := service.VerifyMessage(context.Background(), message, signature)

    require.Error(t, err)
    assert.Nil(t, msg)
}
```

**Step 4 (🟢 GREEN): Implement message verification**
```go
// Add to internal/auth/siwe.go
func (s *SIWEService) VerifyMessage(ctx context.Context, message, signature string) (*SIWEMessage, error) {
    // Parse SIWE message
    msg, err := siwe.ParseMessage(message)
    if err != nil {
        return nil, fmt.Errorf("parse message: %w", err)
    }

    // Validate nonce
    if !s.VerifyNonce(ctx, msg.GetNonce()) {
        return nil, fmt.Errorf("invalid or expired nonce")
    }

    // Verify signature cryptographically
    publicKey, err := msg.Verify(signature, nil, nil, nil)
    if err != nil {
        return nil, fmt.Errorf("verify signature: %w", err)
    }

    // Ensure recovered address matches
    if publicKey.String() != msg.GetAddress().Hex() {
        return nil, fmt.Errorf("address mismatch")
    }

    // Check expiration
    if msg.GetExpirationTime() != nil && time.Now().After(*msg.GetExpirationTime()) {
        return nil, fmt.Errorf("message expired")
    }

    return &SIWEMessage{
        Address:   msg.GetAddress().Hex(),
        Statement: msg.GetStatement(),
        Nonce:     msg.GetNonce(),
    }, nil
}
```

**Deliverable**: SIWE service with nonce generation and message verification, >80% test coverage

---

#### Task 3.2: JWT Token Management (TDD)
**Skills Used**:
- `web3-authentication-authorization` - JWT security best practices
- `test-driven-development` - TDD for JWT service
- `go-backend-development` - Go implementation

**TDD Approach**: FULL TDD

**Step 1 (🔴 RED): Write tests**
```go
// internal/auth/jwt_test.go
package auth

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestGenerateJWT_ValidToken(t *testing.T) {
    secret := []byte("test-secret-key")
    address := "0x1234567890123456789012345678901234567890"
    scopes := []string{"read:data"}

    token, err := GenerateJWT(secret, address, scopes, 24*time.Hour)

    require.NoError(t, err)
    assert.NotEmpty(t, token)
}

func TestGenerateAndVerifyJWT_RoundTrip(t *testing.T) {
    secret := []byte("test-secret-key")
    address := "0x1234567890123456789012345678901234567890"
    scopes := []string{"read:data", "write:data"}

    token, _ := GenerateJWT(secret, address, scopes, 24*time.Hour)
    claims, err := VerifyJWT(secret, token)

    require.NoError(t, err)
    assert.Equal(t, address, claims.Address)
    assert.Equal(t, scopes, claims.Scopes)
}

func TestVerifyJWT_ExpiredToken(t *testing.T) {
    secret := []byte("test-secret-key")
    address := "0x1234567890123456789012345678901234567890"

    token, _ := GenerateJWT(secret, address, []string{}, -1*time.Hour) // Expired
    claims, err := VerifyJWT(secret, token)

    require.Error(t, err)
    assert.Nil(t, claims)
}

func TestVerifyJWT_InvalidSignature(t *testing.T) {
    secret := []byte("test-secret-key")
    wrongSecret := []byte("wrong-secret")
    address := "0x1234567890123456789012345678901234567890"

    token, _ := GenerateJWT(secret, address, []string{}, 24*time.Hour)
    claims, err := VerifyJWT(wrongSecret, token)

    require.Error(t, err)
    assert.Nil(t, claims)
}
```

**Step 2 (🟢 GREEN): Implement JWT**
```go
// internal/auth/jwt.go
package auth

import (
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    Address string   `json:"address"`
    Scopes  []string `json:"scopes"`
    jwt.RegisteredClaims
}

func GenerateJWT(secret []byte, address string, scopes []string, expiry time.Duration) (string, error) {
    now := time.Now()

    claims := Claims{
        Address: address,
        Scopes:  scopes,
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   address,
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
            NotBefore: jwt.NewNumericDate(now),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secret)
}

func VerifyJWT(secret []byte, tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method")
        }
        return secret, nil
    })

    if err != nil {
        return nil, err
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }

    return claims, nil
}
```

**Deliverable**: JWT service with generation and verification, tests passing

---

### Block 4: HTTP Handlers & Middleware (TDD)

#### Task 4.1: JWT Middleware
**Skills Used**:
- `go-backend-development` - Middleware patterns
- `test-driven-development` - TDD for middleware
- `web3-authentication-authorization` - Auth middleware best practices

**TDD Approach**: Test middleware behavior

**Test-Driven Implementation**:
```go
// internal/http/middleware/auth_test.go
package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
    secret := []byte("test-secret")
    token := generateTestJWT(t, secret, "0x1234...")

    handler := AuthMiddleware(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    req := httptest.NewRequest("GET", "/", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
    handler := AuthMiddleware([]byte("secret"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    req := httptest.NewRequest("GET", "/", nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
    handler := AuthMiddleware([]byte("secret"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    req := httptest.NewRequest("GET", "/", nil)
    req.Header.Set("Authorization", "Bearer invalid-token")
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
```

**Implementation**:
```go
// internal/http/middleware/auth.go
package middleware

import (
    "context"
    "net/http"
    "strings"

    "github.com/yourusername/gatekeeper/internal/auth"
)

type contextKey string

const claimsKey contextKey = "claims"

func AuthMiddleware(jwtSecret []byte) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
                return
            }

            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" {
                http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
                return
            }

            claims, err := auth.VerifyJWT(jwtSecret, parts[1])
            if err != nil {
                http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), claimsKey, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func GetClaimsFromContext(ctx context.Context) *auth.Claims {
    claims, ok := ctx.Value(claimsKey).(*auth.Claims)
    if !ok {
        return nil
    }
    return claims
}
```

**Deliverable**: JWT middleware with tests

---

#### Task 4.2: Authentication HTTP Handlers (TDD)
**Skills Used**:
- `go-backend-development` - HTTP handler patterns
- `test-driven-development` - TDD for handlers
- `api-design-openapi` - API design
- `web3-authentication-authorization` - SIWE best practices

**TDD Approach**: Write handler tests first

**Step 1 (🔴 RED): Write comprehensive handler tests**
```go
// internal/http/handlers/auth_test.go
package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestGetNonce_ReturnsNonce(t *testing.T) {
    handler := NewAuthHandler(mockSIWEService, mockJWTService)

    req := httptest.NewRequest("GET", "/auth/siwe/nonce", nil)
    rec := httptest.NewRecorder()

    handler.GetNonce(rec, req)

    require.Equal(t, http.StatusOK, rec.Code)
    assert.NotEmpty(t, rec.Header().Get("Content-Type"))

    var resp map[string]string
    json.NewDecoder(rec.Body).Decode(&resp)
    assert.NotEmpty(t, resp["nonce"])
}

func TestVerifySIWE_ValidSignature(t *testing.T) {
    handler := NewAuthHandler(mockSIWEService, mockJWTService)

    reqBody := map[string]string{
        "message":   "valid-siwe-message",
        "signature": "0xvalidsignature",
    }
    body, _ := json.Marshal(reqBody)

    req := httptest.NewRequest("POST", "/auth/siwe/verify", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    handler.VerifySIWE(rec, req)

    require.Equal(t, http.StatusOK, rec.Code)

    var resp map[string]string
    json.NewDecoder(rec.Body).Decode(&resp)
    assert.NotEmpty(t, resp["token"])
    assert.NotEmpty(t, resp["address"])
}

func TestVerifySIWE_InvalidSignature(t *testing.T) {
    handler := NewAuthHandler(mockSIWEService, mockJWTService)

    reqBody := map[string]string{
        "message":   "valid-message",
        "signature": "0xinvalid",
    }
    body, _ := json.Marshal(reqBody)

    req := httptest.NewRequest("POST", "/auth/siwe/verify", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    handler.VerifySIWE(rec, req)

    require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestVerifySIWE_MissingFields(t *testing.T) {
    handler := NewAuthHandler(mockSIWEService, mockJWTService)

    reqBody := map[string]string{
        "message": "only-message",
    }
    body, _ := json.Marshal(reqBody)

    req := httptest.NewRequest("POST", "/auth/siwe/verify", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    handler.VerifySIWE(rec, req)

    require.Equal(t, http.StatusBadRequest, rec.Code)
}
```

**Step 2 (🟢 GREEN): Implement handlers**
```go
// internal/http/handlers/auth.go
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/yourusername/gatekeeper/internal/auth"
)

type AuthHandler struct {
    siweService *auth.SIWEService
    jwtSecret   []byte
}

func NewAuthHandler(siweService *auth.SIWEService, jwtSecret []byte) *AuthHandler {
    return &AuthHandler{
        siweService: siweService,
        jwtSecret:   jwtSecret,
    }
}

type GetNonceResponse struct {
    Nonce string `json:"nonce"`
}

func (h *AuthHandler) GetNonce(w http.ResponseWriter, r *http.Request) {
    nonce, err := h.siweService.GenerateNonce(r.Context())
    if err != nil {
        http.Error(w, `{"error":"failed to generate nonce"}`, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(GetNonceResponse{Nonce: nonce})
}

type VerifyRequest struct {
    Message   string `json:"message"`
    Signature string `json:"signature"`
}

type VerifyResponse struct {
    Token     string `json:"token"`
    Address   string `json:"address"`
    ExpiresAt string `json:"expiresAt"`
}

func (h *AuthHandler) VerifySIWE(w http.ResponseWriter, r *http.Request) {
    var req VerifyRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
        return
    }

    if req.Message == "" || req.Signature == "" {
        http.Error(w, `{"error":"missing message or signature"}`, http.StatusBadRequest)
        return
    }

    // Verify message
    msg, err := h.siweService.VerifyMessage(r.Context(), req.Message, req.Signature)
    if err != nil {
        http.Error(w, `{"error":"invalid signature"}`, http.StatusUnauthorized)
        return
    }

    // Generate JWT
    token, err := auth.GenerateJWT(h.jwtSecret, msg.Address, []string{}, 24*time.Hour)
    if err != nil {
        http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(VerifyResponse{
        Token:     token,
        Address:   msg.Address,
        ExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
    })
}
```

**Deliverable**: Auth handlers with comprehensive tests

---

## Summary of Phase 1 Task Mapping

| Task Group | Skills Used | TDD | Priority |
|-----------|------------|-----|----------|
| Project Structure | go-backend-development | No | P0 |
| Configuration | go-backend-development + test-driven-development | **TDD** | P0 |
| Database Setup | go-backend-development | No | P0 |
| Logging | go-backend-development | No | P0 |
| SIWE Nonce | web3-authentication-authorization + test-driven-development | **TDD** | P0 |
| JWT Management | web3-authentication-authorization + test-driven-development | **TDD** | P0 |
| Auth Middleware | go-backend-development + test-driven-development | **TDD** | P0 |
| Auth Handlers | go-backend-development + api-design-openapi + test-driven-development | **TDD** | P0 |

## Running Tests

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -cover

# Run specific package
go test ./internal/auth -v

# Watch mode (using air or similar)
air
```

## Next Steps

1. **Clone this execution plan into your IDE**
2. **Start with Block 1**: Project structure and configuration
3. **Apply TDD strictly** for Tasks 1.2, 3.1, 3.2, 4.1, 4.2
4. **Run tests frequently** - after each RED-GREEN-REFACTOR cycle
5. **Reference the skills** as you implement each component
6. **Commit after each block** completes successfully

---

**Ready to start implementing Phase 1?** Let me know when you're ready to dive into Block 1!
