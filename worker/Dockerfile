# Build stage
FROM golang:1.23.12-alpine AS builder

# Set working directory in the container
WORKDIR /app

# Install required packages for building and running
RUN apk add --no-cache gcc musl-dev

# Copy go mod files first to leverage Docker cache
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bin/worker cmd/worker/main.go

# Final stage
FROM alpine:3.18

# Install necessary packages including ffmpeg for video processing
# (keeping this for future video processing capabilities)
RUN apk --no-cache add ca-certificates tzdata ffmpeg

# Set working directory
WORKDIR /root/

# Create a non-root user
RUN addgroup -S worker && adduser -S worker -G worker

# Copy the built binary from builder stage
COPY --from=builder /app/bin/worker ./

# Create directory for temporary files with proper permissions (before switching user)
RUN mkdir -p /tmp && chmod 755 /tmp

# Change ownership to worker user
RUN chown worker:worker ./worker

# Switch to non-root user
USER worker

# Expose any necessary ports (worker doesn't need HTTP port)
# EXPOSE 8080

# Health check (optional - worker doesn't have HTTP endpoints)
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD ps aux | grep -q '[w]orker' || exit 1

# Run the application
CMD ["./worker"]
