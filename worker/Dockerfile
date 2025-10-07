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

# Install necessary packages including ffmpeg for video processing and ImageMagick for watermark
RUN apk --no-cache add ca-certificates tzdata ffmpeg imagemagick

# Set working directory
WORKDIR /app/

# Create a non-root user
RUN addgroup -S worker && adduser -S worker -G worker

# Copy the built binary from builder stage
COPY --from=builder /app/bin/worker ./worker

# Copy assets directory for watermark
COPY assets/ ./assets/

# Create directories with proper permissions (before switching user)
RUN mkdir -p /tmp && chmod 1777 /tmp
RUN mkdir -p ./assets && chmod 755 ./assets

# Create ANB watermark using ImageMagick only if watermark.png doesn't exist
RUN if [ ! -f ./assets/watermark.png ]; then \
    convert -size 200x50 xc:none \
        -font Arial-Bold -pointsize 24 \
        -fill 'rgba(255,255,255,0.8)' \
        -stroke 'rgba(0,0,0,0.5)' -strokewidth 1 \
        -gravity center -annotate 0 "ANB" \
        ./assets/watermark.png || \
        echo "Warning: Could not create watermark with ImageMagick"; \
    else \
        echo "Using existing watermark.png"; \
    fi

# Create ANB bumpers (intro/outro) if they don't exist
RUN if [ ! -f ./assets/intro.mp4 ] || [ ! -f ./assets/outro.mp4 ]; then \
    if [ -f ./assets/create_bumpers.sh ]; then \
        cd ./assets && sh create_bumpers.sh || \
        echo "Warning: Could not create bumpers"; \
    fi; \
    else \
        echo "Using existing bumpers"; \
    fi

# Change ownership to worker user  
RUN chown -R worker:worker ./worker ./assets
# Ensure /tmp is writable by worker user (standard tmp directory permissions)
RUN chmod 1777 /tmp

# Switch to non-root user
USER worker

# Expose any necessary ports (worker doesn't need HTTP port)
# EXPOSE 8080

# Health check (optional - worker doesn't have HTTP endpoints)
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD ps aux | grep -q '[w]orker' || exit 1

# Run the application
CMD ["./worker"]
