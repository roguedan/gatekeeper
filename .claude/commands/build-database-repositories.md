# Build Database Repositories

Generate production-ready Go repository implementations for managing users, API keys, and allowlists in the Gatekeeper authentication gateway.

## Instructions

When the user provides context or requests implementation, help create Go repository code that:

### 1. User Repository (`internal/store/users.go`)

**Struct Definition:**
```go
type User struct {
    ID        int64
    Address   string    // Ethereum address (42 chars, lowercase)
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Methods to Implement:**
- `CreateUser(ctx context.Context, address string) (*User, error)`
  - Validates Ethereum address format (0x + 40 hex chars)
  - Converts to lowercase
  - Stores in `users` table
  - Returns user with ID and timestamps
  - Handle duplicate address error gracefully

- `GetUserByAddress(ctx context.Context, address string) (*User, error)`
  - Query by normalized address
  - Return ErrNotFound if missing
  - Include timestamps

- `GetUserByID(ctx context.Context, id int64) (*User, error)`
  - Query by primary key
  - Return ErrNotFound if missing

- `UpdateUser(ctx context.Context, user *User) error`
  - Update existing user
  - Update `updated_at` timestamp
  - Validate address format

- `DeleteUser(ctx context.Context, id int64) error`
  - Soft or hard delete (decide based on data retention needs)

### 2. API Key Repository (`internal/store/api_keys.go`)

**Struct Definitions:**
```go
type APIKey struct {
    ID        int64
    UserID    int64
    KeyHash   string      // SHA256 hash of actual key
    Name      string
    Scopes    []string    // Postgres TEXT[] array
    LastUsedAt *time.Time
    ExpiresAt *time.Time
    CreatedAt time.Time
    UpdatedAt time.Time
}

type APIKeyCreateRequest struct {
    UserID    int64
    Name      string
    Scopes    []string
    ExpiresIn *time.Duration  // nil = never expire
}

type APIKeyResponse struct {
    ID        int64
    KeyHash   string
    Name      string
    Scopes    []string
    ExpiresAt *time.Time
    CreatedAt time.Time
}
```

**Methods to Implement:**
- `CreateAPIKey(ctx context.Context, req APIKeyCreateRequest) (key string, response *APIKeyResponse, error)`
  - Generate cryptographically secure random key (32 bytes, hex-encoded)
  - Hash key with SHA256
  - Store hash in database
  - Return raw key (only time it's visible) + metadata
  - Set `created_at` and optional `expires_at`

- `ValidateAPIKey(ctx context.Context, key string) (*APIKey, error)`
  - Hash incoming key with SHA256
  - Look up hash in database
  - Check if expired
  - Return APIKey with user info
  - DON'T return raw key (already hashed)
  - Fail-closed on any error

- `GetAPIKey(ctx context.Context, id int64) (*APIKey, error)`
  - Query by ID
  - Return metadata only (no raw key)

- `ListAPIKeys(ctx context.Context, userID int64) ([]APIKey, error)`
  - All active keys for user
  - Include expired keys (for cleanup UI)
  - Order by created_at DESC

- `UpdateLastUsed(ctx context.Context, keyHash string) error`
  - Set `last_used_at = NOW()`
  - Called after successful key validation
  - Fast operation (important for hot path)

- `DeleteAPIKey(ctx context.Context, id int64) error`
  - Hard delete or update is_deleted flag
  - Verify ownership (user_id check) before deleting
  - Return error if not found

- `RevokeExpiredKeys(ctx context.Context) (int, error)`
  - Delete or mark as expired all keys where expires_at < NOW()
  - Return count of revoked keys
  - Run on startup and periodically

### 3. Allowlist Repository (`internal/store/allowlists.go`)

**Struct Definitions:**
```go
type Allowlist struct {
    ID          int64
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type AllowlistEntry struct {
    ID          int64
    AllowlistID int64
    Address     string    // Ethereum address, lowercase
    AddedAt     time.Time
}
```

**Methods to Implement:**
- `CreateAllowlist(ctx context.Context, name, description string) (*Allowlist, error)`
  - Insert into `allowlists` table
  - Return with ID and timestamps
  - Name should be unique

- `GetAllowlist(ctx context.Context, id int64) (*Allowlist, error)`
  - Query by ID
  - Return with metadata

- `ListAllowlists(ctx context.Context) ([]Allowlist, error)`
  - All allowlists
  - Order by created_at DESC
  - Include entry counts

- `UpdateAllowlist(ctx context.Context, allowlist *Allowlist) error`
  - Update name/description
  - Update `updated_at`

- `DeleteAllowlist(ctx context.Context, id int64) error`
  - Delete allowlist and all entries
  - Cascade delete

- `AddAddress(ctx context.Context, allowlistID int64, address string) error`
  - Validate address format
  - Normalize to lowercase
  - Insert into `allowlist_entries`
  - Handle duplicate gracefully (UPSERT)
  - Update allowlist's `updated_at`

- `RemoveAddress(ctx context.Context, allowlistID int64, address string) error`
  - Delete entry
  - Update allowlist's `updated_at`

- `CheckAddress(ctx context.Context, allowlistID int64, address string) (bool, error)`
  - Fast query: does address exist in allowlist?
  - Return true/false
  - This is called in request path - MUST be fast
  - Should be cacheable

- `AddAddresses(ctx context.Context, allowlistID int64, addresses []string) error`
  - Batch add addresses
  - Validate all before inserting
  - Transactional
  - More efficient than looping AddAddress

- `GetAddresses(ctx context.Context, allowlistID int64) ([]string, error)`
  - All addresses in allowlist
  - Return sorted
  - Include count

- `RemoveAllowlist(ctx context.Context, id int64) error`
  - Cascade delete the list and entries

### Implementation Patterns

**1. Error Handling:**
- Use `ErrNotFound`, `ErrDuplicate`, `ErrInvalidAddress` custom error types
- Wrap database errors meaningfully
- Return structured errors, not raw SQL errors

**2. Database Operations:**
- All methods accept `context.Context` for cancellation/timeout
- Use `sqlx.NamedExecContext` for parameterized queries
- Use transactions for multi-table operations
- Use connection pool from `db.go`

**3. Address Validation:**
- Accept addresses in any case
- Validate format: `^0x[0-9a-fA-F]{40}$`
- Store normalized as lowercase
- Return normalized in results

**4. Security:**
- API keys stored as SHA256 hashes only
- Raw key returned only on creation
- Never log raw keys
- Validate key length before hashing (32+ bytes)

**5. Performance:**
- Use indexes: `idx_api_keys_user_id`, `idx_api_keys_key_hash`, `idx_allowlist_entries_address`
- Batch operations use transactions
- `CheckAddress` should use EXISTS subquery (very fast)

**6. Testing:**
- Create table fixtures for tests
- Test CRUD operations
- Test error cases
- Test edge cases (nil values, empty arrays, duplicates)
- Use test transaction rollback for isolation

### Database Context Integration

```go
// In internal/store/db.go, expose these repositories:
type Repositories struct {
    Users      UserRepository
    APIKeys    APIKeyRepository
    Allowlists AllowlistRepository
}

// Add to main.go initialization:
repos := &Repositories{
    Users:      NewUserRepository(db),
    APIKeys:    NewAPIKeyRepository(db),
    Allowlists: NewAllowlistRepository(db),
}
```

### Testing Requirements

**User Repository Tests:**
- Create, read, update, delete user
- Validate address format
- Handle duplicates
- Timestamp behavior

**API Key Repository Tests:**
- Generate unique keys
- Hash verification
- Expiration checking
- Last used tracking
- Batch operations

**Allowlist Repository Tests:**
- CRUD operations
- Address validation
- Batch add/remove
- Case-insensitive lookup
- Performance (CheckAddress)

### Files to Create

1. `internal/store/users.go` - User CRUD (~150 lines)
2. `internal/store/api_keys.go` - API key management (~250 lines)
3. `internal/store/allowlists.go` - Allowlist management (~200 lines)
4. `internal/store/errors.go` - Custom error types (~30 lines)
5. Database migrations (if not already created)

### Success Criteria

- ✅ All CRUD operations working
- ✅ >85% test coverage per repository
- ✅ No raw SQL errors exposed
- ✅ Addresses validated and normalized
- ✅ Performance acceptable (CheckAddress <5ms)
- ✅ Thread-safe operations
- ✅ Proper error handling with custom error types

## When to Use This Skill

Use this skill when you need to:
- Implement database persistence for users, API keys, allowlists
- Create repository layer with proper abstraction
- Add database methods with transaction support
- Write tests for data access layer
- Integrate with existing Gatekeeper database setup

---

**Generated for:** Gatekeeper MVP Phase 2 - Database Layer
**Status:** Database repositories are critical path for API key and allowlist management
