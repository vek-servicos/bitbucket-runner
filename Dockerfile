# Multi-stage build for bitbucket-runner
FROM golang:1.21.5-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o bitbucket-runner .

# Final stage
FROM docker:25-dind

# Install bash and other utilities
RUN apk add --no-cache bash curl git openssh-client

# Copy the binary from builder
COPY --from=builder /build/bitbucket-runner /usr/local/bin/bitbucket-runner

# Set working directory
WORKDIR /workspace

# Create non-root user
RUN addgroup -g 1001 runner && \
    adduser -D -u 1001 -G runner runner

# Set permissions
RUN chmod +x /usr/local/bin/bitbucket-runner

# Switch to non-root user
USER runner

# Default command
ENTRYPOINT ["bitbucket-runner"]
CMD ["--help"]