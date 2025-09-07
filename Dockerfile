# Dockerfile for macOS NAT Manager
# This is primarily used for CI/CD testing and documentation

FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    gcc \
    musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN make build-release

# Create minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    dnsmasq \
    iptables \
    bash

# Create non-root user
RUN addgroup -g 1000 natmanager && \
    adduser -D -s /bin/bash -u 1000 -G natmanager natmanager

# Copy binary from builder
COPY --from=builder /app/nat-manager /usr/local/bin/nat-manager

# Copy scripts and documentation
COPY --from=builder /app/scripts/ /usr/local/bin/
COPY --from=builder /app/README.md /usr/local/share/doc/nat-manager/
COPY --from=builder /app/CHANGELOG.md /usr/local/share/doc/nat-manager/

# Create configuration directory
RUN mkdir -p /etc/nat-manager && \
    chown natmanager:natmanager /etc/nat-manager

# Copy default configuration
COPY --from=builder /app/configs/default.yaml /etc/nat-manager/config.yaml

# Create entrypoint script
COPY scripts/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Set permissions
RUN chmod +x /usr/local/bin/nat-manager

# Labels for metadata
LABEL org.opencontainers.image.title="macOS NAT Manager"
LABEL org.opencontainers.image.description="True NAT with address translation for macOS"
LABEL org.opencontainers.image.vendor="Your Name"
LABEL org.opencontainers.image.url="https://github.com/scttfrdmn/macos-nat-manager"
LABEL org.opencontainers.image.source="https://github.com/scttfrdmn/macos-nat-manager"
LABEL org.opencontainers.image.documentation="https://github.com/scttfrdmn/macos-nat-manager/blob/main/README.md"
LABEL org.opencontainers.image.licenses="MIT"

# Expose documentation
VOLUME ["/usr/local/share/doc/nat-manager"]

# Use natmanager user
USER natmanager

# Set entrypoint
ENTRYPOINT ["/docker-entrypoint.sh"]

# Default command
CMD ["--help"]