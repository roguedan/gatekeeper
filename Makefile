.PHONY: help build test test-verbose test-coverage run clean install-tools fmt lint docker-build docker-up docker-down docker-logs docker-ps docker-clean

help:
	@echo "Gatekeeper - Authentication Gateway"
	@echo ""
	@echo "Available targets:"
	@echo "  build              Build the binary"
	@echo "  run                Run the server locally"
	@echo "  test               Run tests"
	@echo "  test-verbose       Run tests with verbose output"
	@echo "  test-coverage      Run tests with coverage report"
	@echo "  coverage-html      Generate HTML coverage report"
	@echo "  clean              Remove build artifacts"
	@echo "  install-tools      Install development tools"
	@echo "  fmt                Format code"
	@echo "  lint               Run linter"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-build       Build Docker images"
	@echo "  docker-up          Start all services with Docker Compose"
	@echo "  docker-down        Stop all services"
	@echo "  docker-logs        View logs from all services"
	@echo "  docker-ps          Show running containers"
	@echo "  docker-clean       Remove all containers, images, and volumes"
	@echo "  docker-validate    Validate Docker setup and test services"

build:
	go build -o bin/gatekeeper ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./... -v

test-verbose:
	go test ./... -v -race

test-coverage:
	go test ./... -coverprofile=coverage.txt -covermode=atomic

coverage-html: test-coverage
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean:
	rm -rf bin/ dist/ coverage.txt coverage.html

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

# Docker targets
docker-build:
	@echo "Building Docker images..."
	@./scripts/docker-build.sh

docker-up:
	@echo "Starting services with Docker Compose..."
	@if [ ! -f .env ]; then \
		echo "Creating .env from .env.example..."; \
		cp .env.example .env; \
		echo "Please update .env with your configuration"; \
	fi
	docker compose up -d
	@echo "Services started. Run 'make docker-ps' to check status"

docker-down:
	@echo "Stopping services..."
	docker compose down

docker-logs:
	docker compose logs -f

docker-ps:
	docker compose ps

docker-clean:
	@echo "Removing all containers, images, and volumes..."
	@read -p "Are you sure? This will delete all data. [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker compose down -v --rmi all; \
		echo "Cleanup complete"; \
	else \
		echo "Cancelled"; \
	fi

docker-validate:
	@echo "Validating Docker setup..."
	@./scripts/docker-validate.sh

docker-restart:
	@echo "Restarting services..."
	docker compose restart

docker-backend-logs:
	docker compose logs -f backend

docker-frontend-logs:
	docker compose logs -f frontend
