.PHONY: all build run test clean fmt lint tidy test-cover test-cover-html

# Default target
all: clean fmt lint test build

# Build the application
build:
	go build -v ./...

# Run the application
run:
	go run ./...

# Run tests
test:
	go test -v -cover ./...

# Run tests with coverage and generate HTML report
test-cover:
	go test -coverprofile=coverage.out ./...
	@echo "Total Coverage:"
	@go tool cover -func=coverage.out | grep total | awk '{print $$3}'

# Generate and open HTML coverage report in browser
test-cover-html: test-cover
	go tool cover -html=coverage.out -o coverage.html
	open "$$PWD/coverage.html"

# Clean build artifacts
clean:
	go clean
	rm -f ./bin/*
	rm -f coverage.out coverage.html

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Run linter
lint:
	golangci-lint run

# Tidy and verify dependencies
tidy:
	go mod tidy
	go mod verify

# Install development tools
tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
