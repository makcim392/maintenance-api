.PHONY: all build run test clean fmt lint tidy test-cover test-cover-html test-integration test-all wait-for-containers

# Default target (excluding integration tests)
all: clean fmt lint test build

# Build the application
build:
	go build -v ./...

# Run the application
run:
	go run ./...

# Run only unit tests (excluding integration tests)
test:
	go test -v -cover $$(go list ./... | grep -v /tests)

# Wait for containers to be healthy
wait-for-containers:
	@echo "Waiting for containers to be healthy..."
	@for i in $$(seq 1 30); do \
		if docker-compose -f tests/docker-compose.test.yml ps | grep -q "healthy"; then \
			echo "Containers are healthy!"; \
			exit 0; \
		fi; \
		echo "Waiting for containers to be ready... ($$i/30)"; \
		sleep 2; \
	done; \
	echo "Container health check timed out"; \
	docker-compose -f tests/docker-compose.test.yml logs; \
	exit 1

# Run integration tests (requires containers)
test-integration:
	docker-compose -f tests/docker-compose.test.yml up -d
	$(MAKE) wait-for-containers
	go test -v ./tests/...
	docker-compose -f tests/docker-compose.test.yml down

# Run all tests (unit + integration)
test-all: test
	docker-compose -f tests/docker-compose.test.yml up -d
	$(MAKE) wait-for-containers
	go test -v ./tests/...
	docker-compose -f tests/docker-compose.test.yml down

# Run tests with coverage (excluding integration tests)
test-cover:
	go test -coverprofile=coverage.out $$(go list ./... | grep -v /tests | grep -v /cmd/api)
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