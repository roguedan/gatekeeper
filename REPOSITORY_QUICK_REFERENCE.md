# Gatekeeper Repository Quick Reference

## Quick Start

```go
import "github.com/yourusername/gatekeeper/internal/store"

// Setup
db, _ := store.Connect(ctx, dbURL)
db.RunMigrations()

// Create repositories
users := store.NewUserRepository(db)
apiKeys := store.NewAPIKeyRepository(db)
allowlists := store.NewAllowlistRepository(db)
```

## User Repository

### Create User
```go
user, err := users.CreateUser(ctx, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
// Returns: &User{ID: 1, Address: "0x742d35cc...", CreatedAt: ..., UpdatedAt: ...}
```

### Get User
```go
// By address (case-insensitive)
user, err := users.GetUserByAddress(ctx, "0x742D35CC...")

// By ID
user, err := users.GetUserByID(ctx, 1)

// Get or create
user, err := users.GetOrCreateUserByAddress(ctx, "0x742d35Cc...")
```

### Update/Delete User
```go
user.Address = "0x1111111111111111111111111111111111111111"
err := users.UpdateUser(ctx, user)

err := users.DeleteUser(ctx, user.ID)
```

## API Key Repository

### Create API Key
```go
req := store.APIKeyCreateRequest{
    UserID:    user.ID,
    Name:      "Production Key",
    Scopes:    []string{"read", "write"},
    ExpiresIn: &duration, // Optional: time.Duration
}

rawKey, response, err := apiKeys.CreateAPIKey(ctx, req)
// rawKey: "a1b2c3d4..." (64 chars, show to user ONCE)
// response: &APIKeyResponse{ID: 1, KeyHash: "hash...", Name: "Production Key", ...}
```

### Validate API Key
```go
apiKey, err := apiKeys.ValidateAPIKey(ctx, rawKey)
if err != nil {
    // Key invalid, expired, or not found
}
// Returns full APIKey with user_id, scopes, etc.
```

### List/Get/Delete Keys
```go
// List all keys for user
keys, err := apiKeys.ListAPIKeys(ctx, userID)

// Get by ID
key, err := apiKeys.GetAPIKey(ctx, keyID)

// Delete (revoke)
err := apiKeys.DeleteAPIKey(ctx, keyID)

// Update last used
keyHash := store.HashAPIKey(rawKey)
err := apiKeys.UpdateLastUsed(ctx, keyHash)

// Cleanup expired keys
count, err := apiKeys.RevokeExpiredKeys(ctx)
```

## Allowlist Repository

### Create Allowlist
```go
list, err := allowlists.CreateAllowlist(ctx, "Premium Users", "VIP access")
// Returns: &Allowlist{ID: 1, Name: "Premium Users", Description: "VIP access", ...}
```

### Manage Addresses
```go
// Add single address
err := allowlists.AddAddress(ctx, listID, "0x742d35Cc...")

// Add multiple addresses
addresses := []string{"0x1111...", "0x2222...", "0x3333..."}
err := allowlists.AddAddresses(ctx, listID, addresses)

// Remove address
err := allowlists.RemoveAddress(ctx, listID, "0x742d35Cc...")

// Check if address exists (FAST)
exists, err := allowlists.CheckAddress(ctx, listID, "0x742d35Cc...")

// Get all addresses (sorted)
addresses, err := allowlists.GetAddresses(ctx, listID)
```

### List/Update/Delete Allowlists
```go
// List all with entry counts
lists, err := allowlists.ListAllowlists(ctx)
// Returns: []AllowlistWithCount{...}

// Get by ID
list, err := allowlists.GetAllowlist(ctx, listID)

// Update
list.Name = "Updated Name"
err := allowlists.UpdateAllowlist(ctx, list)

// Delete (cascade deletes entries)
err := allowlists.DeleteAllowlist(ctx, listID)
```

## Error Handling

```go
import "errors"

// Check error type
if errors.As(err, &store.NotFoundError{}) {
    // Handle not found
}

if errors.As(err, &store.DuplicateError{}) {
    // Handle duplicate
}

if errors.As(err, &store.InvalidAddressError{}) {
    // Handle invalid address
}

if errors.As(err, &store.ExpiredError{}) {
    // Handle expired resource
}

// Get error details
var notFoundErr *store.NotFoundError
if errors.As(err, &notFoundErr) {
    fmt.Printf("Resource %s not found: %v\n", notFoundErr.Resource, notFoundErr.ID)
}
```

## Common Patterns

### Authentication Flow
```go
// 1. User connects wallet
user, err := users.GetOrCreateUserByAddress(ctx, walletAddress)

// 2. Create API key for user
req := store.APIKeyCreateRequest{
    UserID: user.ID,
    Name:   "Web App",
    Scopes: []string{"read", "write"},
}
rawKey, _, err := apiKeys.CreateAPIKey(ctx, req)

// 3. Return key to user (show ONCE)
fmt.Printf("Your API key: %s\n", rawKey)
```

### Request Authentication
```go
// 1. Extract API key from request header
rawKey := r.Header.Get("X-API-Key")

// 2. Validate key
apiKey, err := apiKeys.ValidateAPIKey(ctx, rawKey)
if err != nil {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
}

// 3. Update last used (async)
go func() {
    keyHash := store.HashAPIKey(rawKey)
    apiKeys.UpdateLastUsed(context.Background(), keyHash)
}()

// 4. Use apiKey.UserID for request context
```

### Access Control
```go
// 1. Check if user's address is on allowlist
allowed, err := allowlists.CheckAddress(ctx, premiumListID, user.Address)
if !allowed {
    http.Error(w, "Forbidden", http.StatusForbidden)
    return
}

// 2. Grant access
// ...
```

### Batch Operations
```go
// Add multiple addresses efficiently
addresses := []string{
    "0x1111111111111111111111111111111111111111",
    "0x2222222222222222222222222222222222222222",
    "0x3333333333333333333333333333333333333333",
}
err := allowlists.AddAddresses(ctx, listID, addresses)
```

## Performance Tips

1. **Use CheckAddress for membership tests** (not GetAddresses)
   ```go
   // Fast - uses EXISTS query
   exists, err := allowlists.CheckAddress(ctx, listID, address)

   // Slow - retrieves all addresses
   addresses, err := allowlists.GetAddresses(ctx, listID)
   contains := slices.Contains(addresses, address)
   ```

2. **Batch operations over loops**
   ```go
   // Good - single transaction
   err := allowlists.AddAddresses(ctx, listID, addresses)

   // Bad - multiple transactions
   for _, addr := range addresses {
       allowlists.AddAddress(ctx, listID, addr)
   }
   ```

3. **Update LastUsed asynchronously**
   ```go
   // Don't block request
   go func() {
       keyHash := store.HashAPIKey(rawKey)
       apiKeys.UpdateLastUsed(context.Background(), keyHash)
   }()
   ```

4. **Periodically cleanup expired keys**
   ```go
   // Run as cron job
   ticker := time.NewTicker(1 * time.Hour)
   go func() {
       for range ticker.C {
           count, err := apiKeys.RevokeExpiredKeys(context.Background())
           log.Printf("Revoked %d expired keys", count)
       }
   }()
   ```

## Security Checklist

- [ ] Never log raw API keys
- [ ] Only show raw key once (on creation)
- [ ] Store only SHA256 hashes
- [ ] Validate all addresses before storage
- [ ] Use context for request cancellation
- [ ] Fail closed on validation errors
- [ ] No sensitive data in error messages

## Testing Example

```go
func TestMyHandler(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    userRepo := store.NewUserRepository(db)
    apiKeyRepo := store.NewAPIKeyRepository(db)

    // Create test user
    user, err := userRepo.CreateUser(context.Background(),
        "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
    require.NoError(t, err)

    // Create test API key
    req := store.APIKeyCreateRequest{
        UserID: user.ID,
        Name:   "Test Key",
        Scopes: []string{"read"},
    }
    rawKey, _, err := apiKeyRepo.CreateAPIKey(context.Background(), req)
    require.NoError(t, err)

    // Test handler with key
    // ...
}
```

## Environment Variables

```bash
# Database connection
export DATABASE_URL="postgres://user:pass@localhost:5432/gatekeeper?sslmode=disable"

# Test database
export TEST_DATABASE_URL="postgres://user:pass@localhost:5432/gatekeeper_test?sslmode=disable"
```

## Common Issues

### "user not found" vs "invalid address"
```go
// Invalid format - returns InvalidAddressError
_, err := users.GetUserByAddress(ctx, "invalid")

// Valid format but doesn't exist - returns NotFoundError
_, err := users.GetUserByAddress(ctx, "0x0000000000000000000000000000000000000000")
```

### Address case sensitivity
```go
// All these work (case-insensitive)
user, _ := users.CreateUser(ctx, "0xaBcDeF...")
user, _ := users.GetUserByAddress(ctx, "0xABCDEF...")
user, _ := users.GetUserByAddress(ctx, "0xabcdef...")
```

### Idempotent operations
```go
// Safe to call multiple times
allowlists.AddAddress(ctx, listID, address)
allowlists.AddAddress(ctx, listID, address) // No error
```

## Database Schema

```
users
  - id (PK)
  - address (unique, lowercase)
  - created_at
  - updated_at

api_keys
  - id (PK)
  - user_id (FK -> users)
  - key_hash (unique, SHA256)
  - name
  - scopes (text[])
  - last_used_at
  - expires_at
  - created_at
  - updated_at

allowlists
  - id (PK)
  - name (unique)
  - description
  - created_at
  - updated_at

allowlist_entries
  - id (PK)
  - allowlist_id (FK -> allowlists, cascade)
  - address (lowercase)
  - added_at
  - UNIQUE(allowlist_id, address)
```

## More Information

- Full documentation: `internal/store/README.md`
- Implementation summary: `IMPLEMENTATION_SUMMARY.md`
- Test examples: `internal/store/*_test.go`
