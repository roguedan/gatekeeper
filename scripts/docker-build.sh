#!/bin/bash
# Docker Build Test Script for Gatekeeper

set -e

echo "========================================="
echo "Gatekeeper Docker Build Test"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    echo "Please install Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

echo -e "${GREEN}✓ Docker is installed${NC}"
docker --version
echo ""

# Check if Docker Compose is installed
if ! command -v docker compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed${NC}"
    echo "Please install Docker Compose: https://docs.docker.com/compose/install/"
    exit 1
fi

echo -e "${GREEN}✓ Docker Compose is installed${NC}"
docker compose version
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠ .env file not found, creating from .env.example${NC}"
    cp .env.example .env
    echo -e "${YELLOW}⚠ Please update .env with your configuration${NC}"
    echo ""
fi

echo "========================================="
echo "Building Backend Image"
echo "========================================="
echo ""

START_TIME=$(date +%s)

# Build backend image
if docker build -t gatekeeper-backend:test -f Dockerfile .; then
    END_TIME=$(date +%s)
    BUILD_TIME=$((END_TIME - START_TIME))
    echo -e "${GREEN}✓ Backend build successful (${BUILD_TIME}s)${NC}"

    # Get image size
    SIZE=$(docker images gatekeeper-backend:test --format "{{.Size}}")
    echo -e "${GREEN}✓ Image size: ${SIZE}${NC}"
    echo ""
else
    echo -e "${RED}❌ Backend build failed${NC}"
    exit 1
fi

echo "========================================="
echo "Building Frontend Image"
echo "========================================="
echo ""

# Check if frontend directory exists
if [ ! -d "web" ]; then
    echo -e "${YELLOW}⚠ Frontend directory (web/) not found${NC}"
    echo -e "${YELLOW}⚠ Skipping frontend build${NC}"
    echo ""
else
    START_TIME=$(date +%s)

    # Build frontend image
    if docker build -t gatekeeper-frontend:test -f web/Dockerfile web/; then
        END_TIME=$(date +%s)
        BUILD_TIME=$((END_TIME - START_TIME))
        echo -e "${GREEN}✓ Frontend build successful (${BUILD_TIME}s)${NC}"

        # Get image size
        SIZE=$(docker images gatekeeper-frontend:test --format "{{.Size}}")
        echo -e "${GREEN}✓ Image size: ${SIZE}${NC}"
        echo ""
    else
        echo -e "${RED}❌ Frontend build failed${NC}"
        exit 1
    fi
fi

echo "========================================="
echo "Build Summary"
echo "========================================="
echo ""

docker images | grep gatekeeper | grep test

echo ""
echo -e "${GREEN}✓ All builds completed successfully!${NC}"
echo ""
echo "Next steps:"
echo "1. Start services: docker compose up -d"
echo "2. View logs: docker compose logs -f"
echo "3. Check health: docker compose ps"
echo "4. Access frontend: http://localhost:3000"
echo "5. Access backend: http://localhost:8080/health"
