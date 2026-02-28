FROM oven/bun:1.3-alpine AS builder
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install
COPY . .
RUN bun run build --minify

FROM caddy:2.11.1-alpine
COPY --from=builder /app/dist /srv/frontend
COPY Caddyfile /etc/caddy/Caddyfile
EXPOSE 80 443 443/udp