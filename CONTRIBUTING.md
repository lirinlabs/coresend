# Contributing to CoreSend

Want to help build CoreSend? Here's how to get started.

## Setting up for development

### Prerequisites
- Go 1.25+
- Docker & Docker Compose
- Redis (for local testing)

### Local development

```bash
# Clone the repo
git clone https://github.com/yourusername/coresend.git
cd coresend

# Start dependencies
docker-compose up -d redis

# Run the server
cd backend
go run cmd/server/main.go
```

The service runs on:
- SMTP: localhost:2525 (development)
- HTTP: localhost:8080

### Running tests

```bash
# All tests
go test ./...

# Short tests (skip integration)
go test -short ./...

# With coverage
go test -cover ./...
```

## Code style

We follow Go conventions with a few preferences:

```go
// Prefer clarity over cleverness
func generateAddress(mnemonic string) (string, error) {
    // Clear variable names, no abbreviations
    hash := sha256.Sum256([]byte(mnemonic))
    return hex.EncodeToString(hash[:8]), nil
}

// Error messages should be helpful
if len(mnemonic) != 12 {
    return "", fmt.Errorf("mnemonic must be exactly 12 words, got %d", len(words))
}
```

## Project structure

```
backend/
├── cmd/server/          # Application entry point
├── internal/
│   ├── identity/        # BIP39 → address conversion
│   ├── smtp/           # SMTP protocol handling  
│   └── store/          # Redis operations
└── testing/            # Test utilities
```

## Making changes

1. Fork and create a feature branch
2. Write tests for new functionality
3. Keep changes focused and small
4. Run the full test suite
5. Submit a pull request

### Testing guidelines

- Unit tests for business logic
- Integration tests for external services
- Mock Redis in unit tests
- Use table-driven tests for multiple cases

Example test:
```go
func TestGenerateAddress(t *testing.T) {
    tests := []struct {
        name     string
        mnemonic string
        want     string
    }{
        {"valid mnemonic", "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about", "7c7e7c7e7c7e7c7e"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := GenerateAddress(tt.mnemonic)
            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Security considerations

- Never log mnemonics or email addresses
- Validate all input thoroughly
- Use timeouts for external calls
- Rate limit API endpoints
- Keep Redis access restricted

## Getting help

- Check existing issues first
- Look at the [API docs](docs/API.md) for endpoint details
- Review the [architecture](ARCHITECTURE.md) for design decisions
- Ask questions in your pull request

Thanks for contributing!