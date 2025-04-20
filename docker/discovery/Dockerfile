FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/discovery ./cmd/discovery

# Create the final image
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/discovery /app/discovery

# Create a non-root user
RUN adduser -D -g '' appuser

# Switch to non-root user
USER appuser

# Set the entrypoint
ENTRYPOINT ["/app/discovery"] 