.PHONY: help build test test-verbose test-coverage run clean install-tools fmt lint

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
