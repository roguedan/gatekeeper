#!/bin/bash
# Docker Compose Validation Script for Gatekeeper

set -e

echo "========================================="
echo "Gatekeeper Docker Compose Validation"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if Docker is running
if ! docker info &> /dev/null; then
    echo -e "${RED}❌ Docker is not running${NC}"
    echo "Please start Docker and try again"
    exit 1
fi

echo -e "${GREEN}✓ Docker is running${NC}"
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠ .env file not found, creating from .env.example${NC}"
    cp .env.example .env
    echo -e "${YELLOW}⚠ Please update .env with your configuration before proceeding${NC}"
    exit 1
fi

echo -e "${GREEN}✓ .env file exists${NC}"
echo ""

# Validate docker-compose.yml
echo -e "${BLUE}Validating docker-compose.yml...${NC}"
if docker compose config > /dev/null 2>&1; then
    echo -e "${GREEN}✓ docker-compose.yml is valid${NC}"
else
    echo -e "${RED}❌ docker-compose.yml is invalid${NC}"
    docker compose config
    exit 1
fi
echo ""

# Start services
echo "========================================="
echo "Starting Services"
echo "========================================="
echo ""

echo -e "${BLUE}Starting all services...${NC}"
docker compose up -d

echo ""
echo -e "${BLUE}Waiting for services to be healthy (max 60s)...${NC}"

# Wait for services to be healthy
TIMEOUT=60
ELAPSED=0
INTERVAL=5

while [ $ELAPSED -lt $TIMEOUT ]; do
    HEALTHY_COUNT=$(docker compose ps --format json | jq -r '.Health // "starting"' | grep -c "healthy" || echo "0")
    TOTAL_COUNT=$(docker compose ps --format json | wc -l)

    echo -e "${BLUE}Health check: ${HEALTHY_COUNT}/${TOTAL_COUNT} services healthy${NC}"

    if [ "$HEALTHY_COUNT" -eq "$TOTAL_COUNT" ]; then
        echo -e "${GREEN}✓ All services are healthy${NC}"
        break
    fi

    sleep $INTERVAL
    ELAPSED=$((ELAPSED + INTERVAL))
done

echo ""

# Show service status
echo "========================================="
echo "Service Status"
echo "========================================="
echo ""
docker compose ps

echo ""
echo "========================================="
echo "Health Checks"
echo "========================================="
echo ""

# Test backend health
echo -e "${BLUE}Testing backend health...${NC}"
if curl -f -s http://localhost:8080/health > /dev/null 2>&1; then
    RESPONSE=$(curl -s http://localhost:8080/health)
    echo -e "${GREEN}✓ Backend is healthy: ${RESPONSE}${NC}"
else
    echo -e "${RED}❌ Backend health check failed${NC}"
    echo "Backend logs:"
    docker compose logs backend | tail -20
fi
echo ""

# Test frontend health
echo -e "${BLUE}Testing frontend health...${NC}"
if curl -f -s http://localhost:3000/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Frontend is healthy${NC}"
else
    echo -e "${YELLOW}⚠ Frontend health check failed (may not be built yet)${NC}"
fi
echo ""

# Test PostgreSQL
echo -e "${BLUE}Testing PostgreSQL...${NC}"
if docker compose exec -T postgres pg_isready -U gatekeeper > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PostgreSQL is ready${NC}"
else
    echo -e "${RED}❌ PostgreSQL is not ready${NC}"
fi
echo ""

# Test Redis
echo -e "${BLUE}Testing Redis...${NC}"
if docker compose exec -T redis redis-cli ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Redis is responding${NC}"
else
    echo -e "${RED}❌ Redis is not responding${NC}"
fi
echo ""

# Check database tables
echo -e "${BLUE}Checking database tables...${NC}"
TABLE_COUNT=$(docker compose exec -T postgres psql -U gatekeeper -d gatekeeper -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public';" 2>/dev/null | tr -d ' ' || echo "0")

if [ "$TABLE_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Database has ${TABLE_COUNT} tables (migrations successful)${NC}"
    docker compose exec -T postgres psql -U gatekeeper -d gatekeeper -c "\dt"
else
    echo -e "${YELLOW}⚠ Database has no tables (migrations may not have run)${NC}"
fi
echo ""

# Show resource usage
echo "========================================="
echo "Resource Usage"
echo "========================================="
echo ""
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"

echo ""
echo "========================================="
echo "Validation Complete"
echo "========================================="
echo ""

echo "Services are running. Access points:"
echo -e "${BLUE}Frontend:${NC} http://localhost:3000"
echo -e "${BLUE}Backend API:${NC} http://localhost:8080"
echo -e "${BLUE}Backend Health:${NC} http://localhost:8080/health"
echo ""

echo "Useful commands:"
echo "  View logs:           docker compose logs -f"
echo "  View backend logs:   docker compose logs -f backend"
echo "  Stop services:       docker compose stop"
echo "  Restart services:    docker compose restart"
echo "  Clean shutdown:      docker compose down"
echo "  Remove all data:     docker compose down -v"
echo ""

# Ask if user wants to view logs
read -p "View logs? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    docker compose logs -f
fi
