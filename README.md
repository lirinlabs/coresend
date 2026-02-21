# CoreSend

Temporary email service with Ed25519 signature-based authentication.

## Overview

CoreSend provides disposable email addresses with cryptographic authentication. Each address is derived from an Ed25519 public key, ensuring only the key holder can access the inbox.

### Features

- **Cryptographic Authentication** - Ed25519 signature-based auth, no passwords
- **Disposable Addresses** - Self-destructing emails after 24 hours
- **Rate Limiting** - Prevents abuse with configurable limits
- **REST API** - Full REST API with OpenAPI/Swagger documentation
- **SMTP Server** - Standard SMTP for receiving emails

## Project Structure

```
coresend/
├── backend/          # Go HTTP API + SMTP server
│   ├── cmd/server/   # Entry point
│   ├── internal/     # API, SMTP, storage, validation
│   └── docs/         # Swagger documentation
└── app/              # React frontend (Vite + TypeScript)
```

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+ (or Bun)
- Redis 6.0+

### 1. Start Redis

```bash
redis-server
```

### 2. Start Backend

```bash
cd backend
make deps
make run
```

Backend runs on:

- HTTP API: `http://localhost:8080`
- SMTP: `localhost:1025`
- Swagger UI: `http://localhost:8080/docs/`

### 3. Start Frontend

```bash
cd app
bun install
bun run dev
```

Frontend runs on `http://localhost:5173`

## Documentation

- [Backend README](./backend/README.md) - Backend setup, configuration, and API
- [API Documentation](./docs/API.md) - Authentication flow and usage examples

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│  HTTP API   │────▶│    Redis    │
│  (Frontend) │     │  (Go/8080)  │     │  (Storage)  │
└─────────────┘     └─────────────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │ SMTP Server │
                    │  (Go/1025)  │
                    └─────────────┘
```

## Authentication

CoreSend uses Ed25519 signatures instead of passwords:

1. Generate Ed25519 keypair
2. Derive address: `sha256(publicKey)[:20]` (hex)
3. Sign each request: `timestamp:method:path`
4. Include headers: `X-Public-Key`, `X-Signature`, `X-Timestamp`

See [API Documentation](./docs/API.md) for detailed examples.

## API Endpoints

| Method   | Path                        | Description        |
| -------- | --------------------------- | ------------------ |
| `POST`   | `/api/register/{address}`   | Register address   |
| `GET`    | `/api/inbox/{address}`      | Get all emails     |
| `GET`    | `/api/inbox/{address}/{id}` | Get specific email |
| `DELETE` | `/api/inbox/{address}/{id}` | Delete email       |
| `DELETE` | `/api/inbox/{address}`      | Clear inbox        |
| `GET`    | `/api/health`               | Health check       |

## Development

### Backend

```bash
cd backend
make test           # Run tests
make swagger        # Generate swagger docs
make build          # Build binary
```

### Frontend

```bash
cd app
bun run lint        # Run linter
bun run build       # Build for production
```

## License

AGLP-3.0 License. See [LICENSE](./LICENSE) for details.
