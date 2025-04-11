# Build stage
FROM golang:1.24 AS builder

WORKDIR /app

# Copy go module files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o secure-shell-server ./cmd/secure-shell

# Final lightweight image
FROM alpine:latest

# Add ca-certificates for secure connections
RUN apk --no-cache add ca-certificates

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy built executable from builder stage
COPY --from=builder /app/secure-shell-server /app/

# Copy examples for demonstration
COPY --from=builder /app/examples /app/examples

# Use non-root user
USER appuser

# Set entrypoint
ENTRYPOINT ["/app/secure-shell-server"]

# Default command
CMD ["--help"]
