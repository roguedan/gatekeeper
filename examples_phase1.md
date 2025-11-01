# Phase 1 Implementation Examples

## Task 1: Database Health Check

### Function Usage

The `CheckDatabaseHealth` function is now available in `/Users/danwilliams/Documents/web3/gatekeeper/internal/store/db.go`:

```go
func CheckDatabaseHealth(ctx context.Context, db *sql.DB) error
```

### Example 1: Basic Usage

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"

    "github.com/yourusername/gatekeeper/internal/store"
    _ "github.com/lib/pq"
)

func main() {
    // Connect to database
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Check database health
    ctx := context.Background()
    if err := store.CheckDatabaseHealth(ctx, db); err != nil {
        log.Printf("Database health check failed: %v", err)
        // Handle unhealthy database (retry, alert, etc.)
    } else {
        fmt.Println("Database is healthy!")
    }
}
```

### Example 2: Using in HTTP Health Check Endpoint

```go
// Update the /health endpoint in cmd/server/main.go to include database health
router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    status := "ok"
    statusCode := http.StatusOK

    // Check database health
    if err := store.CheckDatabaseHealth(r.Context(), db.DB.DB); err != nil {
        logger.Error(fmt.Sprintf("Database health check failed: %v", err))
        status = "degraded"
        statusCode = http.StatusServiceUnavailable
    }

    // Check RPC health if configured
    if provider != nil {
        if !provider.HealthCheck(r.Context()) {
            status = "degraded"
            statusCode = http.StatusServiceUnavailable
        }
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    fmt.Fprintf(w, `{"status":"%s","database":"ok","port":"%s"}`, status, cfg.Port)
}).Methods("GET")
```

### Example 3: Periodic Health Checks

```go
// Run periodic database health checks
go func() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        if err := store.CheckDatabaseHealth(ctx, db.DB.DB); err != nil {
            logger.Error(fmt.Sprintf("Periodic DB health check failed: %v", err))
            // Send alert, increment metrics, etc.
        }
        cancel()
    }
}()
```

### Features
- ✅ Uses `SELECT 1` query for minimal overhead
- ✅ 5-second timeout to prevent hanging
- ✅ Returns detailed error messages
- ✅ Works with `database/sql.DB` interface
- ✅ Context-aware for cancellation

### Testing

```bash
# Test that the function compiles
go test -c ./internal/store -o /dev/null

# To test with a real database (requires DATABASE_URL env var):
# 1. Start PostgreSQL
# 2. Set DATABASE_URL
# 3. Run the application and check /health endpoint
curl http://localhost:8080/health
```

---

## Task 2: .env File Support

### What Was Added

The application now automatically loads environment variables from a `.env` file in the project root during development.

### Changes Made

**File: `/Users/danwilliams/Documents/web3/gatekeeper/cmd/server/main.go`**

```go
import (
    // ... other imports
    "github.com/joho/godotenv"
)

func main() {
    // Load .env file for development (ignore error if file doesn't exist)
    _ = godotenv.Load()

    // Load configuration
    cfg, err := config.Load()
    // ...
}
```

### Creating a .env File

Create a `.env` file in `/Users/danwilliams/Documents/web3/gatekeeper/`:

```bash
# Required variables
DATABASE_URL=postgres://gatekeeper:gatekeeper@localhost:5432/gatekeeper?sslmode=disable
JWT_SECRET=your-randomly-generated-secret-key-here-minimum-32-characters
ETHEREUM_RPC=https://eth-sepolia.g.alchemy.com/v2/demo
PORT=8080

# Optional variables
ENVIRONMENT=development
LOG_LEVEL=debug
CHAIN_ID=11155111
```

### Testing .env File Loading

**Step 1: Create test .env file**

```bash
cd /Users/danwilliams/Documents/web3/gatekeeper
cat > .env << 'EOF'
DATABASE_URL=postgres://user:pass@localhost/testdb
JWT_SECRET=test-secret-key-for-development-only-minimum-32-chars
ETHEREUM_RPC=https://eth-sepolia.g.alchemy.com/v2/demo
PORT=9999
LOG_LEVEL=debug
EOF
```

**Step 2: Build and run**

```bash
go build -o gatekeeper ./cmd/server
./gatekeeper
```

**Step 3: Verify**

The application should:
- Load variables from `.env` file
- Start on port 9999 (from .env)
- Use debug log level (from .env)
- Log: "Starting Gatekeeper (port 9999)"

**Step 4: Test override with environment variable**

```bash
# Environment variables override .env file
PORT=7777 ./gatekeeper
# Should start on port 7777, not 9999
```

### Key Features
- ✅ Automatically loads `.env` file on startup
- ✅ Silent failure if `.env` doesn't exist (production-friendly)
- ✅ Environment variables override `.env` values
- ✅ Works with existing config system
- ✅ No code changes needed in config package

### Production Note

In production, don't use `.env` files. Instead:
- Use environment variables directly
- Use secrets management (AWS Secrets Manager, HashiCorp Vault)
- Use orchestration tools (Kubernetes ConfigMaps/Secrets)

---

## Task 3: Environment Variables Documentation

### What Was Added

Comprehensive environment variables documentation in the README.md file.

### Location

`/Users/danwilliams/Documents/web3/gatekeeper/README.md` - Section: "Configuration > Environment Variables"

### What's Documented

#### Required Variables
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret key for signing JWTs (min 32 chars)
- `ETHEREUM_RPC` - Primary Ethereum RPC provider URL
- `PORT` - HTTP server port

#### Optional Variables (17 total)
- `ENVIRONMENT` - Environment mode (development/staging/production)
- `LOG_LEVEL` - Logging level (debug/info/warn/error)
- `ETHEREUM_RPC_FALLBACK` - Fallback RPC endpoint
- `CHAIN_ID` - Blockchain chain ID
- `CACHE_TTL` - Cache time-to-live in seconds
- `RPC_TIMEOUT` - RPC call timeout in seconds
- `JWT_EXPIRY_HOURS` - JWT token expiration
- `NONCE_TTL_MINUTES` - Nonce expiration
- Database pool settings (4 variables)
- Rate limiting settings (4 variables)

### Features
- ✅ Clear table format with descriptions
- ✅ Type information for each variable
- ✅ Default values listed
- ✅ Complete example `.env` file
- ✅ Security best practices
- ✅ Commands to generate secure secrets

### Example .env File Template

The README includes a complete, copy-paste ready `.env` template:

```bash
# Required
DATABASE_URL=postgres://gatekeeper:gatekeeper@localhost:5432/gatekeeper?sslmode=disable
JWT_SECRET=your-randomly-generated-secret-key-here-minimum-32-characters
ETHEREUM_RPC=https://eth-sepolia.g.alchemy.com/v2/demo
PORT=8080

# Optional
ENVIRONMENT=development
LOG_LEVEL=debug
CHAIN_ID=11155111
CACHE_TTL=300
RPC_TIMEOUT=5
JWT_EXPIRY_HOURS=24
NONCE_TTL_MINUTES=5

# Database Pool
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME_MINUTES=5
DB_CONN_MAX_IDLE_TIME_MINUTES=1

# Rate Limiting
API_KEY_CREATION_RATE_LIMIT=10
API_KEY_CREATION_BURST_LIMIT=3
API_USAGE_RATE_LIMIT=1000
API_USAGE_BURST_LIMIT=100
```

### Verification

To verify the documentation:

```bash
# Check the README has the Environment Variables section
grep -A 5 "### Environment Variables" /Users/danwilliams/Documents/web3/gatekeeper/README.md

# Verify required variables are documented
grep "DATABASE_URL\|JWT_SECRET\|ETHEREUM_RPC\|PORT" /Users/danwilliams/Documents/web3/gatekeeper/README.md
```

---

## Verification Commands

### Complete Verification

Run the automated test suite:

```bash
cd /Users/danwilliams/Documents/web3/gatekeeper
./test_phase1.sh
```

### Manual Verification

**1. Check Database Health Function**

```bash
# Verify function signature
grep -A 10 "CheckDatabaseHealth" /Users/danwilliams/Documents/web3/gatekeeper/internal/store/db.go

# Verify it compiles
go test -c ./internal/store -o /dev/null && echo "✅ Compiles" || echo "❌ Failed"
```

**2. Check .env Support**

```bash
# Verify godotenv is installed
grep "godotenv" /Users/danwilliams/Documents/web3/gatekeeper/go.mod

# Verify it's imported and used
grep -n "godotenv" /Users/danwilliams/Documents/web3/gatekeeper/cmd/server/main.go
```

**3. Check Documentation**

```bash
# View environment variables section
sed -n '/^## Configuration/,/^## API Endpoints/p' /Users/danwilliams/Documents/web3/gatekeeper/README.md
```

**4. Build Test**

```bash
# Verify project builds with all changes
go build -o gatekeeper ./cmd/server && echo "✅ Build successful" || echo "❌ Build failed"
```

---

## Summary

All Phase 1 tasks have been completed successfully:

### ✅ Task 1: Database Health Check
- **File Modified**: `/Users/danwilliams/Documents/web3/gatekeeper/internal/store/db.go`
- **Function Added**: `CheckDatabaseHealth(ctx context.Context, db *sql.DB) error`
- **Features**: 5-second timeout, simple SELECT 1 query, detailed error handling

### ✅ Task 2: .env File Support
- **File Modified**: `/Users/danwilliams/Documents/web3/gatekeeper/cmd/server/main.go`
- **Package Added**: `github.com/joho/godotenv v1.5.1`
- **Features**: Automatic .env loading, silent failure, environment override support

### ✅ Task 3: Environment Variables Documentation
- **File Modified**: `/Users/danwilliams/Documents/web3/gatekeeper/README.md`
- **Content Added**: Complete env vars table, example .env file, security notes
- **Coverage**: 4 required + 17 optional variables fully documented

### Build Status
✅ All changes compile successfully
✅ Project builds without errors
✅ No breaking changes introduced

### Next Steps

To use these features:

1. **Create a .env file** in the project root with your configuration
2. **Test database health** by accessing the `/health` endpoint
3. **Refer to README** for complete configuration options

For production deployment, use environment variables directly instead of .env files.
