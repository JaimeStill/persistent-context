# Build stage
FROM golang:1.24-alpine3.21 AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the service server application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/persistent-context-svc ./cmd/persistent-context-svc/

# Final stage
FROM alpine:3.21

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates-bundle

# Create non-root user
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Create data directory
RUN mkdir -p /data/personas && \
    chown -R appuser:appgroup /data

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/bin/persistent-context-svc .

# Change ownership to non-root user
RUN chown appuser:appgroup /app/persistent-context-svc

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8543

# Health check using /ready endpoint
HEALTHCHECK --interval=10s --timeout=5s --start-period=20s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8543/ready || exit 1

# Run the application
CMD ["./persistent-context-svc"]