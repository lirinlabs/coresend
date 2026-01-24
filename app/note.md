Frontend needs to:
1. Generate Ed25519 keypair from mnemonic using same HMAC-SHA256 domain separation
2. Sign messages with address + "|" + timestamp
3. Include all 4 auth headers in API requests
4. Use Web Crypto API or compatible library