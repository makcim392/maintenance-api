.PHONY: all build run test clean fmt lint tidy

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
	go test -v ./...

# Clean build artifacts
clean:
	go clean
	rm -f ./bin/*

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