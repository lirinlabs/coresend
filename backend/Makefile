.PHONY: swagger build run test clean

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/server/main.go --parseDependency --parseInternal; \
	else \
		go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g cmd/server/main.go --parseDependency --parseInternal; \
	fi
	@echo "✓ Swagger docs generated in docs/"

# Build the server binary
build:
	@echo "Building server..."
	@go build -o bin/server cmd/server/main.go
	@echo "✓ Server built: bin/server"

# Run the server (development)
run:
	@go run cmd/server/main.go

# Run tests
test:
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

# Clean build artifacts
clean:
	@rm -rf bin/ docs/*.go docs/*.json docs/*.yaml coverage.out coverage.html
	@echo "✓ Cleaned build artifacts"

# Install dependencies
deps:
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies installed"

# Install swag CLI tool
install-swag:
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✓ swag installed"

# Format code
fmt:
	@go fmt ./...
	@echo "✓ Code formatted"

# Run linter
lint:
	@golangci-lint run ./... || echo "Install golangci-lint: https://golangci-lint.run/usage/install/"

# Full rebuild (clean + swagger + build)
rebuild: clean swagger build

# Help command
help:
	@echo "Available commands:"
	@echo "  make swagger        - Generate Swagger documentation"
	@echo "  make build          - Build the server binary"
	@echo "  make run            - Run the server (development)"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make deps           - Install/update dependencies"
	@echo "  make install-swag   - Install swag CLI tool"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make rebuild        - Clean + regenerate swagger + build"
