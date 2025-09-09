# Proyecto_1 Worker Service

The Worker service is a separate Go application that processes video files asynchronously by consuming messages from an SQS queue.

## Overview

This worker service:

1. **Polls SQS Queue**: Uses long polling (20-second wait time) to efficiently receive video processing messages
2. **Downloads Videos**: Retrieves video files from S3 for processing
3. **Processes Videos**: Applies video modifications (currently placeholder - to be implemented based on requirements)
4. **Uploads Results**: Overwrites the original video file in S3 with the processed version
5. **Updates Database**: Changes video status from "uploaded" to "processed"

## Architecture

The worker follows the same patterns as the main api service:

- **Clean Architecture**: Separated into layers (config, database, messaging, storage, service)
- **Dependency Injection**: Services are composed with their dependencies
- **Error Handling**: Comprehensive error handling with proper logging
- **Configuration**: Environment variable based configuration
- **Graceful Shutdown**: Proper signal handling for clean shutdowns

## Key Features

### Long Polling

- Uses SQS long polling with maximum wait time (20 seconds)
- Efficient resource usage - only makes requests when messages are available
- Processes up to 10 messages per batch

### Video Processing Pipeline

1. Receive message with S3 key
2. Validate video is in "uploaded" status
3. Download video from S3
4. Process video (placeholder for actual video processing logic)
5. Upload processed video back to S3 (overwrites original)
6. Update database status to "processed"
7. Delete message from SQS queue

### Error Handling

- Failed messages remain in queue for reprocessing
- Individual message failures don't stop batch processing
- Comprehensive logging for debugging

## Running the Worker

### Local Development

```bash
# Build the worker
make build

# Start all services including worker
make local

# View worker logs
make worker-logs

# Rebuild just the worker
make rebuild-worker
```

### Environment Variables

The worker uses the same environment variables as the api service:

**Database Configuration:**

- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_NAME` - Database name (default: proyecto_1)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password (default: password)
- `DB_SSL_MODE` - SSL mode (default: disable)

**AWS Configuration:**

- `AWS_ACCESS_KEY_ID` - AWS access key
- `AWS_SECRET_ACCESS_KEY` - AWS secret key
- `AWS_DEFAULT_REGION` - AWS region (default: us-east-1)
- `AWS_ENDPOINT_URL` - LocalStack endpoint for local development
- `S3_BUCKET_NAME` - S3 bucket for videos (default: proyecto1-videos)
- `SQS_QUEUE_NAME` - SQS queue name (default: proyecto1-video-processing)

**Application Configuration:**

- `APP_NAME` - Application name (default: Proyecto_1_Worker)
- `APP_VERSION` - Application version (default: 1.0.0)
- `APP_ENV` - Environment (default: development)

## Development

### Adding Video Processing Logic

The current video processing is a placeholder in `internal/service.go`:

```go
func (s *WorkerService) processVideo(videoData []byte, video *videos.Video) ([]byte, error) {
    // TODO: Replace this with actual video processing logic
    // Current implementation just returns original data

    // Add your video processing logic here:
    // - Video transcoding
    // - Compression
    // - Watermark addition
    // - Format conversion
    // - Any other video processing operations

    return videoData, nil
}
```

### Testing

The worker can be tested by:

1. Starting the full local environment: `make local`
2. Uploading a video through the API
3. Monitoring worker logs: `make worker-logs`
4. Checking video status changes in the database

## Docker Integration

The worker runs as a separate container that:

- Depends on PostgreSQL and LocalStack
- Shares the same network as other services
- Has no exposed ports (background service)
- Includes FFmpeg for future video processing needs

## Monitoring

Monitor the worker through:

- Container logs: `make worker-logs`
- Database status updates
- SQS queue metrics (through LocalStack or AWS Console)
- Container health: `make health`
