# Rate Limiting Implementation

This document describes the rate limiting implementation for the Gatekeeper API.

## Overview

Rate limiting has been implemented to prevent abuse of API endpoints, particularly for API key creation and general API usage. The implementation uses a token bucket algorithm with per-user and per-IP rate limiting.

## Features

- **Per-User Rate Limiting**: Limits are enforced per authenticated user (by wallet address)
- **IP-Based Fallback**: For unauthenticated requests, limits are applied per IP address
- **Token Bucket Algorithm**: Uses `golang.org/x/time/rate` for efficient, thread-safe rate limiting
- **Configurable Limits**: All rate limits are configurable via environment variables
- **Burst Support**: Allows temporary spikes above the steady-state rate
- **429 Response**: Returns proper HTTP 429 Too Many Requests with retry information

## Configuration

Rate limiting is configured via environment variables:

```bash
# API Key Creation Rate Limit (per user per hour)
API_KEY_CREATION_RATE_LIMIT=10           # Default: 10
API_KEY_CREATION_BURST_LIMIT=3           # Default: 3

# General API Usage Rate Limit (per user per minute)
API_USAGE_RATE_LIMIT=1000                # Default: 1000
API_USAGE_BURST_LIMIT=100                # Default: 100
```

## Rate Limit Tiers

### 1. API Key Creation
- **Endpoint**: `POST /api/keys`
- **Limit**: 10 requests per hour per user
- **Burst**: 3 requests
- **Purpose**: Prevents abuse of API key generation

### 2. General API Usage
- **Endpoints**: All `/api/*` routes
- **Limit**: 1000 requests per minute per user
- **Burst**: 100 requests
- **Purpose**: Protects against general API abuse

## Implementation Details

### Core Components

1. **`rate_limit.go`**
   - `RateLimiter` interface: Defines the rate limiting contract
   - `InMemoryRateLimiter`: Token bucket implementation
   - Automatic cleanup of old entries to prevent memory leaks

2. **`rate_limit_middleware.go`**
   - `RateLimitMiddleware`: HTTP middleware for rate limiting
   - User ID extraction from JWT claims
   - IP extraction with proxy header support (X-Forwarded-For, X-Real-IP)
   - 429 response with Retry-After header

3. **Configuration in `config.go`**
   - Rate limit configuration fields
   - Environment variable loading with defaults

4. **Integration in `main.go`**
   - Rate limiter initialization
   - Middleware mounting on appropriate routes
   - Layered rate limiting (general + specific)

### Middleware Chain

The rate limiting middleware is applied in this order:

```
Request → API Key Middleware → JWT Middleware → General Rate Limit → Endpoint-Specific Rate Limit → Handler
```

For `POST /api/keys`:
1. API Key middleware (optional auth)
2. JWT middleware (required auth)
3. General API usage rate limit (1000/min)
4. API key creation rate limit (10/hour)
5. Handler

For other `/api/*` endpoints:
1. API Key middleware (optional auth)
2. JWT middleware (required auth)
3. General API usage rate limit (1000/min)
4. Handler

## Rate Limit Response

When a rate limit is exceeded, the API returns:

**Status Code**: `429 Too Many Requests`

**Headers**:
- `Retry-After`: Seconds to wait before retrying
- `X-RateLimit-Limit`: Rate limit ceiling
- `X-RateLimit-Remaining`: Remaining requests (0 when limited)
- `X-RateLimit-Reset`: Unix timestamp when limit resets

**Body**:
```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests. Please try again later.",
  "retryAfter": 60
}
```

## Usage Examples

### Checking Rate Limit Headers

```bash
# Make a request and check headers
curl -i -X POST http://localhost:8080/api/keys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Key","scopes":["read"]}'

# Response includes rate limit headers:
# X-RateLimit-Limit: 10
# X-RateLimit-Remaining: 9
# X-RateLimit-Reset: 1704067200
```

### Handling 429 Response

```javascript
async function createAPIKey(name, scopes) {
  const response = await fetch('http://localhost:8080/api/keys', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${jwtToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ name, scopes })
  });

  if (response.status === 429) {
    const retryAfter = response.headers.get('Retry-After');
    console.log(`Rate limited. Retry after ${retryAfter} seconds`);

    // Wait and retry
    await new Promise(resolve => setTimeout(resolve, retryAfter * 1000));
    return createAPIKey(name, scopes);
  }

  return response.json();
}
```

## Testing

Run the rate limiting tests:

```bash
# Test core rate limiter
go test ./internal/http -run TestInMemoryRateLimiter -v

# Test middleware
go test ./internal/http -run TestRateLimitMiddleware -v

# Run all rate limit tests
go test ./internal/http -run "TestInMemoryRateLimiter|TestRateLimitMiddleware" -v
```

## Architecture Decisions

### Why Token Bucket?

The token bucket algorithm was chosen because:
- Allows controlled bursts (better UX than strict rate limiting)
- Smooth rate limiting over time
- Well-tested implementation in `golang.org/x/time/rate`
- Thread-safe and efficient

### Why Per-User Instead of Global?

Per-user rate limiting ensures:
- Fair resource allocation
- One user can't exhaust the API for others
- Better protection against abuse
- Scales with user base

### Why In-Memory?

For this implementation:
- Simple and fast
- No external dependencies (Redis, etc.)
- Sufficient for single-instance deployments
- Easy to upgrade to distributed later if needed

### Future Enhancements

If deploying across multiple instances, consider:
- Redis-backed rate limiter for distributed state
- Configurable rate limits per user tier
- Rate limit metrics and monitoring
- Dynamic rate limit adjustment based on load

## Security Considerations

1. **IP Spoofing**: The middleware checks X-Forwarded-For and X-Real-IP headers. Ensure your reverse proxy is configured to set these correctly and strip client-provided values.

2. **Memory Usage**: The in-memory limiter automatically cleans up old entries, but in high-traffic scenarios, monitor memory usage.

3. **Bypass Protection**: Rate limiting is applied after authentication, so it's tied to the authenticated identity. Ensure JWT validation is secure.

4. **DDoS Protection**: Rate limiting helps but is not a complete DDoS solution. Use additional protections at the infrastructure level.

## Monitoring

To monitor rate limiting effectiveness:

1. Check logs for "Rate limit exceeded" warnings
2. Monitor 429 response counts
3. Track which users hit limits most frequently
4. Adjust limits based on legitimate usage patterns

## References

- [Token Bucket Algorithm](https://en.wikipedia.org/wiki/Token_bucket)
- [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate)
- [RFC 6585 - HTTP 429](https://tools.ietf.org/html/rfc6585)
