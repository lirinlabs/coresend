# CoreSend

A stateless temporary email service with mnemonic-based identity. Like a crypto wallet, but for disposable email addresses.

## What is CoreSend?

CoreSend is a self-hosted temporary email (temp-mail) service that uses BIP39 mnemonic phrases for identity. Instead of creating accounts with usernames and passwords, users generate or enter a 12-word secret phrase that deterministically creates their email address.

**Key concept**: Your mnemonic phrase IS your identity. No registration, no passwords, no account recovery. Just 12 words that you can memorize or store securely.

## How It Works

```
1. User generates or enters a 12-word mnemonic phrase
   Example: "witch collapse practice feed shame open despair creek road again ice least"

2. The phrase is hashed (SHA-256) to create a 16-character hex address
   Example: "b4ebe3e2200cbc90"

3. This becomes the email address
   Example: b4ebe3e2200cbc90@coresend.io

4. Incoming emails are stored in Redis with 24-hour TTL

5. User can "login" anytime with the same phrase to view their inbox
   (Same phrase = same address = same inbox)
```

### Why Mnemonic-Based Identity?

- **Stateless**: No user database, no password hashing, no session management
- **Deterministic**: Same phrase always produces the same address
- **Portable**: Users can access their inbox from anywhere with just their phrase
- **Private**: No email verification, no personal data stored
- **Memorable**: 12 words are easier to remember than random strings

## Architecture

```
                         Internet
                            |
          +-----------------+-----------------+
          |                 |                 |
     Port 25           Port 80            Port 443
      (SMTP)           (HTTP)             (HTTPS)
          |                 |                 |
          v                 +--------+--------+
    +----------+                     |
    | Backend  |                     v
    |  (Go)    |              +------------+
    |          |              |   Caddy    |
    | SMTP:1025|              |            |
    | API:8080 |<-------------| Auto HTTPS |
    +----+-----+   /api/*     | Static SPA |
         |        proxy       +------------+
         |                          |
         v                          v
    +----------+            +---------------+
    |  Redis   |            | Frontend Dist |
    |          |            |  (Static)     |
    | Email    |            +---------------+
    | Storage  |
    +----------+

    Shared: Caddy SSL certs --> Backend (for SMTP STARTTLS)
```

### Request Flow

**Incoming Email (SMTP)**:
```
Sender --> Port 25 --> Backend SMTP --> Validate hex address --> Redis (inbox:{address})
```

**User Access (HTTP)**:
```
Browser --> Caddy (443) --> /api/inbox/{address} --> Backend API --> Redis --> JSON response
```

## Features

### Implemented
- [x] BIP39 mnemonic generation (12 words)
- [x] Deterministic address derivation (mnemonic -> SHA256 -> 16-char hex)
- [x] SMTP server (inbound only)
- [x] Address validation (rejects non-hex addresses like admin@, test@)
- [x] Multi-recipient support
- [x] HTML and plain-text email parsing
- [x] Redis storage with 24-hour TTL
- [x] Maximum 100 emails per inbox
- [x] Automatic HTTPS via Caddy
- [x] STARTTLS support for SMTP
- [x] Graceful shutdown
- [x] Health checks
- [x] Docker Compose deployment

### Planned
- [ ] HTTP API for inbox access
- [ ] Frontend web application
- [ ] Email deletion
- [ ] Attachment support
- [ ] Webhook notifications

## Project Structure

```
coresend/
├── backend/                    # Go backend application
│   ├── cmd/
│   │   └── server/
│   │       └── main.go         # Application entry point
│   ├── internal/
│   │   ├── identity/
│   │   │   ├── generator.go    # Mnemonic & address generation
│   │   │   └── generator_test.go
│   │   ├── smtp/
│   │   │   ├── backend.go      # SMTP session handling
│   │   │   └── backend_test.go
│   │   └── store/
│   │       ├── redis.go        # Redis operations
│   │       └── redis_test.go
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
│
├── frontend/                   # Frontend application (planned)
│   └── dist/                   # Built static files
│
├── caddy/
│   ├── Caddyfile              # Production config
│   └── Caddyfile.dev          # Development config
│
├── docs/
│   └── API.md                 # API documentation
│
├── docker-compose.yml         # Production compose
├── docker-compose.override.yml # Development overrides
├── .env.example               # Environment template
├── DOCKER_SETUP.md            # Docker deployment guide
├── LICENSE                    # MIT License
└── README.md                  # This file
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Domain name with DNS pointing to your server (for production)
- Ports 25, 80, 443 available

### Development (Local)

```bash
# Clone the repository
git clone https://github.com/fn-jakubkarp/coresend.git
cd coresend

# Start services (uses docker-compose.override.yml automatically)
docker-compose up -d

# View logs
docker-compose logs -f

# Access at http://localhost
```

### Production

```bash
# Clone and configure
git clone https://github.com/fn-jakubkarp/coresend.git
cd coresend
cp .env.example .env

# Edit .env with your domain
echo "DOMAIN_NAME=coresend.io" > .env

# Start production services (ignore override file)
docker-compose -f docker-compose.yml up -d

# Caddy will automatically obtain SSL certificates
# Check logs to confirm
docker-compose logs caddy
```

See [DOCKER_SETUP.md](DOCKER_SETUP.md) for detailed deployment instructions.

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DOMAIN_NAME` | Yes (prod) | `localhost` | Your domain name |
| `REDIS_PASSWORD` | No | *(empty)* | Redis authentication password |

### Email Address Format

Valid addresses must be exactly 16 hexadecimal characters:
- `b4ebe3e2200cbc90@coresend.io` - Valid
- `admin@coresend.io` - Rejected (not hex)
- `test@coresend.io` - Rejected (not 16 chars)

This prevents spam to common addresses and ensures only mnemonic-derived addresses receive mail.

### Storage Limits

- **TTL**: 24 hours (inbox expires 24h after last email received)
- **Max emails**: 100 per inbox (oldest trimmed when exceeded)
- **Max size**: 1MB per email

## Tech Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Backend | Go 1.25 | SMTP server, API |
| SMTP Library | go-smtp | SMTP protocol handling |
| Email Parsing | go-message | MIME parsing |
| Mnemonic | go-bip39 | BIP39 phrase generation |
| Database | Redis 7 | Email storage |
| Web Server | Caddy | Reverse proxy, auto HTTPS |
| Container | Docker | Deployment |

## Documentation

- [Docker Setup Guide](DOCKER_SETUP.md) - Deployment and configuration
- [API Documentation](docs/API.md) - HTTP API reference (planned)
- [Testing Guide](backend/TESTING.md) - Running tests

## Contributing

Contributions are welcome! Here's how to get started:

### Development Setup

1. Fork and clone the repository
2. Start the development environment:
   ```bash
   docker-compose up -d
   ```
3. Make your changes
4. Run tests:
   ```bash
   cd backend
   go test ./...
   ```
5. Submit a pull request

### Guidelines

- Write tests for new functionality
- Follow existing code style
- Update documentation as needed
- Keep commits focused and descriptive

### Running Tests

```bash
# Unit tests (no Redis required)
cd backend
go test -short ./...

# Full tests (requires Redis)
docker-compose up -d redis
go test ./...
```

## Testing the SMTP Server

You can test the SMTP server using various tools:

### Using swaks (Swiss Army Knife for SMTP)

```bash
# Install swaks
apt-get install swaks  # Debian/Ubuntu
brew install swaks     # macOS

# Send test email
swaks --to b4ebe3e2200cbc90@coresend.io \
      --from test@example.com \
      --server localhost:25 \
      --header "Subject: Test Email" \
      --body "Hello from swaks!"
```

### Using telnet

```bash
telnet localhost 25
HELO test.com
MAIL FROM:<test@example.com>
RCPT TO:<b4ebe3e2200cbc90@coresend.io>
DATA
Subject: Test

Hello World
.
QUIT
```

### Using Python

```python
import smtplib
from email.mime.text import MIMEText

msg = MIMEText("Hello from Python!")
msg['Subject'] = 'Test Email'
msg['From'] = 'test@example.com'
msg['To'] = 'b4ebe3e2200cbc90@coresend.io'

with smtplib.SMTP('localhost', 25) as server:
    server.send_message(msg)
```

## Security Considerations

### What CoreSend Does NOT Provide

- **Email encryption**: Emails are stored in plain text in Redis
- **Sender verification**: No SPF/DKIM/DMARC validation
- **Spam filtering**: All valid-addressed emails are accepted
- **Long-term storage**: Emails are deleted after 24 hours

### Production Recommendations

1. **Firewall**: Only expose necessary ports
   ```bash
   sudo ufw allow 22/tcp   # SSH (host OS, not Docker)
   sudo ufw allow 25/tcp   # SMTP
   sudo ufw allow 80/tcp   # HTTP (redirects to HTTPS)
   sudo ufw allow 443/tcp  # HTTPS
   sudo ufw enable
   ```

2. **Redis**: Set a password in production
   ```bash
   REDIS_PASSWORD=your-secure-password
   ```

3. **Monitoring**: Check logs regularly
   ```bash
   docker-compose logs -f
   ```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [go-smtp](https://github.com/emersion/go-smtp) - SMTP server library
- [go-bip39](https://github.com/tyler-smith/go-bip39) - BIP39 mnemonic implementation
- [Caddy](https://caddyserver.com/) - Automatic HTTPS web server
