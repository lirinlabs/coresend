# Security Policy

## Reporting a Vulnerability

**Do NOT open a public issue.**

Send details to:

- **lirinlabs@gmail.com**

Include what you can:

- Description of the issue
- Steps to reproduce (if applicable)
- Potential impact

## About This Project

This is an early-stage portfolio project exploring stateless temporary email using BIP39 deterministic logic.

**Current status**: Authentication flow is in progress - client-side cryptographic operations are being developed but not yet integrated with the backend.

**What works**: Inbound email receiving, 24-hour auto-expiry, rate limiting, mnemonic generation, and key generation.

**Not for production use** - this is a learning project.

## Scope

### In-Scope

- SMTP email parsing and sanitization
- Redis storage and TTL manipulation
- Input validation bypass
- Rate limiting evasion
- Container security

### Out-of-Scope

- All third-party dependencies (Redis, Go stdlib, React, Vite, TypeScript, etc.)
- Authentication limitations (see above - in progress)

I really welcome security feedback and contributions!
