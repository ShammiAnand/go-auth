# Build stage
FROM golang:1.24-alpine AS build

# Install build dependencies
RUN apk --update add --no-cache ca-certificates git make

WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o go-auth .

# Run stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from build stage
COPY --from=build /app/go-auth ./go-auth

# Copy configs
COPY --from=build /app/configs ./configs
COPY --from=build /app/.env.sample ./.env

# Create non-root user
RUN addgroup -g 1000 goauth && \
    adduser -D -u 1000 -G goauth goauth && \
    chown -R goauth:goauth /root

USER goauth

# Expose API port
EXPOSE 42069

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:42069/health || exit 1

# Default command: run server
ENTRYPOINT ["./go-auth"]
CMD ["server"]
