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
WORKDIR /root/

# Create a non-root user
RUN addgroup -S worker && adduser -S worker -G worker

# Copy the built binary from builder stage
COPY --from=builder /app/bin/worker ./

# Copy assets directory for watermark
COPY assets/ /app/assets/

# Create directories with proper permissions (before switching user)
RUN mkdir -p /tmp && chmod 1777 /tmp
RUN mkdir -p /app/assets && chmod 755 /app/assets
RUN mkdir -p /app/shared-files && chmod 755 /app/shared-files

# Create ANB watermark using ImageMagick only if watermark.png doesn't exist
RUN if [ ! -f /app/assets/watermark.png ]; then \
    convert -size 200x50 xc:none \
        -font Arial-Bold -pointsize 24 \
        -fill 'rgba(255,255,255,0.8)' \
        -stroke 'rgba(0,0,0,0.5)' -strokewidth 1 \
        -gravity center -annotate 0 "ANB" \
        /app/assets/watermark.png || \
        echo "Warning: Could not create watermark with ImageMagick"; \
    else \
        echo "Using existing watermark.png"; \
    fi

# Change ownership to worker user  
RUN chown -R worker:worker ./worker /app/assets /app/shared-files
# Ensure /tmp is writable by worker user (standard tmp directory permissions)
RUN chmod 1777 /tmp

# Switch to non-root user
USER worker

# Run the application
CMD ["./worker"]