# CoreSend Security Guide

This document covers security best practices, configuration guidelines, and security considerations for the CoreSend project.

## Security Overview

CoreSend is designed with security as a primary concern, implementing multiple layers of protection while maintaining simplicity and statelessness.

### Key Security Features
- **Mnemonic-based identity**: No passwords or user accounts to compromise
- **Address validation**: Only accepts 16-character hexadecimal addresses
- **Stateless design**: No user database or session management
- **Container security**: Non-root execution, minimal attack surface
- **STARTTLS support**: SMTP encryption using Caddy's certificates

## Container Security

### Non-Root Execution
```dockerfile
# Backend Dockerfile
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

USER appuser
```

### Minimal Base Images
```dockerfile
# Use Alpine Linux for reduced attack surface
FROM golang:1.25-alpine AS builder
FROM alpine:latest

# Install only necessary packages
RUN apk --no-cache add ca-certificates tzdata netcat-openbsd
```

### Read-Only Volumes
```yaml
# docker-compose.yml
volumes:
  # Share Caddy's certificates read-only
  - caddy_certs:/certs:ro
```

### Resource Limits
```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.5'
        reservations:
          memory: 128M
          cpus: '0.25'
```

## Network Security

### Firewall Configuration
```bash
# Production firewall rules
sudo ufw allow 22/tcp   # SSH (host OS, not Docker)
sudo ufw allow 25/tcp   # SMTP
sudo ufw allow 80/tcp   # HTTP (redirects to HTTPS)
sudo ufw allow 443/tcp  # HTTPS
sudo ufw enable

# Verify rules
sudo ufw status verbose
```

### Docker Network Isolation
```yaml
# docker-compose.yml
networks:
  coresend-network:
    driver: bridge
    internal: false  # Allows internet access for Let's Encrypt

services:
  backend:
    networks:
      - coresend-network
    # No external ports exposed except SMTP (25:1025)
```

### Port Security
```yaml
services:
  backend:
    ports:
      - "25:1025"  # SMTP only, no admin ports exposed
  
  caddy:
    ports:
      - "80:80"    # HTTP (redirects to HTTPS)
      - "443:443"  # HTTPS
  
  redis:
    # No ports exposed - internal only
    networks:
      - coresend-network
```

## Application Security

### Address Validation
```go
// Only accept 16-character hexadecimal addresses
var validAddressRegex = regexp.MustCompile(`^[a-f0-9]{16}$`)

func IsValidAddress(addr string) bool {
    return validAddressRegex.MatchString(strings.ToLower(addr))
}

// Reject common addresses in SMTP handler
func (s *Session) Rcpt(to string, opts *gosmtp.RcptOptions) error {
    localPart := extractLocalPart(to)
    if !identity.IsValidAddress(localPart) {
        return &gosmtp.SMTPError{
            Code:         550,
            EnhancedCode: gosmtp.EnhancedCode{5, 1, 1},
            Message:      "Mailbox does not exist",
        }
    }
    return nil
}
```

### Input Sanitization
```go
// Clean and validate mnemonic input
func AddressFromMnemonic(mnemonic string) string {
    // Trim whitespace and convert to lowercase
    mnemonic = strings.TrimSpace(strings.ToLower(mnemonic))
    
    // Hash and truncate
    hash := sha256.Sum256([]byte(mnemonic))
    return hex.EncodeToString(hash[:])[:AddressLength]
}
```

### Error Handling
```go
// Log errors without exposing sensitive information
func (s *Session) Data(r io.Reader) error {
    mr, err := mail.CreateReader(r)
    if err != nil {
        log.Printf("Error parsing email: %v", err)  // Don't log email content
        return err
    }
    
    // Process email...
}

// Return generic error messages to clients
return &gosmtp.SMTPError{
    Code:         550,
    EnhancedCode: gosmtp.EnhancedCode{5, 1, 1},
    Message:      "Mailbox does not exist",  // Don't reveal why
}
```

### STARTTLS Configuration
```go
// Use Caddy's certificates for SMTP STARTTLS
if certPath != "" && keyPath != "" {
    cert, err := tls.LoadX509KeyPair(certPath, keyPath)
    if err != nil {
        log.Printf("Warning: TLS certificate failed to load: %v", err)
    } else {
        s.TLSConfig = &tls.Config{
            Certificates: []tls.Certificate{cert},
            MinVersion:   tls.VersionTLS12,  // Enforce secure TLS version
        }
    }
}
```

## Redis Security

### Authentication
```go
// Use password authentication in production
func NewStore(addr, password string) *Store {
    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,  // Set password for production
        DB:       0,
    })
    return &store{client: client}
}
```

### Environment Configuration
```bash
# Production .env file
REDIS_PASSWORD=your-secure-random-password-here
DOMAIN_NAME=your-domain.com

# Development (no password)
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
```

### TTL and Data Limits
```go
// Set 24-hour TTL on all email data
func (s *store) SaveEmail(ctx context.Context, addressBox string, email store.Email) error {
    key := fmt.Sprintf("inbox:%s", addressBox)
    
    // Save email
    pipe := s.client.Pipeline()
    pipe.LPush(ctx, key, email)
    pipe.Expire(ctx, key, 24*time.Hour)  // Auto-delete after 24 hours
    
    // Limit to 100 emails per inbox
    pipe.LTrim(ctx, key, 0, 99)
    
    _, err := pipe.Exec(ctx)
    return err
}
```

### Redis Configuration
```yaml
services:
  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD:-}
    volumes:
      - redis_data:/data
    networks:
      - coresend-network
    # No ports exposed to host
```

## Environment Security

### Sensitive Data Management
```bash
# .env.example (commit to repo)
DOMAIN_NAME=coresend.io
REDIS_PASSWORD=

# .env (local, never commit)
DOMAIN_NAME=production-domain.com
REDIS_PASSWORD=super-secure-password-123
```

### Docker Secrets (Production)
```yaml
# docker-compose.prod.yml
services:
  backend:
    environment:
      - REDIS_PASSWORD_FILE=/run/secrets/redis_password
    secrets:
      - redis_password

secrets:
  redis_password:
    file: ./secrets/redis_password.txt
```

### Configuration Validation
```go
func LoadConfig() (*Config, error) {
    config := &Config{
        Domain: getEnv("DOMAIN_NAME", ""),
    }
    
    // Validate required configuration
    if config.Domain == "" {
        return nil, errors.New("DOMAIN_NAME is required")
    }
    
    if config.Domain == "localhost" && os.Getenv("ENV") == "production" {
        return nil, errors.New("localhost not allowed in production")
    }
    
    return config, nil
}
```

## Web Security (Caddy)

### Security Headers
```caddyfile
# Caddyfile
header {
    X-Frame-Options DENY
    X-Content-Type-Options nosniff
    X-XSS-Protection "1; mode=block"
    Strict-Transport-Security "max-age=31536000; includeSubDomains"
    -Server  # Hide server signature
}
```

### HTTPS Configuration
```caddyfile
# Automatic HTTPS with Let's Encrypt
{$DOMAIN_NAME}:443 {
    # Force HTTPS
    @http {
        protocol http
    }
    redir @http https://{host}{uri} 301
    
    # Security headers
    header {
        X-Frame-Options DENY
        X-Content-Type-Options nosniff
        X-XSS-Protection "1; mode=block"
        Strict-Transport-Security "max-age=31536000; includeSubDomains"
        -Server
    }
}
```

### Rate Limiting
```caddyfile
# Optional: Add rate limiting
rate_limit {
    zone static_files
    key {remote_host}
    events 100
    window 1m
}
```

## Infrastructure Security

### Terraform Security
```hcl
# terraform/main.tf
resource "oci_core_instance" "coresend" {
  # Use security principles
  display_name = "coresend-server"
  
  # Network security
  vcn_id = oci_core_vpc.coresend_vcn.id
  subnet_id = oci_core_subnet.public.id
  
  # Security lists (firewall rules)
  source_details {
    source_type = "IMAGE"
    source_id = var.instance_image_ocid
  }
  
  # Metadata for cloud-init
  metadata = {
    ssh_authorized_keys = var.ssh_public_key
    user_data = base64encode(file("cloudinit.sh"))
  }
}
```

### Cloud Security Groups
```hcl
# Define security rules
resource "oci_core_security_list" "coresend_security" {
  vcn_id = oci_core_vpc.coresend_vcn.id
  display_name = "coresend-security-list"
  
  # Ingress rules
  ingress_security_rules {
    protocol = "6"  # TCP
    source = "0.0.0.0/0"
    tcp_options {
      min = 22
      max = 22
    }
  }
  
  ingress_security_rules {
    protocol = "6"  # TCP
    source = "0.0.0.0/0"
    tcp_options {
      min = 25
      max = 25
    }
  }
  
  ingress_security_rules {
    protocol = "6"  # TCP
    source = "0.0.0.0/0"
    tcp_options {
      min = 80
      max = 80
    }
  }
  
  ingress_security_rules {
    protocol = "6"  # TCP
    source = "0.0.0.0/0"
    tcp_options {
      min = 443
      max = 443
    }
  }
}
```

## Monitoring and Logging

### Security Logging
```go
// Log security events
func (s *Session) Rcpt(to string, opts *gosmtp.RcptOptions) error {
    localPart := extractLocalPart(to)
    if !identity.IsValidAddress(localPart) {
        // Log rejected addresses for monitoring
        log.Printf("Security: Rejected invalid address %s from %s", to, s.From)
        return &gosmtp.SMTPError{
            Code:         550,
            EnhancedCode: gosmtp.EnhancedCode{5, 1, 1},
            Message:      "Mailbox does not exist",
        }
    }
    return nil
}
```

### Log Management
```yaml
# docker-compose.yml
services:
  backend:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
  
  caddy:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

### Health Checks
```go
// Security-focused health check
func (s *Store) HealthCheck(ctx context.Context) error {
    // Test Redis connection
    if err := s.Ping(ctx); err != nil {
        return fmt.Errorf("redis connection failed: %w", err)
    }
    
    // Test basic operations
    testKey := "health-check"
    testValue := "ok"
    
    if err := s.client.Set(ctx, testKey, testValue, time.Second).Err(); err != nil {
        return fmt.Errorf("redis write failed: %w", err)
    }
    
    if err := s.client.Get(ctx, testKey).Err(); err != nil {
        return fmt.Errorf("redis read failed: %w", err)
    }
    
    // Cleanup
    s.client.Del(ctx, testKey)
    
    return nil
}
```

## Security Checklist

### Production Deployment
- [ ] Set strong Redis password
- [ ] Configure firewall rules
- [ ] Use production domain (not localhost)
- [ ] Enable HTTPS with valid certificates
- [ ] Set up log rotation
- [ ] Configure monitoring
- [ ] Review container resource limits
- [ ] Verify non-root execution

### Code Security
- [ ] Input validation for all user inputs
- [ ] Error messages don't leak information
- [ ] No hardcoded secrets
- [ ] Use parameterized queries (if using DB)
- [ ] Implement rate limiting (if needed)
- [ ] Security headers configured
- [ ] TLS encryption enabled

### Infrastructure Security
- [ ] Regular security updates
- [ ] Minimal exposed ports
- [ ] Network segmentation
- [ ] Backup and recovery plan
- [ ] Access control and monitoring
- [ ] Incident response plan

## Common Security Issues

### What CoreSend Does NOT Provide
- **Email encryption**: Emails stored in plain text in Redis
- **Sender verification**: No SPF/DKIM/DMARC validation
- **Spam filtering**: All valid-addressed emails are accepted
- **Long-term storage**: Emails deleted after 24 hours

### Mitigation Strategies
```go
// Limit email size to prevent storage exhaustion
s.MaxMessageBytes = 1024 * 1024  // 1MB limit

// Limit recipients to prevent abuse
s.MaxRecipients = 50

// Set connection timeouts
s.ReadTimeout = 10 * time.Second
s.WriteTimeout = 10 * time.Second
```

### Security Monitoring
```bash
# Monitor for suspicious activity
docker-compose logs -f backend | grep "Security:"

# Monitor Redis connections
docker-compose exec redis redis-cli info clients

# Monitor system resources
docker stats
```

## Incident Response

### Security Events to Monitor
- Repeated invalid address attempts
- Unusual email volume spikes
- Failed authentication attempts
- Resource exhaustion attacks
- Certificate expiration

### Response Procedures
1. **Detection**: Monitor logs and metrics
2. **Assessment**: Determine impact and scope
3. **Containment**: Block malicious IPs if needed
4. **Eradication**: Patch vulnerabilities
5. **Recovery**: Restore normal operations
6. **Lessons Learned**: Update security measures

---

This security guide provides comprehensive security practices for the CoreSend project. For development guidelines, see [DEVELOPMENT.md](DEVELOPMENT.md). For coding standards, see [CODING_STANDARDS.md](CODING_STANDARDS.md).