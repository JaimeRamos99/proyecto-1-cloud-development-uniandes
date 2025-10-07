# Proyecto_1 - SQS Queue
# Message queue for asynchronous video processing

resource "aws_sqs_queue" "video_processing" {
  name                       = var.sqs_queue_name
  visibility_timeout_seconds = var.sqs_visibility_timeout
  message_retention_seconds  = var.sqs_message_retention
  receive_wait_time_seconds  = 20  # Long polling

  # Dead Letter Queue configuration
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.video_processing_dlq.arn
    maxReceiveCount     = 3
  })

  tags = {
    Name = "${var.project_name}-video-processing"
  }
}

# Dead Letter Queue for failed messages
resource "aws_sqs_queue" "video_processing_dlq" {
  name                      = "${var.sqs_queue_name}-dlq"
  message_retention_seconds = var.sqs_message_retention

  tags = {
    Name = "${var.project_name}-video-processing-dlq"
  }
}

