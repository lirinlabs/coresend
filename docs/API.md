# CoreSend API Documentation

Complete API reference with authentication examples.

## Base URL

```
http://localhost:8080
```

## Authentication

All protected endpoints require Ed25519 signature-based authentication.

### Required Headers

| Header         | Description                            |
| -------------- | -------------------------------------- |
| `X-Public-Key` | Ed25519 public key (64 hex characters) |
| `X-Signature`  | Ed25519 signature (128 hex characters) |
| `X-Timestamp`  | Unix timestamp in seconds              |

### Signature Generation

The message to sign is constructed as:

```
{timestamp}:{method}:{path}
```

Example:

```
1700000000:GET:/api/inbox/abc123def456...
```

### Complete Example (cURL)

```bash
# Generate keys (using openssl)
openssl genpkey -algorithm ED25519 -out private.pem
openssl pkey -in private.pem -pubout -out public.pem

# Extract raw public key (last 32 bytes of DER)
PUBKEY_HEX=$(openssl pkey -in public.pem -pubin -outform DER | tail -c 32 | xxd -p -c 32)

# Derive address (first 20 bytes of SHA256)
ADDRESS=$(echo -n "$PUBKEY_HEX" | xxd -r -p | sha256sum | cut -c1-40)

# Create timestamp
TIMESTAMP=$(date +%s)

# Create message to sign
MESSAGE="${TIMESTAMP}:GET:/api/inbox/${ADDRESS}"

# Sign with private key
# Note: This requires pkeyutl or a custom script
# For simplicity, use the Go or Node.js examples for signing

# Make request (signature must be generated programmatically)
curl -X GET "http://localhost:8080/api/inbox/${ADDRESS}" \
  -H "X-Public-Key: ${PUBKEY_HEX}" \
  -H "X-Signature: ${SIGNATURE}" \
  -H "X-Timestamp: ${TIMESTAMP}"
```

## Endpoints

### Health Check

```http
GET /api/health
```

No authentication required.

## Rate Limiting

| Endpoint Type     | Rate Limit       |
| ----------------- | ---------------- |
| Inbox read (GET)  | 60/minute per IP |
| Delete operations | 30/minute per IP |
| Health check      | No limit         |

Rate limits are enforced via Redis sliding window algorithm.

## Email Lifecycle

1. **Registration** - Register address before receiving emails
2. **Reception** - Emails received via SMTP are stored automatically
3. **Storage** - Emails expire after 24 hours (TTL)
4. **Deletion** - Manually delete or wait for expiration

## Swagger UI

Interactive API documentation available at:

```
http://localhost:8080/docs/
```

## OpenAPI Specification

Available at:

```
http://localhost:8080/docs/swagger.json
http://localhost:8080/docs/swagger.yaml
```
