# Authentication Service Specification

## ADDED Requirements

### REQ-AUTH-001: SIWE Nonce Generation
The system SHALL provide an endpoint to generate cryptographically secure nonces for Sign-In With Ethereum authentication.

**Scenario: User requests authentication nonce**
- GIVEN a user wants to authenticate with their wallet
- WHEN the user requests a nonce from GET /auth/siwe/nonce
- THEN the system SHALL generate a cryptographically secure random nonce
- AND SHALL store the nonce with a 5-minute expiration time
- AND SHALL return the nonce in JSON format: {"nonce": "hex-encoded-random-bytes"}
- AND the nonce SHALL be at least 128 bits of entropy

### REQ-AUTH-002: SIWE Message Verification
The system MUST verify Sign-In With Ethereum messages and signatures according to EIP-4361.

**Scenario: User submits valid SIWE signature**
- GIVEN a user has created and signed a valid SIWE message
- WHEN the user submits the message and signature to POST /auth/siwe/verify
- THEN the system SHALL parse the SIWE message according to EIP-4361
- AND SHALL validate that the message domain matches the application domain
- AND SHALL verify the nonce exists in storage and has not expired
- AND SHALL cryptographically verify the signature against the message
- AND SHALL confirm the recovered address matches the address in the message
- AND SHALL check the message has not expired
- AND SHALL consume the nonce (mark as used)
- AND SHALL generate a JWT token for the authenticated address
- AND SHALL return {"token": "jwt", "address": "0x...", "expiresAt": "ISO8601"}

**Scenario: User submits invalid signature**
- GIVEN a user submits an invalid or tampered signature
- WHEN the signature verification fails
- THEN the system SHALL return HTTP 401 Unauthorized
- AND SHALL return error: {"error": "Invalid signature", "code": "INVALID_SIGNATURE"}
- AND SHALL NOT consume the nonce

**Scenario: User reuses a nonce**
- GIVEN a nonce has already been used for authentication
- WHEN a user attempts to reuse the same nonce
- THEN the system SHALL return HTTP 401 Unauthorized
- AND SHALL return error: {"error": "Invalid or expired nonce", "code": "NONCE_INVALID"}

**Scenario: User submits expired message**
- GIVEN a SIWE message with an expiration time in the past
- WHEN the user submits the expired message
- THEN the system SHALL return HTTP 401 Unauthorized
- AND SHALL return error: {"error": "Message expired", "code": "MESSAGE_EXPIRED"}

### REQ-AUTH-003: JWT Token Generation
The system SHALL generate JWT tokens with wallet address as subject and configurable scopes.

**Scenario: Generate JWT for authenticated user**
- GIVEN a user has successfully verified their SIWE signature
- WHEN the system generates a JWT token
- THEN the token SHALL include the wallet address as the "sub" (subject) claim
- AND SHALL include an "iat" (issued at) timestamp
- AND SHALL include an "exp" (expiration) timestamp set to 24 hours from issue
- AND SHALL include a "nbf" (not before) timestamp set to current time
- AND SHALL optionally include a "scopes" claim as an array of strings
- AND SHALL be signed with HMAC SHA-256 using a secret key
- AND SHALL be returned in the verification response

**Scenario: JWT includes custom scopes**
- GIVEN a user is authenticated with specific permissions
- WHEN the JWT is generated
- THEN the token SHALL include the granted scopes in the "scopes" claim
- AND the scopes SHALL be validated against known scope values
- AND SHALL default to empty array if no scopes provided

### REQ-AUTH-004: JWT Token Validation
The system MUST validate JWT tokens on every protected request.

**Scenario: Request with valid JWT**
- GIVEN a protected endpoint requires authentication
- WHEN a request includes a valid JWT in the Authorization header as "Bearer <token>"
- THEN the system SHALL parse the JWT
- AND SHALL verify the signature using the secret key
- AND SHALL check the token has not expired
- AND SHALL verify the "nbf" (not before) time has passed
- AND SHALL extract the claims (address, scopes)
- AND SHALL add the claims to the request context
- AND SHALL allow the request to proceed

**Scenario: Request with expired JWT**
- GIVEN a JWT token that has passed its expiration time
- WHEN a request is made with the expired token
- THEN the system SHALL return HTTP 401 Unauthorized
- AND SHALL return error: {"error": "Token expired", "code": "TOKEN_EXPIRED"}

**Scenario: Request with invalid JWT signature**
- GIVEN a JWT token with an invalid or tampered signature
- WHEN a request is made with the invalid token
- THEN the system SHALL return HTTP 401 Unauthorized
- AND SHALL return error: {"error": "Invalid token", "code": "INVALID_TOKEN"}

**Scenario: Request without JWT**
- GIVEN a protected endpoint requires authentication
- WHEN a request is made without an Authorization header
- THEN the system SHALL return HTTP 401 Unauthorized
- AND SHALL return error: {"error": "Missing authorization header", "code": "AUTH_REQUIRED"}

**Scenario: Request with malformed Authorization header**
- GIVEN a request with Authorization header not in "Bearer <token>" format
- WHEN the header is parsed
- THEN the system SHALL return HTTP 401 Unauthorized
- AND SHALL return error: {"error": "Invalid authorization format", "code": "INVALID_AUTH_FORMAT"}

### REQ-AUTH-005: Nonce Lifecycle Management
The system SHALL manage nonce lifecycle to prevent replay attacks.

**Scenario: Nonce cleanup after expiration**
- GIVEN nonces are stored with 5-minute TTL
- WHEN a nonce reaches its expiration time
- THEN the system SHALL automatically remove the expired nonce from storage
- AND SHALL run cleanup at regular intervals (every 1 minute)
- AND SHALL prevent unbounded memory growth

**Scenario: Nonce consumption on successful verification**
- GIVEN a SIWE signature is successfully verified
- WHEN the verification completes
- THEN the system SHALL immediately remove the nonce from storage
- AND the nonce SHALL NOT be usable for subsequent authentication attempts

### REQ-AUTH-006: Security Headers and CORS
The system SHALL implement security best practices for HTTP responses.

**Scenario: CORS configuration**
- GIVEN requests may come from different origins
- WHEN a CORS preflight request is made
- THEN the system SHALL return appropriate CORS headers
- AND SHALL allow configured origins only
- AND SHALL include "Authorization" in allowed headers
- AND SHALL support credentials if needed

**Scenario: Security headers on responses**
- GIVEN any HTTP response
- WHEN the response is sent
- THEN the system SHALL include Content-Type headers
- AND SHALL NOT leak sensitive information in error messages
- AND SHALL NOT include stack traces in production responses

### REQ-AUTH-007: Rate Limiting on Auth Endpoints
The system SHOULD implement rate limiting on authentication endpoints to prevent abuse.

**Scenario: Rate limiting nonce generation**
- GIVEN a single IP address or identifier
- WHEN more than 60 nonce requests are made within 1 minute
- THEN the system SHALL return HTTP 429 Too Many Requests
- AND SHALL include Retry-After header
- AND SHALL log the rate limit violation

**Scenario: Rate limiting signature verification**
- GIVEN a single IP address
- WHEN more than 20 signature verification attempts are made within 1 minute
- THEN the system SHALL return HTTP 429 Too Many Requests
- AND SHALL increase the rate limit duration on repeated violations
- AND SHALL alert on suspicious activity patterns

### REQ-AUTH-008: Audit Logging
The system SHALL log all authentication events for security auditing.

**Scenario: Log successful authentication**
- GIVEN a user successfully authenticates
- WHEN the JWT is issued
- THEN the system SHALL log an event with
  - Timestamp
  - Wallet address
  - IP address (if available)
  - User agent
  - Success status

**Scenario: Log failed authentication**
- GIVEN an authentication attempt fails
- WHEN the failure occurs
- THEN the system SHALL log an event with
  - Timestamp
  - Attempted address (if parseable)
  - IP address
  - Failure reason
  - User agent

**Scenario: Log suspicious patterns**
- GIVEN multiple failed authentication attempts from same address
- WHEN the pattern is detected
- THEN the system SHALL log a security alert
- AND SHALL include pattern details for investigation

## MODIFIED Requirements

None - this is a new service specification.

## REMOVED Requirements

None - this is a new service specification.
