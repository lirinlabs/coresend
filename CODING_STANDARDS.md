# CoreSend Coding Standards

This document outlines the coding standards and conventions for the CoreSend project.

## Go Standards

### General Guidelines
- Follow official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for code formatting
- Use `goimports` for import organization
- Keep lines under 120 characters when possible
- Use meaningful variable names, avoid abbreviations

### Naming Conventions

#### Exported vs Private
```go
// Exported (public) - PascalCase
type EmailStore interface { ... }
func (s *Store) SaveEmail() error { ... }
const AddressLength = 16

// Private (internal) - camelCase
type store struct { ... }
func (s *store) saveEmail() error { ... }
const defaultTTL = 24 * time.Hour
```

#### Package Names
- Simple, lowercase, single words
- No underscores or mixed case
- Descriptive but concise
```go
package identity    // Good
package smtp        // Good
package store       // Good
package email_store // Bad - too long
package EmailStore  // Bad - mixed case
```

#### Interface Naming
- Interfaces end with `er` suffix when possible
- Simple, descriptive names
```go
type EmailStore interface { ... }  // Good
type Saver interface { ... }       // Too generic
type StoreInterface interface { ... } // Redundant
```

#### Function Names
```go
// Constructor functions - NewTypeName()
func NewStore(addr, password string) *Store { ... }

// Getter functions - TypeName() or GetTypeName()
func (e *Email) Subject() string { ... }
func (s *Store) GetEmails() []Email { ... }

// Boolean functions - IsTypeName() or HasTypeName()
func IsValidAddress(addr string) bool { ... }
func (s *Session) HasRecipients() bool { ... }
```

### Import Organization

#### Import Groups
```go
import (
    // Standard library
    "context"
    "crypto/tls"
    "log"
    "os"
    "time"
    
    // Third-party packages
    "github.com/emersion/go-smtp"
    "github.com/redis/go-redis/v9"
    "github.com/tyler-smith/go-bip39"
    
    // Internal packages
    "github.com/fn-jakubkarp/coresend/internal/identity"
    "github.com/fn-jakubkarp/coresend/internal/store"
)
```

#### Import Aliases
```go
import (
    gosmtp "github.com/emersion/go-smtp"  // Avoid conflict
    "github.com/fn-jakubkarp/coresend/internal/store"
)
```

### Error Handling

#### Standard Error Pattern
```go
func (s *Store) SaveEmail(ctx context.Context, inbox string, email Email) error {
    if inbox == "" {
        return fmt.Errorf("inbox cannot be empty")
    }
    
    if err := s.client.LPush(ctx, key, email).Err(); err != nil {
        return fmt.Errorf("failed to save email: %w", err)
    }
    
    return nil
}
```

#### Error Wrapping
```go
// Use %w for error wrapping to preserve stack traces
return fmt.Errorf("operation failed: %w", err)

// Use %v for simple error messages
return fmt.Errorf("invalid input: %v", input)
```

#### Error Logging
```go
// Log errors with context
log.Printf("Error saving email for %s: %v", recipient, err)

// Don't log sensitive data
log.Printf("Authentication failed for user")  // Good
log.Printf("Authentication failed for user %s", username)  // Bad
```

### Constants and Variables

#### Constants
```go
const (
    AddressLength = 16
    DefaultTTL    = 24 * time.Hour
    MaxEmails     = 100
    MaxEmailSize  = 1024 * 1024  // 1MB
)

// Configuration constants
const (
    DefaultRedisAddr = "localhost:6379"
    DefaultSMTPPort   = ":1025"
)
```

#### Variables
```go
var (
    validAddressRegex = regexp.MustCompile(`^[a-f0-9]{16}$`)
    errInvalidAddress = errors.New("invalid address format")
)
```

### Struct Definitions

#### Field Ordering
```go
type Email struct {
    // Public fields (exported)
    From       string    `json:"from"`
    To         []string  `json:"to"`
    Subject    string    `json:"subject"`
    Body       string    `json:"body"`
    ReceivedAt time.Time `json:"received_at"`
    
    // Private fields (unexported)
    id     string
    stored bool
}
```

#### Constructor Pattern
```go
func NewEmail(from, to, subject, body string) *Email {
    return &Email{
        From:       from,
        To:         []string{to},
        Subject:    subject,
        Body:       body,
        ReceivedAt: time.Now(),
    }
}
```

### Function Design

#### Parameter Order
```go
// Context first
func (s *Store) SaveEmail(ctx context.Context, inbox string, email Email) error { ... }

// Configuration options last
func NewServer(addr string, opts ...ServerOption) *Server { ... }
```

#### Return Values
```go
// Single value
func (e *Email) Subject() string { ... }

// Error as last return value
func (s *Store) GetEmails(ctx context.Context, inbox string) ([]Email, error) { ... }

// Multiple return values
func ParseAddress(email string) (local, domain string, err error) { ... }
```

### Comments and Documentation

#### Package Documentation
```go
// Package identity provides BIP39 mnemonic generation and deterministic
// address derivation for the CoreSend temporary email service.
package identity
```

#### Exported Function Documentation
```go
// AddressFromMnemonic derives a deterministic 16-character hexadecimal
// address from a BIP39 mnemonic phrase using SHA-256 hashing.
// The mnemonic is trimmed and lowercased before processing.
func AddressFromMnemonic(mnemonic string) string { ... }
```

#### Complex Logic Comments
```go
// Prefer HTML over plain text when both are present in multipart emails
if contentType == "text/html" {
    email.Body = string(body)
} else if contentType == "text/plain" && email.Body == "" {
    email.Body = string(body)
}
```

### Code Organization

#### File Structure
```
internal/
├── identity/
│   ├── generator.go      # Main functionality
│   └── generator_test.go # Tests
├── smtp/
│   ├── backend.go        # SMTP backend
│   ├── session.go        # SMTP session (if split)
│   └── backend_test.go   # Tests
└── store/
    ├── redis.go          # Redis implementation
    └── redis_test.go     # Tests
```

#### Function Organization
```go
// Public functions first
func NewStore(addr, password string) *Store { ... }
func (s *Store) SaveEmail() error { ... }
func (s *Store) GetEmails() []Email { ... }

// Private functions last
func (s *store) validateEmail() error { ... }
func (s *store) serializeEmail() []byte { ... }
```

### Environment Variables

#### Getter Pattern
```go
func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}

func main() {
    redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
    domain := getEnv("DOMAIN_NAME", "localhost")
}
```

#### Configuration Struct
```go
type Config struct {
    RedisAddr     string
    RedisPassword string
    Domain         string
    SMTPListenAddr string
    CertPath       string
    KeyPath        string
}

func LoadConfig() *Config {
    return &Config{
        RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
        RedisPassword: os.Getenv("REDIS_PASSWORD"),
        Domain:        getEnv("DOMAIN_NAME", "localhost"),
        SMTPListenAddr: getEnv("SMTP_LISTEN_ADDR", ":1025"),
        CertPath:      os.Getenv("SMTP_CERT_PATH"),
        KeyPath:       os.Getenv("SMTP_KEY_PATH"),
    }
}
```

### Testing Patterns

#### Test Naming
```go
func TestFunctionName(t *testing.T) { ... }
func TestFunctionName_EdgeCase(t *testing.T) { ... }
func TestFunctionName_ErrorCondition(t *testing.T) { ... }
```

#### Table-Driven Tests
```go
func TestAddressFromMnemonic(t *testing.T) {
    tests := []struct {
        name     string
        mnemonic string
        expected string
    }{
        {"valid mnemonic", "witch collapse practice...", "b4ebe3e2200cbc90"},
        {"empty string", "", ""},
        {"whitespace", "  test  ", "test"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := AddressFromMnemonic(tt.mnemonic)
            if result != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

### Performance Considerations

#### String Building
```go
// Bad - creates multiple strings
result := prefix + "-" + suffix

// Good - efficient string building
var builder strings.Builder
builder.WriteString(prefix)
builder.WriteByte('-')
builder.WriteString(suffix)
result := builder.String()
```

#### Slice Operations
```go
// Pre-allocate when size is known
emails := make([]Email, 0, expectedCount)

// Use append for dynamic growth
emails = append(emails, newEmail)
```

#### Context Usage
```go
// Pass context through the call chain
func (s *Store) SaveEmail(ctx context.Context, inbox string, email Email) error {
    return s.client.LPush(ctx, key, email).Err()
}
```

## Linting and Tools

### Recommended Tools
```bash
# Install tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
```

### golangci-lint Configuration
```yaml
# .golangci.yml
run:
  timeout: 5m

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck

linters-settings:
  goimports:
    local-prefixes: github.com/fn-jakubkarp/coresend
```

### Pre-commit Hooks
```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.4.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
  
  - repo: https://github.com/golangci/golangci-lint
    rev: v1.54.0
    hooks:
      - id: golangci-lint
        args: [--timeout=5m]
```

## Code Review Checklist

### Functionality
- [ ] Code implements requirements correctly
- [ ] Error handling is comprehensive
- [ ] Edge cases are considered
- [ ] Tests cover main scenarios

### Style
- [ ] Code follows Go conventions
- [ ] Names are descriptive and consistent
- [ ] Comments explain complex logic
- [ ] Imports are properly organized

### Performance
- [ ] No obvious performance issues
- [ ] Resources are properly managed
- [ ] Context is used appropriately
- [ ] Memory allocation is reasonable

### Security
- [ ] No sensitive data in logs
- [ ] Input validation is present
- [ ] Error messages don't leak information
- [ ] Dependencies are up to date

---

Following these coding standards ensures consistency, maintainability, and quality across the CoreSend codebase. For testing patterns, see [TESTING.md](TESTING.md). For security guidelines, see [SECURITY.md](SECURITY.md).