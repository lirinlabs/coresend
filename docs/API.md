# CoreSend API Documentation

## Status

> **Note**: The HTTP API is not yet implemented. This document describes the planned API design for future implementation.

## Overview

The CoreSend API provides programmatic access to temporary email inboxes. It follows REST conventions and returns JSON responses.

### Base URLs

| Environment | URL |
|-------------|-----|
| Production | `https://coresend.io/api` |
| Development | `http://localhost/api` or `http://localhost:8080/api` |

### Authentication

The API does not use traditional authentication. Instead, the mnemonic phrase serves as proof of ownership:

- **Public endpoints**: Accessible without authentication
- **Inbox access**: Requires knowledge of the mnemonic phrase (which derives the address)

Since anyone who knows the 16-character hex address can access the inbox, the security model relies on the address being unguessable (derived from a 12-word mnemonic = 128 bits of entropy).

## Planned Endpoints

### Identity

#### Generate New Mnemonic

Creates a new 12-word BIP39 mnemonic phrase and derives the corresponding email address.

```
POST /api/identity/generate
```

**Request**: No body required

**Response**:
```json
{
  "mnemonic": "witch collapse practice feed shame open despair creek road again ice least",
  "address": "b4ebe3e2200cbc90",
  "email": "b4ebe3e2200cbc90@coresend.io"
}
```

**Status Codes**:
- `200 OK`: Success
- `500 Internal Server Error`: Generation failed

---

#### Derive Address from Mnemonic

Derives the email address from an existing mnemonic phrase. Use this for "login" functionality.

```
POST /api/identity/derive
```

**Request**:
```json
{
  "mnemonic": "witch collapse practice feed shame open despair creek road again ice least"
}
```

**Response**:
```json
{
  "address": "b4ebe3e2200cbc90",
  "email": "b4ebe3e2200cbc90@coresend.io",
  "valid": true
}
```

**Status Codes**:
- `200 OK`: Success (even if mnemonic is invalid, returns `valid: false`)
- `400 Bad Request`: Malformed request body

**Notes**:
- The mnemonic is normalized (lowercased, trimmed) before hashing
- Invalid BIP39 mnemonics still produce an address but return `valid: false`

---

#### Validate Address

Checks if a string is a valid CoreSend address format.

```
GET /api/identity/validate/{address}
```

**Response**:
```json
{
  "address": "b4ebe3e2200cbc90",
  "valid": true,
  "reason": ""
}
```

Invalid address response:
```json
{
  "address": "invalid",
  "valid": false,
  "reason": "address must be exactly 16 hexadecimal characters"
}
```

---

### Inbox

#### Get All Emails

Retrieves all emails for a given address.

```
GET /api/inbox/{address}
```

**Parameters**:
- `address` (path): 16-character hex address

**Response**:
```json
{
  "address": "b4ebe3e2200cbc90",
  "email": "b4ebe3e2200cbc90@coresend.io",
  "count": 2,
  "emails": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "from": "sender@example.com",
      "to": ["b4ebe3e2200cbc90"],
      "subject": "Hello World",
      "body": "<html>...</html>",
      "received_at": "2025-01-20T12:00:00Z"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "from": "another@example.com",
      "to": ["b4ebe3e2200cbc90"],
      "subject": "Test Email",
      "body": "Plain text body",
      "received_at": "2025-01-20T11:30:00Z"
    }
  ]
}
```

**Status Codes**:
- `200 OK`: Success (empty array if no emails)
- `400 Bad Request`: Invalid address format

**Notes**:
- Emails are returned in reverse chronological order (newest first)
- Maximum 100 emails returned
- Empty inbox returns `emails: []`, not an error

---

#### Get Single Email

Retrieves a specific email by ID.

```
GET /api/inbox/{address}/{emailId}
```

**Parameters**:
- `address` (path): 16-character hex address
- `emailId` (path): UUID of the email

**Response**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "from": "sender@example.com",
  "to": ["b4ebe3e2200cbc90"],
  "subject": "Hello World",
  "body": "<html>...</html>",
  "received_at": "2025-01-20T12:00:00Z"
}
```

**Status Codes**:
- `200 OK`: Success
- `400 Bad Request`: Invalid address format
- `404 Not Found`: Email not found

---

#### Delete Email

Deletes a specific email from the inbox.

```
DELETE /api/inbox/{address}/{emailId}
```

**Parameters**:
- `address` (path): 16-character hex address
- `emailId` (path): UUID of the email

**Response**:
```json
{
  "deleted": true,
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Status Codes**:
- `200 OK`: Successfully deleted
- `400 Bad Request`: Invalid address format
- `404 Not Found`: Email not found

---

#### Delete All Emails

Clears all emails from an inbox.

```
DELETE /api/inbox/{address}
```

**Response**:
```json
{
  "deleted": true,
  "count": 5
}
```

**Status Codes**:
- `200 OK`: Successfully deleted (even if inbox was empty)
- `400 Bad Request`: Invalid address format

---

### Health

#### Health Check

Simple health check endpoint.

```
GET /api/health
```

**Response**:
```json
{
  "status": "healthy",
  "services": {
    "redis": "connected",
    "smtp": "running"
  }
}
```

**Status Codes**:
- `200 OK`: All services healthy
- `503 Service Unavailable`: One or more services unhealthy

---

## Error Responses

All errors follow a consistent format:

```json
{
  "error": {
    "code": "INVALID_ADDRESS",
    "message": "Address must be exactly 16 hexadecimal characters",
    "details": {
      "provided": "invalid",
      "expected_length": 16
    }
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_ADDRESS` | 400 | Address format is invalid |
| `INVALID_MNEMONIC` | 400 | Mnemonic phrase is malformed |
| `NOT_FOUND` | 404 | Resource not found |
| `INTERNAL_ERROR` | 500 | Server error |
| `SERVICE_UNAVAILABLE` | 503 | Dependency unavailable |

## Rate Limiting

> **Note**: Rate limiting is not yet implemented.

Planned limits:
- `POST /api/identity/generate`: 10 requests per minute per IP
- `GET /api/inbox/*`: 60 requests per minute per IP
- `DELETE /api/inbox/*`: 30 requests per minute per IP

Rate limit headers:
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1705752000
```

## CORS

The API will support CORS for browser-based access:

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type
```

## Examples

### cURL

```bash
# Generate new identity
curl -X POST https://coresend.io/api/identity/generate

# Derive address from mnemonic
curl -X POST https://coresend.io/api/identity/derive \
  -H "Content-Type: application/json" \
  -d '{"mnemonic": "witch collapse practice feed shame open despair creek road again ice least"}'

# Get inbox
curl https://coresend.io/api/inbox/b4ebe3e2200cbc90

# Delete email
curl -X DELETE https://coresend.io/api/inbox/b4ebe3e2200cbc90/550e8400-e29b-41d4-a716-446655440000
```

### JavaScript (fetch)

```javascript
// Generate new identity
const response = await fetch('https://coresend.io/api/identity/generate', {
  method: 'POST'
});
const { mnemonic, address, email } = await response.json();

// Get inbox
const inbox = await fetch(`https://coresend.io/api/inbox/${address}`);
const { emails } = await inbox.json();
```

### Python (requests)

```python
import requests

# Generate new identity
response = requests.post('https://coresend.io/api/identity/generate')
data = response.json()
print(f"Your email: {data['email']}")

# Get inbox
inbox = requests.get(f"https://coresend.io/api/inbox/{data['address']}")
for email in inbox.json()['emails']:
    print(f"From: {email['from']} - {email['subject']}")
```

## Implementation Status

| Endpoint | Status |
|----------|--------|
| `POST /api/identity/generate` | Planned |
| `POST /api/identity/derive` | Planned |
| `GET /api/identity/validate/{address}` | Planned |
| `GET /api/inbox/{address}` | Planned |
| `GET /api/inbox/{address}/{emailId}` | Planned |
| `DELETE /api/inbox/{address}/{emailId}` | Planned |
| `DELETE /api/inbox/{address}` | Planned |
| `GET /api/health` | Planned |

## Changelog

### Unreleased
- Initial API design documented
