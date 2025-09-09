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

### Video Processing Implementation ✅ COMPLETED

The worker now implements **complete video processing** with all requirements:

#### 🎯 **Implemented Features:**

1. **Duration Clipping**: Automatically limits videos to **30 seconds maximum**
2. **Resolution & Aspect Ratio**: Converts to **1280x720 (720p)** with **16:9 aspect ratio**
3. **No Content Cropping**: Uses **Opción B** - maintains all original content with black bars if needed
4. **Audio Removal**: Completely removes audio tracks (`-an`)
5. **ANB Watermark**: Adds ANB logo in top-right corner with 10px margin
6. **File Management**: Preserves original in `original/` folder, saves processed in `processed/`

#### 🔧 **Processing Pipeline:**

```
1. Download from S3 → 2. Backup to original/ → 3. Process with FFmpeg → 4. Upload to processed/ → 5. Update DB status
```

#### 📋 **FFmpeg Command Used:**

```bash
ffmpeg -i input.mp4 -i watermark.png \
  -t 30 \
  -filter_complex "[0:v]scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2:black[scaled];[scaled][1:v]overlay=main_w-overlay_w-10:10" \
  -an -c:v libx264 -crf 23 -preset medium -pix_fmt yuv420p -movflags +faststart \
  output.mp4
```

#### 🎨 **Visual Result Example:**

```
Original 4:3 video → Final 16:9 (1280x720):
┌────────────────────────────────────────┐
│▓▓│                              │▓▓│   │ ← Black bars (no content loss)
│▓▓│     COMPLETE VIDEO           │▓▓│   │
│▓▓│     CONTENT PRESERVED        │▓▓│ANB│ ← ANB watermark
│▓▓│     (Opción B)               │▓▓│   │
└────────────────────────────────────────┘
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

### Log Output Examples:

```
Processing video ID 123: Original->original/123.mp4, Processed->processed/123.mp4
Processing video file (ID: 123, Size: 15728640 bytes) - applying transformations
Starting video processing for ID 123 with Opción B (sin recorte)
Executing FFmpeg with Opción B (sin recorte): ffmpeg -i /tmp/input_123.mp4 -i /app/assets/watermark.png -t 30 ...
Video processing completed for ID 123. Original: 15728640 bytes, Processed: 8234567 bytes
Successfully processed video: 123 (Original: original/123.mp4, Processed: processed/123.mp4)
Transformations applied: ≤30s, 1280x720, 16:9, no audio, ANB watermark, no content cropping
```

## File Structure After Processing

```
S3 Bucket Layout:
├── original/
│   ├── 1.mp4          ← Original uploaded videos (preserved)
│   ├── 2.mp4
│   └── ...
├── processed/
│   ├── 1.mp4          ← Processed videos (30s, 720p, 16:9, no audio, ANB watermark)
│   ├── 2.mp4
│   └── ...
└── [legacy files]     ← Original upload location (for backwards compatibility)
```

## Testing the Complete Pipeline

1. **Start all services**: `make local`
2. **Upload a video** through the API:
   ```bash
   curl -X POST http://localhost:80/api/videos/upload \
     -H "Authorization: Bearer YOUR_JWT" \
     -F "file=@test_video.mp4" \
     -F "title=Test Video"
   ```
3. **Monitor processing**: `make worker-logs`
4. **Check results**:
   - Original video: `original/{video_id}.mp4`
   - Processed video: `processed/{video_id}.mp4`
   - Database status: `processed`

## 🛡️ Resiliencia y Manejo de Errores

El worker incluye mecanismos avanzados de resiliencia:

### Exponential Backoff

- **Reintentos automáticos** con retraso exponencial (2s, 4s, 8s...)
- **Detección inteligente** de errores permanentes vs temporales
- **Configuración flexible** vía variables de entorno

### Dead Letter Queue (DLQ)

- **Cola separada** para mensajes que fallan definitivamente
- **Visibilidad completa** de videos problemáticos
- **Configuración automática** en LocalStack

### Variables de Configuración:

```bash
WORKER_ENABLE_BACKOFF=true    # Activar/desactivar reintentos
WORKER_MAX_RETRIES=3          # Máximo reintentos
WORKER_BASE_DELAY=2           # Retraso base en segundos
WORKER_MAX_DELAY=60           # Retraso máximo en segundos
DLQ_QUEUE_NAME=proyecto1-video-processing-dlq
```

**📖 Documentación completa**: Ver [`RESILIENCE_IMPLEMENTATION.md`](./RESILIENCE_IMPLEMENTATION.md) para detalles técnicos completos.
