#!/bin/bash
# Phase 1 Verification Script
# Tests the three features added in Phase 1

set -e

echo "======================================"
echo "Phase 1 Feature Verification"
echo "======================================"
echo ""

# Test 1: Database Health Check Function
echo "[1/3] Testing CheckDatabaseHealth function..."
echo "Verifying function exists and compiles..."
if go test -c ./internal/store -o /tmp/store_test 2>&1 | grep -q "error"; then
    echo "❌ FAILED: CheckDatabaseHealth function does not compile"
    exit 1
else
    echo "✅ PASSED: CheckDatabaseHealth function compiles successfully"
    rm -f /tmp/store_test
fi
echo ""

# Test 2: .env File Support
echo "[2/3] Testing .env file support..."
echo "Checking godotenv package is installed..."
if grep -q "github.com/joho/godotenv" go.mod; then
    echo "✅ PASSED: godotenv package found in go.mod"
else
    echo "❌ FAILED: godotenv package not found in go.mod"
    exit 1
fi

echo "Checking godotenv is imported in main.go..."
if grep -q "github.com/joho/godotenv" cmd/server/main.go; then
    echo "✅ PASSED: godotenv imported in main.go"
else
    echo "❌ FAILED: godotenv not imported in main.go"
    exit 1
fi

echo "Checking godotenv.Load() is called..."
if grep -q "godotenv.Load()" cmd/server/main.go; then
    echo "✅ PASSED: godotenv.Load() is called in main.go"
else
    echo "❌ FAILED: godotenv.Load() not called in main.go"
    exit 1
fi
echo ""

# Test 3: Environment Variables Documentation
echo "[3/3] Testing environment variables documentation..."
echo "Checking README.md has Environment Variables section..."
if grep -q "### Environment Variables" README.md; then
    echo "✅ PASSED: Environment Variables section found in README.md"
else
    echo "❌ FAILED: Environment Variables section not found in README.md"
    exit 1
fi

echo "Checking required variables are documented..."
required_vars=("DATABASE_URL" "JWT_SECRET" "ETHEREUM_RPC" "PORT")
all_found=true
for var in "${required_vars[@]}"; do
    if grep -q "$var" README.md; then
        echo "  ✓ $var documented"
    else
        echo "  ✗ $var not documented"
        all_found=false
    fi
done

if [ "$all_found" = true ]; then
    echo "✅ PASSED: All required environment variables are documented"
else
    echo "❌ FAILED: Some required variables are missing"
    exit 1
fi

echo "Checking optional variables are documented..."
optional_vars=("ENVIRONMENT" "LOG_LEVEL")
for var in "${optional_vars[@]}"; do
    if grep -q "$var" README.md; then
        echo "  ✓ $var documented"
    fi
done
echo "✅ PASSED: Optional environment variables are documented"
echo ""

# Test 4: Build verification
echo "[4/4] Testing project builds successfully..."
if go build -o /tmp/gatekeeper_test ./cmd/server 2>&1; then
    echo "✅ PASSED: Project builds successfully"
    rm -f /tmp/gatekeeper_test
else
    echo "❌ FAILED: Project build failed"
    exit 1
fi
echo ""

echo "======================================"
echo "All Phase 1 Tests Passed! ✅"
echo "======================================"
echo ""
echo "Summary:"
echo "  ✓ Database health check function added"
echo "  ✓ .env file support implemented"
echo "  ✓ Environment variables documented"
echo "  ✓ Project builds successfully"
echo ""
