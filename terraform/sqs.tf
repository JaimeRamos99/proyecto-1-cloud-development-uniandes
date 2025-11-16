# Proyecto_1 - SQS Queue
# Message queue for asynchronous video processing

resource "aws_sqs_queue" "video_processing" {
  name = var.sqs_queue_name

  # Visibility timeout: Lambda timeout (900s) + buffer (60s) = 960s (16 minutes)
  # This ensures messages aren't redelivered while Lambda is still processing
  visibility_timeout_seconds = 960

  # Message retention: 24 hours (enough time for retries and debugging)
  message_retention_seconds = 86400

  # Long polling: reduces empty receives and costs
  receive_wait_time_seconds = 20

  # Dead Letter Queue configuration
  # After 3 failed processing attempts, messages go to DLQ for investigation
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.video_processing_dlq.arn
    maxReceiveCount     = 3
  })

  tags = {
    Name = "${var.project_name}-video-processing"
  }
}

# Dead Letter Queue for failed messages
# Messages that fail processing 3 times are moved here for investigation
resource "aws_sqs_queue" "video_processing_dlq" {
  name = "${var.sqs_queue_name}-dlq"

  # Retain failed messages for 14 days for debugging
  message_retention_seconds = 1209600 # 14 days

  tags = {
    Name = "${var.project_name}-video-processing-dlq"
  }
}

