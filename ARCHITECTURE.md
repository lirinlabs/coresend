# CoreSend Architecture

CoreSend's architecture is built around a simple but powerful idea: **deterministic identity through mnemonic phrases**. No database, no user accounts, just cryptography.

## Core concept: Mnemonic → Address

The entire system revolves around this conversion:

```
12 BIP39 words → SHA256 → First 8 bytes → Hex email address
```

### Why this works

1. **Deterministic**: Same words = same address, always
2. **Stateless**: No need to store user data
3. **Private**: Words never transmitted, only locally
4. **Memorable**: Humans can remember 12 words better than random hex

### Example conversion

```go
mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
hash := sha256.Sum256([]byte(mnemonic))
address := hex.EncodeToString(hash[:8]) // "7c7e7c7e7c7e7c7e"
```

## System design

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client        │    │   SMTP Server   │    │     Redis       │
│                 │    │                 │    │                 │
│ 12 words →      │───▶│  Parse email    │───▶│  Store 24h      │
│ address calc    │    │  Validate dest  │    │  TTL cleanup    │
│                 │    │  Forward        │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   HTTP API      │
                       │                 │
                       │  GET /messages  │
                       │  POST /address  │
                       └─────────────────┘
```

## Key components

### 1. Identity Generator (`internal/identity/`)

```go
type Generator struct {
    // No state needed, purely functional
}

func (g *Generator) Generate(mnemonic string) (string, error) {
    // Validate BIP39 words
    // Hash and truncate
    // Return hex address
}
```

**Key decisions:**
- Pure functions, no state
- Fast SHA256, not bcrypt (we want speed, not security)
- 8 bytes = 16 hex chars = reasonable collision odds

### 2. SMTP Backend (`internal/smtp/`)

```go
type Backend struct {
    store Store
}

func (b *Backend) Send(from string, to []string, r io.Reader) error {
    // Parse email with standard library
    // Validate destination is our domain
    // Extract address from local-part
    // Store in Redis with 24h TTL
}
```

**Key decisions:**
- Use Go's `net/smtp` for standards compliance
- Parse with standard `mail` package
- Store raw email, preserve headers
- Redis TTL handles automatic cleanup

### 3. Redis Store (`internal/store/`)

```go
type RedisStore struct {
    client *redis.Client
}

func (s *RedisStore) Store(address string, email []byte) error {
    // Redis key: "email:{address}:{timestamp}"
    // 24 hour TTL
}

func (s *RedisStore) Get(address string) ([]Email, error) {
    // SCAN for keys matching pattern
    // Return sorted by timestamp
}
```

**Key decisions:**
- No complex data structures, simple key-value
- TTL handles cleanup automatically
- SCAN over KEYS for production safety
- JSON storage for API simplicity

## Security considerations

### What we don't store
- Mnemonics (never transmitted)
- User data of any kind
- IP addresses or logs

### What we protect against
- **Enumeration**: Brute forcing addresses is computationally expensive
- **Persistence**: 24-hour TTL limits data exposure
- **Privacy**: No tracking, no analytics

### Attack surface
```
SMTP (port 25)  → Standard email protocols, rate limited
HTTP (port 80)  → Simple API, minimal endpoints
Redis (internal) → Network isolated, auth required
```

## Deployment architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Caddy       │    │   CoreSend      │    │     Redis       │
│                 │    │                 │    │                 │
│  TLS termination│───▶│  HTTP API       │───▶│  Email storage  │
│  Reverse proxy  │    │  SMTP server    │    │  TTL management │
│  Static files   │    │  Identity logic │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Container design
- **App container**: Go binary, minimal base image
- **Redis container**: Official redis-alpine
- **Caddy container**: Automatic HTTPS, static serving
- **All containers**: Read-only filesystems, dropped capabilities

## Scaling considerations

### Horizontal scaling
- Stateless app can run multiple instances
- Redis clustering for large scale
- Load balancer needed for SMTP (rare)

### Bottlenecks
- **Redis I/O**: Memory-bound, predictable patterns
- **SMTP connections**: Connection pooling helps
- **API reads**: Redis can handle millions of GETs

### Monitoring points
- Redis memory usage (24h window)
- SMTP connection rates  
- API response times
- TTL cleanup efficiency

This architecture prioritizes simplicity and privacy over complex features. The mnemonic-based identity system eliminates the need for most traditional email service infrastructure.