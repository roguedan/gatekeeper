# Build API Key System

Generate production-ready API key management endpoints and middleware for the Gatekeeper authentication gateway.

## Instructions

When the user requests API key system implementation, help create:

### 1. API Key HTTP Handlers (`internal/http/api_key_handlers.go`)

**Request/Response Structures:**
```go
// CreateAPIKeyRequest - POST /api/keys
type CreateAPIKeyRequest struct {
    Name      string        `json:"name"`
    Scopes    []string      `json:"scopes"`
    ExpiresIn *time.Duration `json:"expiresInSeconds"`
}

// CreateAPIKeyResponse - Contains only key on creation
type CreateAPIKeyResponse struct {
    Key       string     `json:"key"`       // Raw key - only time visible
    KeyHash   string     `json:"keyHash"`   // For reference only
    Name      string     `json:"name"`
    Scopes    []string   `json:"scopes"`
    ExpiresAt *time.Time `json:"expiresAt"`
    CreatedAt time.Time  `json:"createdAt"`
}

// ListAPIKeysResponse
type ListAPIKeysResponse struct {
    Keys []APIKeyMetadata `json:"keys"`
}

type APIKeyMetadata struct {
    ID        int64       `json:"id"`
    KeyHash   string      `json:"keyHash"`  // First 8 chars visible for identification
    Name      string      `json:"name"`
    Scopes    []string    `json:"scopes"`
    ExpiresAt *time.Time  `json:"expiresAt"`
    LastUsedAt *time.Time `json:"lastUsedAt"`
    CreatedAt time.Time   `json:"createdAt"`
}

// RevokeAPIKeyRequest
type RevokeAPIKeyRequest struct {
    ID int64 `json:"id"`
}

// Error responses
type ErrorResponse struct {
    Error   string `json:"error"`
    Details string `json:"details"`
}
```

**Handler Functions:**

### `POST /api/keys` - Generate New API Key
```go
func (h *Handler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
    // 1. Get user from JWT context (from auth middleware)
    // 2. Parse CreateAPIKeyRequest from JSON body
    // 3. Validate:
    //    - Name is not empty (max 255 chars)
    //    - Scopes array is not empty
    //    - ExpiresIn is positive if provided
    // 4. Call repo.CreateAPIKey(ctx, request)
    //    - This generates key, hashes it, stores in DB
    // 5. Return CreateAPIKeyResponse with:
    //    - Raw key (hex-encoded, 64 chars)
    //    - Key hash (for reference)
    //    - Metadata
    // 6. Status: 201 Created
    // 7. Log: "API key created" with user address and key name
}
```

**Implementation Details:**
- Extract address from JWT claims via `ClaimsFromContext(r.Context())`
- Look up user ID from address via UserRepository
- Validate request body with clear error messages
- Handle errors: 400 (bad request), 404 (user not found), 500 (db error)
- Response should clearly state: "Save this key securely - you won't see it again"
- Set response header: `Cache-Control: no-store`

### `GET /api/keys` - List User's API Keys
```go
func (h *Handler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
    // 1. Get user address from JWT context
    // 2. Resolve to user ID via UserRepository
    // 3. Call repo.ListAPIKeys(ctx, userID)
    // 4. Return array of APIKeyMetadata (no raw keys)
    // 5. Show first 8 chars of key hash for identification
    // 6. Status: 200 OK
}
```

**Implementation Details:**
- Include both active and expired keys (for UI clarity)
- Order by created_at DESC (newest first)
- Show last_used_at for tracking usage
- Include expiration status indicator

### `DELETE /api/keys/{id}` - Revoke API Key
```go
func (h *Handler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
    // 1. Extract key ID from URL parameter
    // 2. Get user address from JWT context
    // 3. Verify key belongs to user (ownership check)
    // 4. Call repo.DeleteAPIKey(ctx, id)
    // 5. Return 204 No Content on success
    // 6. Return 404 if key not found or belongs to different user
    // 7. Log: "API key revoked" with user and key ID
}
```

**Implementation Details:**
- Fetch key metadata first to verify ownership
- Prevent users from revoking other users' keys
- Return 403 Forbidden if ownership check fails
- Idempotent: 204 even if already deleted

### 2. API Key Validation Middleware (`internal/http/api_key_middleware.go`)

**Middleware Function:**
```go
func (h *Handler) APIKeyMiddleware(next http.Handler) http.Handler {
    // Signature: http.Handler -> http.Handler

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Check for API key in header:
        //    - Primary: "X-API-Key" header
        //    - Fallback: "Authorization: Bearer <key>" (alt format)

        // 2. If no key found:
        //    - Pass through (let JWT middleware handle)
        //    - Mark request as "key_auth_skipped"

        // 3. If key found:
        //    - Validate key format (hex, 64 chars)
        //    - Call repo.ValidateAPIKey(ctx, key)
        //    - On error, return 401 Unauthorized with message

        // 4. On success:
        //    - Extract user info from returned APIKey
        //    - Create Claims struct with address and scopes
        //    - Inject into context (use ClaimsContextKey)
        //    - Update last_used_at async (non-blocking)
        //    - Call next.ServeHTTP(w, r)

        // 5. Audit logging:
        //    - Log: "API key validation" with key hash, address, decision
    })
}
```

**Implementation Details:**
- Extract from `X-API-Key: <key>` header (preferred)
- Support alternative: `Authorization: Bearer <key>`
- Validate hex format before DB lookup (fast fail)
- Call `repo.ValidateAPIKey()` which checks:
  - Key hash exists
  - Not expired
  - Not revoked
- Convert API key to Claims with same scopes
- Update `last_used_at` in background goroutine (don't block request)
- If both JWT and API key valid, JWT takes precedence
- Return 401 with clear error messages

**Error Responses:**
```json
{
  "error": "invalid_api_key",
  "details": "API key not found or expired"
}
```

### 3. Integration with Existing HTTP Server

**In `cmd/server/main.go`:**
```go
// After creating repositories:
apiKeyHandlers := http.NewAPIKeyHandlers(repos.APIKeys, repos.Users)

// Mount handlers:
router.HandleFunc("/api/keys", apiKeyHandlers.CreateAPIKey).Methods("POST")
router.HandleFunc("/api/keys", apiKeyHandlers.ListAPIKeys).Methods("GET")
router.HandleFunc("/api/keys/{id}", apiKeyHandlers.RevokeAPIKey).Methods("DELETE")

// Add middleware chain:
// All routes should support both JWT and API key auth
router.Use(h.APIKeyMiddleware)      // Try API key first
router.Use(h.JWTMiddleware)         // Fall back to JWT
```

**Middleware Ordering:**
1. API Key Middleware - Validates X-API-Key header
2. JWT Middleware - Validates Bearer token (if no API key)
3. Policy Middleware - Evaluates policies (existing)

### 4. OpenAPI Documentation Updates

**In `openapi.yaml`, add:**

```yaml
/api/keys:
  post:
    summary: "Create new API key"
    tags: ["API Keys"]
    requestBody:
      required: true
      content:
        application/json:
          schema:
            type: object
            required: [name, scopes]
            properties:
              name:
                type: string
                description: "Human-readable name"
                maxLength: 255
              scopes:
                type: array
                items:
                  type: string
                minItems: 1
              expiresInSeconds:
                type: integer
                description: "Optional expiration time in seconds"
    responses:
      '201':
        description: "API key created"
        content:
          application/json:
            schema:
              type: object
              properties:
                key:
                  type: string
                  description: "Raw API key (save securely - won't be shown again)"
                keyHash:
                  type: string
                name:
                  type: string
                scopes:
                  type: array
                  items: string
                expiresAt:
                  type: string
                  format: date-time
                createdAt:
                  type: string
                  format: date-time
      '400':
        $ref: '#/components/responses/BadRequest'
      '401':
        $ref: '#/components/responses/Unauthorized'

  get:
    summary: "List user's API keys"
    tags: ["API Keys"]
    security:
      - BearerAuth: []
      - APIKey: []
    responses:
      '200':
        description: "List of API keys"
        content:
          application/json:
            schema:
              type: object
              properties:
                keys:
                  type: array
                  items:
                    $ref: '#/components/schemas/APIKeyMetadata'
      '401':
        $ref: '#/components/responses/Unauthorized'

/api/keys/{id}:
  delete:
    summary: "Revoke API key"
    tags: ["API Keys"]
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: integer
    security:
      - BearerAuth: []
    responses:
      '204':
        description: "Key revoked successfully"
      '401':
        $ref: '#/components/responses/Unauthorized'
      '404':
        description: "Key not found"

components:
  securitySchemes:
    APIKey:
      type: apiKey
      in: header
      name: X-API-Key
      description: "API key for programmatic access"

    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: "JWT token from SIWE authentication"

  schemas:
    APIKeyMetadata:
      type: object
      properties:
        id:
          type: integer
        keyHash:
          type: string
          description: "First 8 chars visible for identification"
        name:
          type: string
        scopes:
          type: array
          items: string
        expiresAt:
          type: string
          format: date-time
        lastUsedAt:
          type: string
          format: date-time
        createdAt:
          type: string
          format: date-time
```

### 5. Testing Requirements

**Unit Tests (`internal/http/api_key_handlers_test.go`):**
- Create valid API key request
- Create with missing name/scopes
- Create with invalid scopes
- List keys for user
- List empty keys
- Revoke own key
- Try to revoke another user's key (403)
- Revoke non-existent key (404)

**Unit Tests (`internal/http/api_key_middleware_test.go`):**
- Valid API key in X-API-Key header
- Valid API key in Authorization header
- Invalid key format
- Expired key
- Non-existent key
- Multiple authentication methods (preference)
- Fallback to JWT if no API key

**Integration Tests:**
- Full workflow: create key -> use in request -> verify in DB
- Key expiration enforcement
- Last-used tracking
- Revocation immediate effect

### Implementation Order

1. **Phase 1:** Implement handlers (CreateAPIKey, ListAPIKeys, RevokeAPIKey)
2. **Phase 2:** Implement middleware (APIKeyMiddleware)
3. **Phase 3:** Integration tests
4. **Phase 4:** OpenAPI documentation
5. **Phase 5:** Examples and quickstart

### Success Criteria

- ✅ API keys generated securely (32 bytes, cryptographically random)
- ✅ Keys hashed before storage (SHA256)
- ✅ Raw key shown only on creation
- ✅ Validation middleware works with both JWT and API keys
- ✅ Ownership checks prevent cross-user access
- ✅ Expiration enforced automatically
- ✅ Usage tracking works (last_used_at)
- ✅ >85% test coverage
- ✅ No raw keys in logs or responses (except creation)
- ✅ Clear error messages for client debugging

## When to Use This Skill

Use this skill when you need to:
- Implement API key generation and management endpoints
- Add API key validation middleware
- Create secure key handling (generation, hashing, storage)
- Integrate with database repositories
- Write tests for API key workflows
- Document API key usage in OpenAPI spec

---

**Generated for:** Gatekeeper MVP Phase 2 - API Key Management
**Importance:** Critical for programmatic access to Gatekeeper
