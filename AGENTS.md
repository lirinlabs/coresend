# CoreSend Agent Documentation

This is the main entry point for AI agents working on the CoreSend project. Below you'll find quick reference information and links to detailed documentation.

## Project Overview

CoreSend is a stateless temporary email service built with Go, using BIP39 mnemonic phrases for identity. The architecture follows clean architecture principles with containerized deployment via Docker Compose.

### Quick Tech Stack
- **Backend**: Go 1.25.6, SMTP server, HTTP API
- **Storage**: Redis 7 (24-hour TTL email storage)
- **Web Server**: Caddy (automatic HTTPS, reverse proxy)
- **Infrastructure**: Docker, Docker Compose, Terraform (OCI)

### Key Concepts
- **Mnemonic-based identity**: 12-word BIP39 phrases → 16-char hex addresses
- **Stateless design**: No user database, deterministic address generation
- **Clean architecture**: Internal packages with clear separation of concerns
- **Container-first**: Docker Compose deployment with production/development configs

## Documentation Structure

| File | Purpose | Size |
|------|---------|------|
| [DEVELOPMENT.md](DEVELOPMENT.md) | Development commands, setup, and workflow | ~100 lines |
| [CODING_STANDARDS.md](CODING_STANDARDS.md) | Go coding conventions and patterns | ~80 lines |
| [TESTING.md](TESTING.md) | Test structure, patterns, and requirements | ~120 lines |
| [SECURITY.md](SECURITY.md) | Security best practices and configuration | ~100 lines |

## Quick Reference

### Essential Commands
```bash
# Start development
docker-compose up -d

# Run tests
cd backend && go test -short ./...

# Production deployment
docker-compose -f docker-compose.yml up -d
```

### Project Structure
```
backend/
├── cmd/server/           # Main entry point
├── internal/
│   ├── identity/         # BIP39 mnemonic generation
│   ├── smtp/             # SMTP server implementation
│   └── store/            # Redis storage operations
└── TESTING.md           # Testing guide
```

### Key Files to Understand
- `backend/cmd/server/main.go` - Application entry point and configuration
- `backend/internal/identity/generator.go` - Mnemonic/address logic
- `backend/internal/smtp/backend.go` - SMTP protocol handling
- `backend/internal/store/redis.go` - Redis operations

## Getting Started

1. **For development setup**: See [DEVELOPMENT.md](DEVELOPMENT.md)
2. **For coding patterns**: See [CODING_STANDARDS.md](CODING_STANDARDS.md)
3. **For testing guidelines**: See [TESTING.md](TESTING.md)
4. **For security requirements**: See [SECURITY.md](SECURITY.md)

## Agent Guidelines

- Always follow the coding standards in [CODING_STANDARDS.md](CODING_STANDARDS.md)
- Write tests according to patterns in [TESTING.md](TESTING.md)
- Consider security implications as outlined in [SECURITY.md](SECURITY.md)
- Use the development workflow from [DEVELOPMENT.md](DEVELOPMENT.md)

## Additional Resources

- [README.md](README.md) - Main project documentation
- [DOCKER_SETUP.md](DOCKER_SETUP.md) - Docker deployment guide
- [backend/TESTING.md](backend/TESTING.md) - Detailed testing documentation

---

**Note**: This file serves as an index. For detailed information on specific topics, please refer to the linked documentation files above.