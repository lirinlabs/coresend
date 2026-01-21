#!/bin/bash

# Test script for coresend backend
# This script runs all unit tests and integration tests

set -e

echo "🧪 Running tests for coresend backend..."

# Check if Redis is available for integration tests
if ! redis-cli ping > /dev/null 2>&1; then
    echo "⚠️  Redis not available - integration tests will be skipped"
    echo "💡 Start Redis with: docker run -d -p 6379:6379 redis:alpine"
fi

# Run unit tests
echo "📋 Running unit tests..."
go test -v ./internal/identity/...
go test -v ./internal/smtp/...

# Run integration tests (will skip if Redis not available)
echo "🔗 Running integration tests..."
go test -v ./internal/store/... -short=false

echo "✅ All tests completed!"