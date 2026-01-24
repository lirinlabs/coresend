# Start with a small Go image
FROM golang:1.25-alpine AS builder

# Set working directory inside the container
WORKDIR /app

# Install ca-certificates for SSL/TLS
RUN apk add --no-cache ca-certificates

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application with optimizations
# Limits the compiler to work as single-thread (my current machine can't use the command below)
# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main ./cmd/server/main.go
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
    CMD nc -z localhost 1025 && nc -z localhost 8080 || exit 1

CMD ["./main"]