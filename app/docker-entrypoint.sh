#!/bin/sh
set -e

# Copy built assets from staging directory to the shared volume
# This ensures the volume is populated with the latest build on each container start
if [ -d "/srv/frontend-build" ]; then
    cp -r /srv/frontend-build/* /srv/frontend/ 2>/dev/null || true
    echo "Frontend assets copied to shared volume"
fi

# Run Caddy with the provided arguments, or default to running with Caddyfile
exec caddy run --config /etc/caddy/Caddyfile --adapter caddyfile
