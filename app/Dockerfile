FROM oven/bun:1.3-alpine AS builder
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install
COPY . .
RUN bun run build --minify

FROM caddy:2.11.1-alpine
# Copy built assets to staging directory (not directly to /srv/frontend)
# The entrypoint script will copy to the shared volume at runtime
COPY --from=builder /app/dist /srv/frontend-build
COPY Caddyfile /etc/caddy/Caddyfile
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh
EXPOSE 80 443 443/udp
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]