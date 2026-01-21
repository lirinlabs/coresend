# CoreSend Backend Testing

Guide for running and understanding the backend test suite.

## Test Structure

```
backend/
├── internal/
│   ├── identity/
│   │   └── generator_test.go    # Mnemonic and address tests
│   ├── smtp/
│   │   └── backend_test.go      # SMTP session tests
│   └── store/
│       └── redis_test.go        # Redis integration tests
```

## Running Tests

### Quick Test (Unit Tests Only)

Runs tests without external dependencies (uses mocks):

```bash
cd backend
go test -short ./...
```

### Full Test Suite

Requires Redis running:

```bash
# Start Redis
docker run -d -p 6379:6379 --name redis-test redis:alpine

# Run all tests
cd backend
go test ./...

# Cleanup
docker stop redis-test && docker rm redis-test
```

### Using Docker Compose

```bash
# Start Redis from project root
docker-compose up -d redis

# Run tests
cd backend
go test ./...
```

### Individual Package Tests

```bash
# Identity package (no external deps)
go test ./internal/identity/...

# SMTP package (uses mocks)
go test ./internal/smtp/...

# Store package (requires Redis)
go test ./internal/store/...
```

### Verbose Output

```bash
go test -v ./...
```

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# View in browser
go tool cover -html=coverage.out
```

## Test Coverage by Package

### Identity Package (`internal/identity/`)

| Test | Description |
|------|-------------|
| `TestGenerateNewMnemonic` | Generates valid 12-word BIP39 mnemonic |
| `TestAddressFromMnemonic` | Derives deterministic 16-char hex address |
| `TestAddressFromMnemonic_Deterministic` | Same mnemonic always produces same address |
| `TestAddressFromMnemonic_EmptyString` | Handles empty input gracefully |
| `TestIsValidAddress` | Validates 16-char hex format |

**What's tested**:
- Mnemonic generation produces 12 words
- Consecutive generations produce unique mnemonics
- Address derivation is case-insensitive
- Address derivation trims whitespace
- Valid addresses: exactly 16 hex characters (a-f, 0-9)
- Invalid addresses: wrong length, non-hex chars, spaces

### SMTP Package (`internal/smtp/`)

| Test | Description |
|------|-------------|
| `TestBackend_NewSession` | Creates new session with store reference |
| `TestSession_Mail` | Handles MAIL FROM command |
| `TestSession_Rcpt` | Handles RCPT TO with valid hex address |
| `TestSession_Rcpt_Multiple` | Accumulates multiple recipients |
| `TestSession_Rcpt_InvalidAddress` | Rejects non-hex addresses with 550 error |
| `TestSession_Reset` | Clears session state |
| `TestSession_Logout` | Graceful logout |
| `TestSession_Data` | Parses and saves plain text email |
| `TestSession_Data_HTML` | Parses HTML email body |
| `TestSession_Data_MultiPart` | Prefers HTML over plain text in multipart |
| `TestSession_Data_InvalidEmail` | Rejects malformed email data |
| `TestSession_Data_EmptyBody` | Handles email with empty body |
| `TestSession_Data_MultipleRecipients` | Saves to each recipient's inbox |

**What's tested**:
- SMTP protocol flow (MAIL → RCPT → DATA)
- Address validation (rejects `admin@`, `test@`, etc.)
- Local part extraction from full email address
- Multi-recipient handling
- MIME parsing (plain text, HTML, multipart)
- Error handling for malformed emails

**Mock usage**: Uses `mockStore` to isolate SMTP logic from Redis.

### Store Package (`internal/store/`)

| Test | Description |
|------|-------------|
| `TestNewStore` | Creates store with correct Redis config |
| `TestNewStoreWithPassword` | Configures Redis password |
| `TestSaveEmail` | Saves email and sets TTL |
| `TestSaveEmail_MultipleEmails` | Stores multiple emails in list |
| `TestSaveEmail_GeneratesID` | Auto-generates UUID for emails |
| `TestGetEmails` | Retrieves emails in reverse chronological order |
| `TestGetEmails_EmptyInbox` | Returns empty array for nonexistent inbox |

**What's tested**:
- Redis connection configuration
- Email serialization (JSON)
- Redis list operations (LPUSH, LTRIM, LRANGE)
- TTL setting (24 hours)
- Email ID generation
- Empty inbox handling

**Note**: Store tests require Redis and are skipped with `-short` flag.

## Test Patterns

### Mock Store

The SMTP tests use a mock implementation of `EmailStore`:

```go
type mockStore struct {
    savedEmails []store.Email
    savedTo     []string
}

func (m *mockStore) SaveEmail(ctx context.Context, addressBox string, email store.Email) error {
    m.savedEmails = append(m.savedEmails, email)
    m.savedTo = append(m.savedTo, addressBox)
    return nil
}

func (m *mockStore) GetEmails(ctx context.Context, addressBox string) ([]store.Email, error) {
    return m.savedEmails, nil
}

func (m *mockStore) Ping(ctx context.Context) error {
    return nil
}
```

### Integration Test Skip

Store tests skip when Redis is unavailable:

```go
func TestSaveEmail(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // ... test setup ...
    
    if err := store.Ping(ctx); err != nil {
        t.Skipf("Redis not available: %v", err)
    }
    
    // ... actual test ...
}
```

### Test Cleanup

Integration tests clean up after themselves:

```go
defer func() {
    store.client.Del(ctx, key)
}()
```

## Writing New Tests

### Adding a Unit Test

```go
func TestMyNewFeature(t *testing.T) {
    // Arrange
    input := "test input"
    expected := "expected output"
    
    // Act
    result := MyFunction(input)
    
    // Assert
    if result != expected {
        t.Errorf("MyFunction(%q) = %q, want %q", input, result, expected)
    }
}
```

### Adding a Table-Driven Test

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"valid input", "foo", "bar"},
        {"empty input", "", ""},
        {"special chars", "a@b", "a_b"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := MyFunction(tt.input)
            if result != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

### Adding an Integration Test

```go
func TestRedisIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    ctx := context.Background()
    store := NewStore("localhost:6379", "")
    
    defer store.client.Close()
    
    if err := store.Ping(ctx); err != nil {
        t.Skipf("Redis not available: %v", err)
    }
    
    // Test code here...
    
    // Cleanup
    store.client.Del(ctx, "test-key")
}
```

## Continuous Integration

For CI pipelines, run tests with Redis service:

```yaml
# GitHub Actions example
services:
  redis:
    image: redis:alpine
    ports:
      - 6379:6379

steps:
  - uses: actions/checkout@v4
  - uses: actions/setup-go@v5
    with:
      go-version: '1.25'
  - run: go test ./...
    working-directory: backend
```

## Troubleshooting

### "Redis not available" Skip

Ensure Redis is running:
```bash
docker ps | grep redis
# or
redis-cli ping
```

### "Address already in use"

Previous test didn't clean up. Restart Redis:
```bash
docker restart redis-test
```

### Test Timeout

Increase timeout for slow environments:
```bash
go test -timeout 60s ./...
```
