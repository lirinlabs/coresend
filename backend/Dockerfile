FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Explicitly disable parallel builds - allows to run on low resource environments
RUN GOMAXPROCS=1 go build -p 1 -v -o main ./cmd/server/main.go


# --- Final Stage ---
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata netcat-openbsd

WORKDIR /app

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

COPY --from=builder /app/main .

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 1025 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD nc -z localhost 1025 || exit 1

CMD ["./main"]