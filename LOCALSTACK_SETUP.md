# LocalStack Setup for Proyecto 1

This document explains how to use LocalStack for local development with AWS services (S3 and SQS).

## What is LocalStack?

LocalStack is a cloud service emulator that runs in a single container on your laptop or in your CI environment. It provides a testing environment that closely matches the real AWS cloud environment.

## Services Configured

- **S3**: For storing video files and assets
- **SQS**: For handling video processing queues

## Getting Started

### 1. Start the Services

Use the provided docker-compose file to start all services including LocalStack:

```bash
docker-compose -f docker-compose.local.yml up -d
```

This will start:

- PostgreSQL database
- LocalStack (with S3 and SQS)
- Backend API
- Nginx reverse proxy

### 2. Verify LocalStack is Running

Check if LocalStack is healthy:

```bash
curl http://localhost:4566/_localstack/health
```

You should see a response indicating that S3 and SQS services are available.

### 3. Initialized Resources

The LocalStack container automatically creates the following resources:

#### S3 Buckets:

- `proyecto1-videos`: Main bucket for storing video files

#### SQS Queues:

- `proyecto1-video-processing`: Main queue for video processing tasks
- `proyecto1-video-processing-dlq`: Dead letter queue for failed processing

### 4. Accessing LocalStack Services

LocalStack provides a single endpoint for all AWS services:

- **Endpoint**: `http://localhost:4566`
- **Access Key**: `test`
- **Secret Key**: `test`
- **Region**: `us-east-1`

## Using AWS CLI with LocalStack

You can use the AWS CLI to interact with LocalStack services. Install `awslocal` which is a wrapper around AWS CLI:

```bash
pip install awscli-local
```

### S3 Examples:

```bash
# List buckets
awslocal s3 ls

# List objects in the videos bucket
awslocal s3 ls s3://proyecto1-videos/

# Upload a file
awslocal s3 cp test-video.mp4 s3://proyecto1-videos/videos/test-video.mp4
```

### SQS Examples:

```bash
# List queues
awslocal sqs list-queues

# Send a message to the processing queue
awslocal sqs send-message --queue-url http://localhost:4566/000000000000/proyecto1-video-processing --message-body '{"video_id":"123","action":"process"}'

# Receive messages from the queue
awslocal sqs receive-message --queue-url http://localhost:4566/000000000000/proyecto1-video-processing
```

## Environment Variables

The following environment variables are configured for LocalStack:

```env
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
AWS_DEFAULT_REGION=us-east-1
AWS_ENDPOINT_URL=http://localstack:4566
S3_BUCKET_NAME=proyecto1-videos
SQS_QUEUE_NAME=proyecto1-video-processing
```

## Application Integration

The backend application is configured to use LocalStack in development mode. The AWS service clients in `/backend/internal/aws/` will automatically connect to LocalStack when the `AWS_ENDPOINT_URL` environment variable is set.

### Using S3 Service:

```go
s3Service, err := aws.NewS3Service(&config.AWS)
if err != nil {
    log.Fatal(err)
}

// Upload a file
err = s3Service.UploadFile(ctx, "videos/example.mp4", fileReader, "video/mp4")
```

### Using SQS Service:

```go
sqsService, err := aws.NewSQSService(&config.AWS)
if err != nil {
    log.Fatal(err)
}

// Send a video processing message
msg := &aws.VideoProcessingMessage{
    VideoID: "123",
    UserID:  "user456",
    S3Key:   "videos/example.mp4",
    Action:  "process",
}
err = sqsService.SendVideoProcessingMessage(ctx, msg)
```

## Troubleshooting

### LocalStack Health Check Fails

If the LocalStack health check fails, check the logs:

```bash
docker-compose -f docker-compose.local.yml logs localstack
```

### Can't Connect to LocalStack

Make sure LocalStack is running and accessible:

```bash
curl http://localhost:4566/_localstack/health
```

### Persistence

LocalStack is configured with persistence enabled, so your data will survive container restarts. Data is stored in the `localstack_data` Docker volume.

### Reset LocalStack Data

To reset all LocalStack data:

```bash
docker-compose -f docker-compose.local.yml down -v
docker-compose -f docker-compose.local.yml up -d
```

## Production Considerations

For production deployment:

1. Remove or comment out the LocalStack service from docker-compose
2. Update environment variables to use real AWS credentials
3. Set `AWS_ENDPOINT_URL` to empty string or remove it entirely
4. Ensure S3 buckets and SQS queues are created in your AWS account
5. Configure proper IAM roles and policies for your application
