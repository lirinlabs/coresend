# CoreSend Development Guide

This guide covers development setup, commands, and workflow for the CoreSend project.

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.25.6 (for local development)
- Git

### Development Setup
```bash
# Clone the repository
git clone https://github.com/fn-jakubkarp/coresend.git
cd coresend

# Start development environment
docker-compose up -d

# View logs
docker-compose logs -f
```

## Development Commands

### Local Development
```bash
# Start all services (uses docker-compose.override.yml automatically)
docker-compose up -d

# Start specific services
docker-compose up -d redis backend

# View logs
docker-compose logs -f
docker-compose logs -f backend

# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

### Backend Development
```bash
# Navigate to backend directory
cd backend

# Build the application
go build -o main ./cmd/server/main.go

# Run locally (requires Redis)
go run ./cmd/server/main.go

# Install dependencies
go mod download
go mod tidy
```

### Testing
```bash
# Navigate to backend directory
cd backend

# Unit tests only (no Redis required)
go test -short ./...

# Full test suite (requires Redis)
docker-compose up -d redis
go test ./...

# Individual package tests
go test ./internal/identity/...    # No external deps
go test ./internal/smtp/...       # Uses mocks
go test ./internal/store/...      # Requires Redis

# Verbose output
go test -v ./...

# Test coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out

# Run all tests via script
./test.sh
```

### Production Deployment
```bash
# Production deployment (ignore override file)
docker-compose -f docker-compose.yml up -d

# View production logs
docker-compose -f docker-compose.yml logs -f

# Scale services
docker-compose -f docker-compose.yml up -d --scale backend=2

# Infrastructure deployment (OCI)
cd terraform
terraform init
terraform apply
```

## Environment Configuration

### Development Environment
Uses `docker-compose.override.yml` automatically:
- HTTP only (no SSL complexity)
- Direct API port exposure (8080)
- Shared Redis instance
- Debug logging enabled

### Production Environment
Uses `docker-compose.yml` only:
- Automatic HTTPS via Let's Encrypt
- API proxied through Caddy
- Isolated services with health checks
- Optimized logging

### Environment Variables
```bash
# Required for production
DOMAIN_NAME=coresend.io

# Optional Redis security
REDIS_PASSWORD=your-secure-password

# Development defaults
REDIS_ADDR=localhost:6379
SMTP_LISTEN_ADDR=:1025
SMTP_CERT_PATH=/certs/${DOMAIN_NAME}/${DOMAIN_NAME}.crt
SMTP_KEY_PATH=/certs/${DOMAIN_NAME}/${DOMAIN_NAME}.key
```

## Workflow

### Development Workflow
1. **Start services**: `docker-compose up -d`
2. **Make changes**: Edit code in `backend/`
3. **Run tests**: `cd backend && go test -short ./...`
4. **Test changes**: Use Docker Compose services
5. **Stop services**: `docker-compose down`

### Git Workflow
```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Make changes and commit
git add .
git commit -m "feat: add new feature"

# Push and create PR
git push origin feature/your-feature-name
```

### Debugging
```bash
# Check service status
docker-compose ps

# View service logs
docker-compose logs -f backend

# Test Redis connection
docker-compose exec redis redis-cli ping

# Test SMTP server
telnet localhost 25

# Enter backend container
docker-compose exec backend sh
```

## Common Development Tasks

### Adding New Dependencies
```bash
# Add Go dependency
cd backend
go get github.com/example/package
go mod tidy

# Update Dockerfile if system packages needed
# Edit backend/Dockerfile
```

### Database Operations
```bash
# Connect to Redis
docker-compose exec redis redis-cli

# View all keys
docker-compose exec redis redis-cli keys "*"

# Clear all data (development)
docker-compose exec redis redis-cli flushall
```

### SMTP Testing
```bash
# Test with swaks
swaks --to b4ebe3e2200cbc90@localhost \
       --from test@example.com \
       --server localhost:25 \
       --body "Test email"

# Test with telnet
telnet localhost 25
HELO test.com
MAIL FROM:<test@example.com>
RCPT TO:<b4ebe3e2200cbc90@localhost>
DATA
Subject: Test

Hello World
.
QUIT
```

## Performance Monitoring

### Resource Usage
```bash
# View container resource usage
docker stats

# View disk usage
docker system df

# Prune unused resources
docker system prune -f
```

### Application Monitoring
```bash
# View backend logs
docker-compose logs -f backend

# Monitor Redis
docker-compose exec redis redis-cli info memory
docker-compose exec redis redis-cli info clients
```

## Troubleshooting

### Common Issues

#### Redis Connection Failed
```bash
# Check Redis status
docker-compose ps redis

# Restart Redis
docker-compose restart redis

# Check logs
docker-compose logs redis
```

#### Port Conflicts
```bash
# Check what's using ports
sudo netstat -tulpn | grep :25
sudo netstat -tulpn | grep :80

# Kill processes using ports
sudo kill -9 <PID>
```

#### Certificate Issues
```bash
# Check DOMAIN_NAME is set
echo $DOMAIN_NAME

# View Caddy logs
docker-compose logs caddy

# Restart Caddy
docker-compose restart caddy
```

### Health Checks
```bash
# Test backend health
curl http://localhost:1025

# Test Caddy health
curl http://localhost/health

# Test Redis
docker-compose exec redis redis-cli ping
```

## Build Optimization

### Development Builds
```bash
# Fast build without optimization
go build -o main ./cmd/server/main.go

# Build with race detection
go build -race -o main ./cmd/server/main.go
```

### Production Builds
```bash
# Optimized build
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main ./cmd/server/main.go

# Build Docker image
docker build -t coresend-backend ./backend
```

## IDE Configuration

### VS Code Settings
```json
{
    "go.useLanguageServer": true,
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v"],
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter",
        "coveredHighlightColor": "rgba(64,128,64,0.5)",
        "uncoveredHighlightColor": "rgba(128,64,64,0.25)"
    }
}
```

### Go Configuration
```bash
# Install useful tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/air-verse/air@latest  # Hot reload
```

## Hot Reload (Optional)

For local development with hot reload:
```bash
# Install air
go install github.com/air-verse/air@latest

# Create .air.toml configuration
# Run with hot reload
air -c .air.toml
```

---

This development guide provides the essential commands and workflows for effective CoreSend development. For coding standards, see [CODING_STANDARDS.md](docs/CODING_STANDARDS.md). For testing guidelines, see [TESTING.md](docs/TESTING.md).