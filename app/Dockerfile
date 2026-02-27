FROM oven/bun:1.3-alpine AS builder
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . .
RUN bun run build --minify --outdir=out 

# --- Final Stage ---
FROM caddy:2.7-alpine
COPY --from=builder /app/out /srv/frontend
COPY Caddyfile /etc/caddy/Caddyfile
RUN chown -R caddy:caddy /srv/frontend
EXPOSE 80 443 443/udp