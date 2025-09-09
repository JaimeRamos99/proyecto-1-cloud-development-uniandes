#!/bin/bash

# LocalStack initialization script for AWS resources
# This script runs when LocalStack is ready and creates the necessary S3 buckets and SQS queues

set -e

echo "Starting AWS resources setup..."

# Set AWS CLI configuration for LocalStack
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-east-1
export AWS_ENDPOINT_URL=http://localhost:4566

# Wait for LocalStack to be ready
until curl -s http://localhost:4566/_localstack/health | grep -q '"s3": "available"'; do
    echo "Waiting for LocalStack S3 to be ready..."
    sleep 2
done

until curl -s http://localhost:4566/_localstack/health | grep -q '"sqs": "available"'; do
    echo "Waiting for LocalStack SQS to be ready..."
    sleep 2
done

echo "LocalStack is ready, setting up resources..."

# Create S3 bucket for video storage
echo "Creating S3 bucket: proyecto1-videos"
awslocal s3 mb s3://proyecto1-videos

# Set bucket CORS configuration for web uploads
echo "Setting CORS configuration for S3 bucket..."
awslocal s3api put-bucket-cors --bucket proyecto1-videos --cors-configuration '{
  "CORSRules": [
    {
      "AllowedHeaders": ["*"],
      "AllowedMethods": ["GET", "PUT", "POST", "DELETE", "HEAD"],
      "AllowedOrigins": ["*"],
      "ExposeHeaders": ["ETag", "x-amz-request-id"],
      "MaxAgeSeconds": 3000
    }
  ]
}'

# Create SQS queue for video processing
echo "Creating SQS queue: proyecto1-video-processing"
awslocal sqs create-queue --queue-name proyecto1-video-processing

# Create Dead Letter Queue for failed video processing
echo "Creating SQS dead letter queue: proyecto1-video-processing-dlq"
awslocal sqs create-queue --queue-name proyecto1-video-processing-dlq

# Get queue URLs and ARNs
VIDEO_QUEUE_URL=$(awslocal sqs get-queue-url --queue-name proyecto1-video-processing --query 'QueueUrl' --output text)
DLQ_URL=$(awslocal sqs get-queue-url --queue-name proyecto1-video-processing-dlq --query 'QueueUrl' --output text)
DLQ_ARN=$(awslocal sqs get-queue-attributes --queue-url "$DLQ_URL" --attribute-names QueueArn --query 'Attributes.QueueArn' --output text)

# Set up redrive policy for main queue to use DLQ
echo "Setting up dead letter queue redrive policy..."
awslocal sqs set-queue-attributes --queue-url "$VIDEO_QUEUE_URL" --attributes '{
  "RedrivePolicy": "{\"deadLetterTargetArn\":\"'$DLQ_ARN'\",\"maxReceiveCount\":3}",
  "VisibilityTimeoutSeconds": "300",
  "MessageRetentionPeriod": "1209600"
}'

echo "AWS resources setup completed successfully!"

# List created resources for verification
echo ""
echo "Created resources:"
echo "S3 Buckets:"
awslocal s3 ls

echo ""
echo "SQS Queues:"
awslocal sqs list-queues

echo ""
echo "LocalStack setup complete!"
