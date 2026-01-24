# Docker Setup Guide

Complete guide for deploying CoreSend with Docker and Caddy.

## Prerequisites

- **Docker**: Version 20.10 or later
- **Docker Compose**: Version 2.0 or later (V2 syntax)
- **Domain**: DNS A record pointing to your server (production only)
- **Ports available**:
  - 22 (SSH - host OS, not Docker)
  - 25 (SMTP)
  - 80 (HTTP)
  - 443 (HTTPS)

### Check Docker Installation

```bash
docker --version
docker compose version
```

## Quick Start

### Development (Local)

For local development, Docker Compose automatically loads `docker-compose.override.yml` which configures everything for localhost without SSL.

```bash
# Clone repository
git clone https://github.com/fn-jakubkarp/coresend.git
cd coresend

# Start all services
docker-compose up -d

# Verify services are running
docker-compose ps

# View logs
docker-compose logs -f
```

Access the application at `http://localhost`

### Production

Production deployment uses only `docker-compose.yml` with your domain and automatic SSL.

```bash
# Clone repository
git clone https://github.com/fn-jakubkarp/coresend.git
cd coresend

# Create environment file
cp .env.example .env

# Configure your domain
nano .env
# Set: DOMAIN_NAME=yourdomain.com

# Start with production config only (ignore override file)
docker-compose -f docker-compose.yml up -d

# Check Caddy obtained SSL certificates
docker-compose logs caddy
```

**Important**: Ensure your DNS is pointing to the server BEFORE starting. Caddy will attempt to obtain SSL certificates immediately.

## Services Overview

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Docker Network                        │
│                  (coresend-network)                      │
│                                                          │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐             │
│  │  Redis  │    │ Backend │    │  Caddy  │             │
│  │  :6379  │◄───│  :1025  │◄───│  :80    │◄── HTTP     │
│  │         │    │  :8080  │    │  :443   │◄── HTTPS    │
│  └─────────┘    └─────────┘    └─────────┘             │
│       │              │              │                   │
│       │              │              ├── Static files    │
│       │              │              └── /api/* proxy    │
│       │              │                                  │
│       │         Port 25 ◄────────────────── SMTP       │
│       │         (host)                                  │
│       │                                                 │
│  [redis_data]  [caddy_certs]  [caddy_data]            │
│    volume        volume         volume                 │
└─────────────────────────────────────────────────────────┘
```

### Service Details

#### Redis
- **Image**: `redis:7-alpine`
- **Purpose**: Email storage
- **Volume**: `redis_data` (persistent)
- **Health check**: `redis-cli ping`

#### Backend
- **Build**: `./backend/Dockerfile`
- **Ports**: 
  - `25:1025` - SMTP (exposed to host)
  - `8080` - API (internal, proxied by Caddy)
- **Depends on**: Redis (healthy)
- **Health check**: `nc -z localhost 1025`

#### Caddy
- **Image**: `caddy:alpine`
- **Ports**:
  - `80:80` - HTTP (redirects to HTTPS)
  - `443:443` - HTTPS
- **Volumes**:
  - `./caddy/Caddyfile` - Configuration
  - `./frontend/dist` - Static files
  - `caddy_data` - SSL certificates
  - `caddy_certs` - Shared certs for backend STARTTLS
- **Depends on**: Backend

## Configuration

### Environment Variables

Create a `.env` file in the project root:

```bash
# Required for production
DOMAIN_NAME=coresend.io

# Optional: Redis password
REDIS_PASSWORD=your-secure-password
```

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DOMAIN_NAME` | Yes (prod) | `localhost` | Your domain name |
| `REDIS_PASSWORD` | No | *(empty)* | Redis authentication |

### Caddyfile Configuration

**Production** (`caddy/Caddyfile`):
- Automatic HTTPS via Let's Encrypt
- HTTP to HTTPS redirect
- www to non-www redirect
- Static file serving
- API reverse proxy
- Security headers

**Development** (`caddy/Caddyfile.dev`):
- HTTP only (port 80)
- No SSL complexity
- Same routing as production

### Override File

`docker-compose.override.yml` is automatically loaded in development:
- Uses `Caddyfile.dev`
- Sets `DOMAIN_NAME=localhost`
- Disables TLS certificates for backend
- Exposes API port 8080 directly

To run production config locally (testing):
```bash
docker-compose -f docker-compose.yml up -d
```

## SSL/TLS Certificates

### How Caddy Handles SSL

Caddy automatically:
1. Detects your domain from the Caddyfile
2. Requests certificates from Let's Encrypt
3. Configures HTTPS
4. Renews certificates before expiry

**No manual certificate management required.**

### SMTP STARTTLS

The backend uses Caddy's certificates for SMTP STARTTLS:
- Caddy stores certs in the `caddy_certs` volume
- Backend mounts this volume read-only
- Certs are automatically renewed by Caddy

### First-Time SSL Setup

1. Ensure DNS points to your server
2. Start services: `docker-compose -f docker-compose.yml up -d`
3. Check Caddy logs: `docker-compose logs caddy`
4. Look for "certificate obtained successfully"

If SSL fails:
- Verify DNS: `dig +short yourdomain.com`
- Check port 80 is accessible (Let's Encrypt HTTP challenge)
- Review Caddy logs for specific errors

## Common Commands

### Service Management

```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# Restart a specific service
docker-compose restart backend

# Rebuild after code changes
docker-compose build backend
docker-compose up -d backend

# Rebuild without cache
docker-compose build --no-cache
```

### Viewing Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f caddy
docker-compose logs -f redis

# Last 100 lines
docker-compose logs --tail=100 backend
```

### Health Checks

```bash
# Check service status
docker-compose ps

# Test SMTP
nc -zv localhost 25

# Test HTTP
curl http://localhost/health

# Test Redis
docker-compose exec redis redis-cli ping
```

### Debugging

```bash
# Shell into container
docker-compose exec backend sh
docker-compose exec caddy sh
docker-compose exec redis sh

# Check container resource usage
docker stats

# Inspect network
docker network inspect coresend_coresend-network
```

## Troubleshooting

### Port 25 Requires Privileges

SMTP port 25 is a privileged port (< 1024). Solutions:

**Option 1**: Run Docker as root (default on most systems)

**Option 2**: Use port mapping (already configured)
```yaml
ports:
  - "25:1025"  # Host 25 -> Container 1025
```

**Option 3**: If still failing, check if another service uses port 25:
```bash
sudo lsof -i :25
sudo netstat -tlnp | grep :25
```

### Caddy Can't Obtain Certificates

1. **DNS not propagated**: Wait or check with `dig +short yourdomain.com`
2. **Port 80 blocked**: Firewall or another service
   ```bash
   sudo ufw allow 80/tcp
   sudo lsof -i :80
   ```
3. **Rate limited**: Let's Encrypt has rate limits. Check logs for details.

### Backend Can't Connect to Redis

```bash
# Check Redis is running
docker-compose ps redis

# Test connectivity from backend
docker-compose exec backend nc -zv redis 6379

# Check network
docker network inspect coresend_coresend-network
```

### Container Keeps Restarting

```bash
# Check logs for errors
docker-compose logs backend

# Common causes:
# - Redis not ready (check depends_on health check)
# - Invalid environment variables
# - Port already in use
```

## Maintenance

### Backups

**Redis Data**:
```bash
# Trigger Redis save
docker-compose exec redis redis-cli BGSAVE

# Copy backup from volume
docker run --rm \
  -v coresend_redis_data:/data \
  -v $(pwd)/backups:/backup \
  alpine tar czf /backup/redis-$(date +%Y%m%d).tar.gz -C /data .
```

**Caddy Certificates** (automatic, but for reference):
```bash
docker run --rm \
  -v coresend_caddy_data:/data \
  -v $(pwd)/backups:/backup \
  alpine tar czf /backup/caddy-$(date +%Y%m%d).tar.gz -C /data .
```

### Updates

```bash
# Pull latest images
docker-compose pull

# Rebuild custom images
docker-compose build --no-cache

# Restart with new images
docker-compose up -d
```

### Cleanup

```bash
# Remove stopped containers
docker-compose down

# Remove containers and volumes (DELETES DATA)
docker-compose down -v

# Prune unused Docker resources
docker system prune -f

# Prune unused volumes (CAREFUL)
docker volume prune
```

## Production Checklist

- [ ] DNS A record points to server
- [ ] Firewall allows ports 22, 25, 80, 443
- [ ] `.env` file created with `DOMAIN_NAME`
- [ ] `REDIS_PASSWORD` set (optional but recommended)
- [ ] Started with `-f docker-compose.yml` (no override)
- [ ] Caddy logs show "certificate obtained"
- [ ] SMTP test successful (`nc -zv yourdomain.com 25`)
- [ ] HTTPS working (`curl https://yourdomain.com/health`)
- [ ] Monitoring/alerting configured
- [ ] Backup strategy in place

## File Reference

| File | Purpose |
|------|---------|
| `docker-compose.yml` | Production service definitions |
| `docker-compose.override.yml` | Development overrides (auto-loaded) |
| `.env` | Environment variables (create from .env.example) |
| `caddy/Caddyfile` | Production Caddy config |
| `caddy/Caddyfile.dev` | Development Caddy config |
| `backend/Dockerfile` | Backend container build |
