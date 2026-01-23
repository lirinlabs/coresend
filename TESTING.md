# CoreSend Testing Guide

This document covers testing patterns, requirements, and best practices for the CoreSend project.

## Test Structure

### Directory Organization
```
backend/
├── internal/
│   ├── identity/
│   │   ├── generator.go
│   │   └── generator_test.go    # Unit tests
│   ├── smtp/
│   │   ├── backend.go
│   │   └── backend_test.go      # Unit tests with mocks
│   └── store/
│       ├── redis.go
│       └── redis_test.go        # Integration tests
└── test.sh                      # Test runner script
```

### Test Categories

#### Unit Tests
- Test individual functions in isolation
- No external dependencies
- Fast execution
- High coverage of business logic

#### Integration Tests
- Test with external services (Redis)
- Verify real interactions
- Slower execution
- Skip with `-short` flag

#### Mock Tests
- Use interfaces to isolate components
- Test protocol handling
- Verify error conditions
- No external dependencies

## Running Tests

### Quick Development Tests
```bash
# Unit tests only (no Redis required)
cd backend
go test -short ./...

# Verbose output
go test -short -v ./...

# Specific package
go test -short ./internal/identity/...
```

### Full Test Suite
```bash
# Start Redis
docker-compose up -d redis

# Run all tests
cd backend
go test ./...

# With race detection
go test -race ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out
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

### Test Script
```bash
# Use the provided test script
cd backend
./test.sh
```

## Test Patterns

### Table-Driven Tests
```go
func TestAddressFromMnemonic(t *testing.T) {
    tests := []struct {
        name     string
        mnemonic string
        expected string
    }{
        {
            name:     "valid mnemonic",
            mnemonic: "witch collapse practice feed shame open despair creek road again ice least",
            expected: "b4ebe3e2200cbc90",
        },
        {
            name:     "empty string",
            mnemonic: "",
            expected: "",
        },
        {
            name:     "whitespace",
            mnemonic: "  test  ",
            expected: "098f6bcd4621d373cade4e832627b4f6",
        },
        {
            name:     "case insensitive",
            mnemonic: "TEST MNEMONIC",
            expected: "098f6bcd4621d373cade4e832627b4f6",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := AddressFromMnemonic(tt.mnemonic)
            if result != tt.expected {
                t.Errorf("AddressFromMnemonic() = %q, want %q", result, tt.expected)
            }
        })
    }
}
```

### Mock Implementation
```go
type mockStore struct {
    savedEmails []store.Email
    savedTo     []string
    mu          sync.Mutex
}

func (m *mockStore) SaveEmail(ctx context.Context, addressBox string, email store.Email) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.savedEmails = append(m.savedEmails, email)
    m.savedTo = append(m.savedTo, addressBox)
    return nil
}

func (m *mockStore) GetEmails(ctx context.Context, addressBox string) ([]store.Email, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    return m.savedEmails, nil
}

func (m *mockStore) Ping(ctx context.Context) error {
    return nil
}
```

### Integration Test Setup
```go
func TestSaveEmail(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    ctx := context.Background()
    store := NewStore("localhost:6379", "")
    
    defer func() {
        store.client.Close()
        // Cleanup test data
        store.client.Del(ctx, "test-inbox")
    }()
    
    // Verify Redis is available
    if err := store.Ping(ctx); err != nil {
        t.Skipf("Redis not available: %v", err)
    }
    
    // Test implementation...
    email := store.Email{
        From:       "test@example.com",
        To:         []string{"test"},
        Subject:    "Test Subject",
        Body:       "Test Body",
        ReceivedAt: time.Now(),
    }
    
    err := store.SaveEmail(ctx, "test-inbox", email)
    if err != nil {
        t.Fatalf("SaveEmail() error = %v", err)
    }
    
    // Verify email was saved
    emails, err := store.GetEmails(ctx, "test-inbox")
    if err != nil {
        t.Fatalf("GetEmails() error = %v", err)
    }
    
    if len(emails) != 1 {
        t.Errorf("GetEmails() returned %d emails, want 1", len(emails))
    }
}
```

### Error Testing
```go
func TestSession_Rcpt_InvalidAddress(t *testing.T) {
    mockStore := &mockStore{}
    session := &Session{Store: mockStore}
    
    tests := []struct {
        name    string
        address string
        wantErr bool
    }{
        {"valid hex", "b4ebe3e2200cbc90", false},
        {"admin rejected", "admin", true},
        {"test rejected", "test", true},
        {"wrong length", "abc", true},
        {"non-hex", "ghijklmnopqrstuv", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            fullAddress := tt.address + "@example.com"
            err := session.Rcpt(fullAddress, nil)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Rcpt() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if tt.wantErr {
                var smtpErr *gosmtp.SMTPError
                if !errors.As(err, &smtpErr) {
                    t.Errorf("Rcpt() error type = %T, want *gosmtp.SMTPError", err)
                }
            }
        })
    }
}
```

## Test Coverage Requirements

### Coverage Targets by Package

#### Identity Package (`internal/identity/`)
- **Target**: 100% coverage
- **Reason**: Pure functions, critical security logic
- **Tests**: All input variations, edge cases

#### SMTP Package (`internal/smtp/`)
- **Target**: >90% coverage
- **Reason**: Protocol handling, complex state management
- **Tests**: Protocol flow, error conditions, multipart parsing

#### Store Package (`internal/store/`)
- **Target**: >85% coverage
- **Reason**: External dependency, data persistence
- **Tests**: Redis operations, error handling, TTL behavior

### Coverage Commands
```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage by function
go tool cover -func=coverage.out

# View coverage in browser
go tool cover -html=coverage.out

# Coverage threshold check
go test -coverprofile=coverage.out ./...
coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$coverage < 85" | bc -l) )); then
    echo "Coverage $coverage% is below threshold 85%"
    exit 1
fi
```

## Test Data Management

### Test Data Fixtures
```go
// Test data constants
const (
    TestMnemonic = "witch collapse practice feed shame open despair creek road again ice least"
    TestAddress  = "b4ebe3e2200cbc90"
    TestEmail    = "test@example.com"
)

// Test helper functions
func createTestEmail(from, to, subject, body string) store.Email {
    return store.Email{
        From:       from,
        To:         []string{to},
        Subject:    subject,
        Body:       body,
        ReceivedAt: time.Now(),
    }
}
```

### Test Cleanup
```go
func TestRedisOperations(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    ctx := context.Background()
    store := NewStore("localhost:6379", "")
    testKey := "test-inbox"
    
    // Ensure cleanup runs even if test fails
    defer func() {
        store.client.Del(ctx, testKey)
        store.client.Close()
    }()
    
    // Test implementation...
}
```

## Mock Testing Guidelines

### When to Use Mocks
- Testing SMTP protocol handling
- Verifying error conditions
- Isolating components from external dependencies
- Performance testing (avoid Redis overhead)

### Mock Best Practices
```go
// 1. Implement only needed methods
type mockStore struct {
    savedEmails []store.Email
    // Don't implement Ping if not needed
}

// 2. Use thread-safe operations
func (m *mockStore) SaveEmail(ctx context.Context, addressBox string, email store.Email) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    // Implementation...
}

// 3. Verify mock calls
func TestSMTPBackend(t *testing.T) {
    mockStore := &mockStore{}
    backend := &Backend{Store: mockStore}
    
    // Test code...
    
    // Verify mock was called correctly
    if len(mockStore.savedTo) != 1 {
        t.Errorf("Expected 1 call to SaveEmail, got %d", len(mockStore.savedTo))
    }
}
```

## Integration Testing

### Redis Integration Tests
```go
func TestRedisIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Redis integration test")
    }
    
    // Test configuration
    redisAddr := os.Getenv("TEST_REDIS_ADDR")
    if redisAddr == "" {
        redisAddr = "localhost:6379"
    }
    
    ctx := context.Background()
    store := NewStore(redisAddr, "")
    
    // Verify connection
    if err := store.Ping(ctx); err != nil {
        t.Skipf("Redis not available at %s: %v", redisAddr, err)
    }
    
    // Test implementation...
}
```

### Docker Compose Testing
```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d redis

# Run tests
cd backend
go test ./...

# Cleanup
docker-compose -f docker-compose.test.yml down
```

## Performance Testing

### Benchmark Tests
```go
func BenchmarkAddressFromMnemonic(b *testing.B) {
    mnemonic := "witch collapse practice feed shame open despair creek road again ice least"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        AddressFromMnemonic(mnemonic)
    }
}

func BenchmarkRedisSave(b *testing.B) {
    if testing.Short() {
        b.Skip("Skipping benchmark in short mode")
    }
    
    ctx := context.Background()
    store := NewStore("localhost:6379", "")
    email := createTestEmail("from@test.com", "to", "Subject", "Body")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        inbox := fmt.Sprintf("bench-%d", i)
        store.SaveEmail(ctx, inbox, email)
    }
}
```

### Race Condition Testing
```bash
# Run tests with race detection
go test -race ./...

# Run specific package with race detection
go test -race ./internal/store/...
```

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      redis:
        image: redis:alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      
      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-
      
      - name: Download dependencies
        run: go mod download
        working-directory: backend
      
      - name: Run tests
        run: go test -race -coverprofile=coverage.out ./...
        working-directory: backend
      
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          file: ./backend/coverage.out
          flags: unittests
          name: codecov-umbrella
```

### Pre-commit Hooks
```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: go-test
        name: go test
        entry: go test -short ./...
        language: system
        files: '\.go$'
        pass_filenames: false
      
      - id: go-vet
        name: go vet
        entry: go vet ./...
        language: system
        files: '\.go$'
        pass_filenames: false
      
      - id: go-fmt
        name: go fmt
        entry: gofmt -d
        language: system
        files: '\.go$'
```

## Test Documentation

### Test Comments
```go
// TestAddressFromMnemonic verifies deterministic address generation
// from BIP39 mnemonic phrases using SHA-256 hashing.
func TestAddressFromMnemonic(t *testing.T) {
    // Test case: Valid mnemonic produces expected address
    t.Run("valid mnemonic", func(t *testing.T) {
        // Arrange
        mnemonic := "witch collapse practice feed shame open despair creek road again ice least"
        expected := "b4ebe3e2200cbc90"
        
        // Act
        result := AddressFromMnemonic(mnemonic)
        
        // Assert
        if result != expected {
            t.Errorf("AddressFromMnemonic(%q) = %q, want %q", mnemonic, result, expected)
        }
    })
}
```

### README Testing Section
```markdown
## Running Tests

### Quick Tests
```bash
cd backend && go test -short ./...
```

### Full Tests
```bash
docker-compose up -d redis
cd backend && go test ./...
```

### Coverage
```bash
cd backend && go test -coverprofile=coverage.out ./...
```
```

## Troubleshooting

### Common Test Issues

#### "Redis not available" Skip
```bash
# Check Redis is running
docker-compose ps redis

# Start Redis
docker-compose up -d redis

# Test connection
docker-compose exec redis redis-cli ping
```

#### "Address already in use"
```bash
# Find process using port
lsof -i :6379

# Kill process
kill -9 <PID>

# Or restart Redis container
docker-compose restart redis
```

#### Test Timeouts
```bash
# Increase timeout
go test -timeout 30s ./...

# Run tests in verbose mode to see progress
go test -v -timeout 30s ./...
```

#### Coverage Issues
```bash
# Check which lines aren't covered
go tool cover -html=coverage.out

# Run specific package for better coverage
go test -coverprofile=coverage.out ./internal/package/...
```

---

This testing guide ensures comprehensive, maintainable tests for the CoreSend project. For coding standards, see [CODING_STANDARDS.md](CODING_STANDARDS.md). For development workflow, see [DEVELOPMENT.md](DEVELOPMENT.md).