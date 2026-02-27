# CoreSend Backend

Temporary email service with Ed25519 signature-based authentication.

## Architecture

- **HTTP API** - RESTful API for email management (port 8080)
- **SMTP Server** - Receives incoming emails (port 1025)
- **Redis** - Email storage and rate limiting

## Requirements

- Go 1.21+
- Redis 6.0+

## Quick Start

```bash
# Install dependencies
make deps

# Start Redis (if not running)
redis-server

# Run the server
make run
```

Server starts on:

- HTTP API: `http://localhost:8080`
- SMTP: `localhost:1025`
- Swagger UI: `http://localhost:8080/docs/`

## Environment Variables

| Variable           | Default          | Description                         |
| ------------------ | ---------------- | ----------------------------------- |
| `REDIS_ADDR`       | `localhost:6379` | Redis server address                |
| `REDIS_PASSWORD`   | (empty)          | Redis password                      |
| `DOMAIN_NAME`      | `localhost`      | Domain for email addresses          |
| `SMTP_LISTEN_ADDR` | `:1025`          | SMTP server listen address          |
| `HTTP_LISTEN_ADDR` | `:8080`          | HTTP API listen address             |
| `SMTP_CERT_PATH`   | (empty)          | TLS certificate path (for STARTTLS) |
| `SMTP_KEY_PATH`    | (empty)          | TLS private key path                |

## Authentication

API uses Ed25519 signature-based authentication. All protected endpoints require:

| Header         | Description                                |
| -------------- | ------------------------------------------ |
| `X-Public-Key` | Ed25519 public key (hex-encoded, 64 chars) |
| `X-Signature`  | Ed25519 signature (hex-encoded)            |
| `X-Timestamp`  | Unix timestamp (seconds)                   |

### How It Works

1. Derive your address from public key: `sha256(publicKey)[:20]` (hex-encoded)
2. Sign the message: `timestamp + ":" + method + ":" + path`
3. Include public key, signature, and timestamp in headers

### Example

```bash
# Generate Ed25519 keypair (example)
openssl genpkey -algorithm ED25519 -out private.pem
openssl pkey -in private.pem -pubout -out public.pem

# Your address is derived from public key
# Address = first 20 bytes of SHA256(public_key_hex)

# For each request, sign: timestamp:method:path
# Example message to sign: "1700000000:GET:/api/inbox/abc123..."
```

## API Endpoints

| Method   | Path                             | Auth | Rate Limit | Description                |
| -------- | -------------------------------- | ---- | ---------- | -------------------------- |
| `POST`   | `/api/register/{address}`        | Yes  | -          | Register a new address     |
| `GET`    | `/api/inbox/{address}`           | Yes  | 60/min     | Get all emails for address |
| `GET`    | `/api/inbox/{address}/{emailId}` | Yes  | 60/min     | Get specific email         |
| `DELETE` | `/api/inbox/{address}/{emailId}` | Yes  | 30/min     | Delete specific email      |
| `DELETE` | `/api/inbox/{address}`           | Yes  | 30/min     | Clear entire inbox         |
| `GET`    | `/api/health`                    | No   | -          | Health check               |

## Rate Limiting

- Inbox operations: 60 requests/minute per IP
- Delete operations: 30 requests/minute per IP
- Rate limits are enforced via Redis with sliding window

## Development

```bash
# Generate swagger docs
make swagger

# Build binary
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run linter
make lint

# Full rebuild
make rebuild
```

## Project Structure

```
backend/
├── cmd/server/main.go    # Application entry point
├── internal/
│   ├── api/              # HTTP API handlers, middleware, router
│   ├── smtp/             # SMTP server backend
│   ├── store/            # Redis storage layer
│   └── validator/        # Input validation
├── docs/                 # Swagger documentation
├── Makefile              # Build commands
└── Dockerfile            # Container image
```

## Email Storage

Emails are stored in Redis with:

- **TTL**: 24 hours (configurable in store)
- **Structure**: ZSet (ordered by timestamp) + Hash (email data)
- **Address format**: 40 hex characters derived from Ed25519 public key

## TLS/STARTTLS

To enable TLS for SMTP:

```bash
export SMTP_CERT_PATH=/path/to/cert.pem
export SMTP_KEY_PATH=/path/to/key.pem
make run
```

## License

AGPL-3.0 License. See [LICENSE](LICENSE) for details.
