# API Service Specification

## ADDED Requirements

### REQ-API-001: OpenAPI Documentation
The system SHALL provide comprehensive OpenAPI 3.0 specification for all endpoints.

**Scenario: OpenAPI specification completeness**
- GIVEN the API has multiple endpoints
- WHEN the OpenAPI spec is generated
- THEN the spec SHALL document ALL public endpoints
- AND SHALL include request/response schemas for each endpoint
- AND SHALL define all security schemes (JWT, API keys)
- AND SHALL include examples for all requests and responses
- AND SHALL specify all possible HTTP status codes
- AND SHALL include descriptions for all parameters

**Scenario: Serve OpenAPI specification file**
- GIVEN a user wants to access the API specification
- WHEN GET /openapi.yaml is requested
- THEN the system SHALL return the OpenAPI 3.0 specification file
- AND SHALL set Content-Type: application/yaml
- AND the spec SHALL be valid according to OpenAPI 3.0 schema

**Scenario: Interactive API documentation**
- GIVEN a user wants to explore the API
- WHEN GET /docs is requested
- THEN the system SHALL serve an interactive documentation page (Redoc)
- AND the documentation SHALL load the OpenAPI spec from /openapi.yaml
- AND SHALL allow users to view all endpoints and schemas
- AND SHALL provide "Try it out" functionality (if enabled)

### REQ-API-002: API Key Generation
The system SHALL allow authenticated users to create scoped API keys.

**Scenario: Create API key with scopes**
- GIVEN an authenticated user with valid JWT
- WHEN POST /keys is called with {"name": "Production Bot", "scopes": ["read:data"], "expiresIn": "30d"}
- THEN the system SHALL generate a cryptographically secure random key
- AND SHALL hash the key with bcrypt before storage
- AND SHALL store metadata: name, scopes, created_at, expires_at
- AND SHALL return the plain key ONLY once: {"key": "gk_abc...", "apiKey": {metadata}}
- AND SHALL associate the key with the authenticated user
- AND the key SHALL have prefix "gk_" for identification

**Scenario: API key without expiration**
- GIVEN a user creates an API key
- WHEN "expiresIn" is not specified
- THEN the API key SHALL have no expiration date
- AND SHALL remain valid until manually revoked

**Scenario: Invalid scope in API key creation**
- GIVEN a user attempts to create an API key with invalid scope
- WHEN POST /keys is called with {"name": "Test", "scopes": ["invalid:scope"]}
- THEN the system SHALL return HTTP 400 Bad Request
- AND SHALL return error: {"error": "Invalid scope", "code": "INVALID_SCOPE", "details": {"validScopes": [...]}}

**Scenario: API key name validation**
- GIVEN a user creates an API key
- WHEN the name is empty or >100 characters
- THEN the system SHALL return HTTP 400 Bad Request
- AND SHALL return validation error explaining the constraint

### REQ-API-003: API Key Authentication
The system SHALL support authentication using API keys in requests.

**Scenario: Request with valid API key**
- GIVEN a request to a protected endpoint
- WHEN the request includes X-API-Key: gk_validkey123
- THEN the system SHALL hash the provided key
- AND SHALL lookup the key in the database
- AND SHALL verify the key has not expired
- AND SHALL extract the associated scopes
- AND SHALL add user ID and scopes to request context
- AND SHALL allow the request to proceed

**Scenario: Request with expired API key**
- GIVEN an API key with expires_at in the past
- WHEN a request is made with the expired key
- THEN the system SHALL return HTTP 401 Unauthorized
- AND SHALL return error: {"error": "API key expired", "code": "KEY_EXPIRED"}

**Scenario: Request with invalid API key**
- GIVEN a request with X-API-Key header
- WHEN the key doesn't exist in the database
- THEN the system SHALL return HTTP 401 Unauthorized
- AND SHALL return error: {"error": "Invalid API key", "code": "INVALID_KEY"}
- AND SHALL NOT leak information about whether key exists

**Scenario: Request without API key or JWT**
- GIVEN a protected endpoint accepts either JWT or API key
- WHEN a request has neither Authorization header nor X-API-Key header
- THEN the system SHALL return HTTP 401 Unauthorized
- AND SHALL return error: {"error": "Authentication required", "code": "AUTH_REQUIRED"}

### REQ-API-004: API Key Management
The system SHALL allow users to manage their API keys.

**Scenario: List user's API keys**
- GIVEN an authenticated user
- WHEN GET /keys is called
- THEN the system SHALL return all API keys owned by the user
- AND SHALL NOT return the key secrets (only metadata)
- AND SHALL include: id, name, scopes, created_at, expires_at, last_used_at (if tracked)
- AND SHALL support pagination with ?limit and ?offset

**Scenario: Revoke API key**
- GIVEN an authenticated user owns an API key
- WHEN DELETE /keys/{id} is called
- THEN the system SHALL verify the key belongs to the user
- AND SHALL delete the API key from the database
- AND SHALL return HTTP 204 No Content on success
- AND the key SHALL immediately become invalid

**Scenario: Attempt to revoke another user's key**
- GIVEN a user attempts to delete a key they don't own
- WHEN DELETE /keys/{id} is called
- THEN the system SHALL return HTTP 404 Not Found
- AND SHALL NOT leak information about the key's existence

**Scenario: View single API key details**
- GIVEN an authenticated user owns an API key
- WHEN GET /keys/{id} is called
- THEN the system SHALL return the key's metadata
- AND SHALL verify the key belongs to the user
- AND SHALL return HTTP 404 if key doesn't exist or isn't owned by user

### REQ-API-005: Scope Validation and Enforcement
The system SHALL define and enforce a standard set of scopes.

**Scenario: Define standard scopes**
- GIVEN the application has protected resources
- WHEN scopes are defined
- THEN the following standard scopes SHALL be available:
  - "read:data" - Read access to data endpoints
  - "write:data" - Write access to data endpoints
  - "read:keys" - Read own API keys
  - "write:keys" - Create/delete own API keys
  - "admin:*" - Administrative access (reserved)
- AND custom scopes MAY be added through configuration

**Scenario: Scope enforcement on endpoints**
- GIVEN an endpoint requires scope "write:data"
- WHEN a request is made with token having scopes ["read:data"]
- THEN the system SHALL return HTTP 403 Forbidden
- AND SHALL return error: {"error": "Insufficient permissions", "code": "INSUFFICIENT_SCOPE", "details": {"required": "write:data"}}

**Scenario: Multiple scopes satisfied**
- GIVEN an endpoint accepts any of: ["read:data", "admin:read"]
- WHEN a request has scope "admin:read"
- THEN the system SHALL allow access

### REQ-API-006: Health Check Endpoint
The system SHALL provide a health check endpoint for monitoring.

**Scenario: Healthy system**
- GIVEN all system dependencies are operational
- WHEN GET /health is called
- THEN the system SHALL return HTTP 200 OK
- AND SHALL return {"status": "ok", "version": "x.y.z", "checks": {"database": "ok", "rpc": "ok"}}

**Scenario: Database unavailable**
- GIVEN the database connection fails
- WHEN GET /health is called
- THEN the system SHALL return HTTP 503 Service Unavailable
- AND SHALL return {"status": "degraded", "checks": {"database": "failed", "rpc": "ok"}}

**Scenario: RPC provider unavailable**
- GIVEN the primary RPC provider is down
- WHEN GET /health is called
- THEN the system SHALL check fallback provider
- AND SHALL return status based on whether any provider is available

### REQ-API-007: Error Response Format
The system SHALL use a consistent error response format across all endpoints.

**Scenario: Standard error response**
- GIVEN any error occurs
- WHEN an error response is sent
- THEN the response SHALL include:
  - "error" (string) - Human-readable error message
  - "code" (string) - Machine-readable error code
  - "details" (object, optional) - Additional context
- AND SHALL have appropriate HTTP status code

**Scenario: Validation error details**
- GIVEN request validation fails
- WHEN validation error occurs
- THEN the error details SHALL include field-specific errors
- AND SHALL specify which fields are invalid and why
- AND SHALL help clients correct the request

**Scenario: No information leakage**
- GIVEN an internal server error
- WHEN error response is sent to client
- THEN the response SHALL NOT include stack traces
- AND SHALL NOT include internal paths or configurations
- AND SHALL NOT expose database schema details
- AND detailed errors SHALL be logged server-side only

### REQ-API-008: Request Validation
The system SHALL validate all incoming requests before processing.

**Scenario: JSON request body validation**
- GIVEN an endpoint accepts JSON body
- WHEN invalid JSON is sent
- THEN the system SHALL return HTTP 400 Bad Request
- AND SHALL return error: {"error": "Invalid JSON", "code": "INVALID_JSON"}

**Scenario: Required field validation**
- GIVEN a request body with required fields
- WHEN a required field is missing
- THEN the system SHALL return HTTP 400 Bad Request
- AND SHALL specify which field is missing

**Scenario: Field type validation**
- GIVEN a field expecting integer
- WHEN a string is provided
- THEN the system SHALL return HTTP 400 Bad Request
- AND SHALL specify the expected type

**Scenario: Field length validation**
- GIVEN a string field with max length
- WHEN the value exceeds the limit
- THEN the system SHALL return HTTP 400 Bad Request
- AND SHALL specify the length constraint

### REQ-API-009: Pagination
The system SHALL support pagination on list endpoints.

**Scenario: Paginate API keys list**
- GIVEN a user has 50 API keys
- WHEN GET /keys?limit=20&offset=0 is called
- THEN the system SHALL return the first 20 keys
- AND SHALL include total count in response
- AND SHALL include hasMore boolean flag
- AND response SHALL be: {"keys": [...], "total": 50, "limit": 20, "offset": 0, "hasMore": true}

**Scenario: Default pagination values**
- GIVEN no pagination parameters are provided
- WHEN GET /keys is called
- THEN the system SHALL default to limit=20, offset=0

**Scenario: Maximum page size limit**
- GIVEN a request with limit=1000
- WHEN the limit exceeds maximum (100)
- THEN the system SHALL cap at maximum limit
- AND SHALL use limit=100

### REQ-API-010: API Versioning
The system SHALL support API versioning for backward compatibility.

**Scenario: Version in URL path**
- GIVEN the API is at version 1
- WHEN endpoints are designed
- THEN routes MAY include /v1/ prefix (e.g., /v1/keys)
- OR version MAY be in Accept header
- AND versioning strategy SHALL be documented

**Scenario: Breaking changes**
- GIVEN a breaking API change is needed
- WHEN the change is implemented
- THEN a new API version SHALL be created
- AND old version SHALL be maintained for deprecation period
- AND deprecation timeline SHALL be communicated to clients

### REQ-API-011: CORS Configuration
The system SHALL support Cross-Origin Resource Sharing (CORS).

**Scenario: CORS preflight request**
- GIVEN a browser makes a preflight OPTIONS request
- WHEN the request is received
- THEN the system SHALL return appropriate CORS headers:
  - Access-Control-Allow-Origin
  - Access-Control-Allow-Methods
  - Access-Control-Allow-Headers
  - Access-Control-Max-Age

**Scenario: CORS origin whitelist**
- GIVEN configured allowed origins
- WHEN a request comes from an allowed origin
- THEN the system SHALL include that origin in Access-Control-Allow-Origin
- AND requests from disallowed origins SHALL be rejected

**Scenario: Credentials support**
- GIVEN requests include credentials (cookies, auth headers)
- WHEN CORS headers are set
- THEN Access-Control-Allow-Credentials SHALL be true if configured
- AND Access-Control-Allow-Origin SHALL NOT be * when credentials are allowed

### REQ-API-012: Request ID Tracing
The system SHALL support request tracing across the call chain.

**Scenario: Generate request ID**
- GIVEN any incoming request
- WHEN the request is received
- THEN the system SHALL generate a unique request ID (UUID)
- AND SHALL add the ID to request context
- AND SHALL include it in all logs for that request
- AND SHALL return it in X-Request-ID response header

**Scenario: Client-provided request ID**
- GIVEN a client provides X-Request-ID header
- WHEN the request is received
- THEN the system SHALL use the provided ID if valid
- AND SHALL generate new ID if provided ID is invalid
- AND SHALL include the ID in logs and response

## MODIFIED Requirements

None - this is a new service specification.

## REMOVED Requirements

None - this is a new service specification.
