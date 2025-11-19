# Proyecto_1 - Lambda Worker Function
# Video processing Lambda function triggered by SQS events
# Replaces EC2 worker with serverless, event-driven architecture

# ============================================================================
# Lambda Function
# ============================================================================

resource "aws_lambda_function" "video_processor" {
  function_name = "${var.project_name}-video-processor"
  description   = "Video processing Lambda function (processes videos with FFmpeg)"
  role          = aws_iam_role.lambda_worker.arn

  # Container image configuration
  package_type = "Image"
  image_uri    = "${aws_ecr_repository.worker.repository_url}:lambda"

  # Performance configuration
  timeout     = 900  # 15 minutes (maximum for Lambda)
  memory_size = 3008 # 3GB RAM (good for video processing)

  # Ephemeral storage for video processing (/tmp)
  ephemeral_storage {
    size = 10240 # 10GB (maximum for Lambda)
  }

  # Use unreserved concurrency pool (scaling controlled by SQS event source mapping)
  # reserved_concurrent_executions = 10  # Removed to avoid account quota issues

  # Environment variables
  environment {
    variables = {
      # Database configuration
      DB_HOST           = aws_db_instance.main.address
      DB_PORT           = aws_db_instance.main.port
      DB_NAME           = var.db_name
      DB_USER           = var.db_username
      DB_PASSWORD       = var.db_password
      DB_SSL_MODE       = "require"
      DB_MAX_OPEN_CONNS = "10"
      DB_MAX_IDLE_CONNS = "5"

      # AWS configuration (AWS_DEFAULT_REGION is reserved and automatically set by Lambda)
      S3_BUCKET_NAME = aws_s3_bucket.videos.id
      SQS_QUEUE_NAME = var.sqs_queue_name

      # App configuration
      APP_NAME    = "Proyecto_1_Lambda_Worker"
      APP_VERSION = "1.0.0"
      APP_ENV     = var.environment

      # Assets path
      ASSETS_PATH = "/app/assets"
    }
  }

  # VPC configuration (required for RDS access)
  # Deployed across multiple AZs for high availability
  vpc_config {
    subnet_ids         = data.aws_subnets.multi_az.ids
    security_group_ids = [aws_security_group.lambda_worker.id]
  }

  # Lifecycle
  depends_on = [
    aws_iam_role_policy.lambda_worker,
    aws_cloudwatch_log_group.lambda_worker
  ]

  tags = {
    Name = "${var.project_name}-video-processor"
    Type = "Lambda"
  }
}

# ============================================================================
# SQS Event Source Mapping
# ============================================================================

resource "aws_lambda_event_source_mapping" "sqs_trigger" {
  event_source_arn = aws_sqs_queue.video_processing.arn
  function_name    = aws_lambda_function.video_processor.arn

  # Batch configuration
  batch_size                         = 1 # Process one video at a time
  maximum_batching_window_in_seconds = 0 # Invoke immediately when message arrives

  # Error handling
  function_response_types = ["ReportBatchItemFailures"]

  # Scaling configuration
  scaling_config {
    maximum_concurrency = 10 # Max 10 concurrent Lambda executions
  }

  # Enabled by default
  enabled = true

  depends_on = [
    aws_iam_role_policy.lambda_worker
  ]
}

# ============================================================================
# IAM Role for Lambda
# ============================================================================

resource "aws_iam_role" "lambda_worker" {
  name        = "${var.project_name}-lambda-worker-role"
  description = "IAM role for Lambda video processor"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "${var.project_name}-lambda-worker-role"
  }
}

# ============================================================================
# IAM Policy for Lambda
# ============================================================================

resource "aws_iam_role_policy" "lambda_worker" {
  name = "${var.project_name}-lambda-worker-policy"
  role = aws_iam_role.lambda_worker.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      # SQS permissions (read and delete messages)
      {
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:ChangeMessageVisibility"
        ]
        Resource = aws_sqs_queue.video_processing.arn
      },
      # S3 permissions (read/write video files)
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject"
        ]
        Resource = "${aws_s3_bucket.videos.arn}/*"
      },
      {
        Effect = "Allow"
        Action = [
          "s3:ListBucket"
        ]
        Resource = aws_s3_bucket.videos.arn
      },
      # CloudWatch Logs permissions (Lambda logs)
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "${aws_cloudwatch_log_group.lambda_worker.arn}:*"
      },
      # VPC permissions (required for VPC configuration)
      {
        Effect = "Allow"
        Action = [
          "ec2:CreateNetworkInterface",
          "ec2:DescribeNetworkInterfaces",
          "ec2:DeleteNetworkInterface",
          "ec2:AssignPrivateIpAddresses",
          "ec2:UnassignPrivateIpAddresses"
        ]
        Resource = "*"
      }
    ]
  })
}

# ============================================================================
# Security Group for Lambda
# ============================================================================

resource "aws_security_group" "lambda_worker" {
  name        = "${var.project_name}-lambda-worker-sg"
  description = "Security group for Lambda video processor"
  vpc_id      = data.aws_vpc.default.id

  # No ingress rules needed (Lambda doesn't accept inbound connections)

  # Allow all outbound traffic (for RDS, S3, SQS access)
  egress {
    description = "Allow all outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-lambda-worker-sg"
  }
}

# Update RDS security group to allow Lambda access
resource "aws_security_group_rule" "rds_from_lambda" {
  type                     = "ingress"
  description              = "PostgreSQL from Lambda worker"
  from_port                = 5432
  to_port                  = 5432
  protocol                 = "tcp"
  security_group_id        = aws_security_group.rds.id
  source_security_group_id = aws_security_group.lambda_worker.id
}

# ============================================================================
# CloudWatch Log Group
# ============================================================================

resource "aws_cloudwatch_log_group" "lambda_worker" {
  name              = "/aws/lambda/${var.project_name}-video-processor"
  retention_in_days = 7

  tags = {
    Name = "${var.project_name}-lambda-worker-logs"
  }
}

# ============================================================================
# CloudWatch Alarms (Monitoring)
# ============================================================================

# Alarm for Lambda errors
resource "aws_cloudwatch_metric_alarm" "lambda_errors" {
  alarm_name          = "${var.project_name}-lambda-errors"
  alarm_description   = "Triggers when Lambda video processor has errors"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = "300" # 5 minutes
  statistic           = "Sum"
  threshold           = "5"
  treat_missing_data  = "notBreaching"

  dimensions = {
    FunctionName = aws_lambda_function.video_processor.function_name
  }

  tags = {
    Name = "${var.project_name}-lambda-errors-alarm"
  }
}

# Alarm for Lambda duration (approaching timeout)
resource "aws_cloudwatch_metric_alarm" "lambda_duration" {
  alarm_name          = "${var.project_name}-lambda-duration"
  alarm_description   = "Triggers when Lambda processing takes too long"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "Duration"
  namespace           = "AWS/Lambda"
  period              = "300" # 5 minutes
  statistic           = "Average"
  threshold           = "600000" # 10 minutes (in milliseconds)
  treat_missing_data  = "notBreaching"

  dimensions = {
    FunctionName = aws_lambda_function.video_processor.function_name
  }

  tags = {
    Name = "${var.project_name}-lambda-duration-alarm"
  }
}

# Alarm for Lambda throttles
resource "aws_cloudwatch_metric_alarm" "lambda_throttles" {
  alarm_name          = "${var.project_name}-lambda-throttles"
  alarm_description   = "Triggers when Lambda invocations are throttled"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "Throttles"
  namespace           = "AWS/Lambda"
  period              = "300" # 5 minutes
  statistic           = "Sum"
  threshold           = "1"
  treat_missing_data  = "notBreaching"

  dimensions = {
    FunctionName = aws_lambda_function.video_processor.function_name
  }

  tags = {
    Name = "${var.project_name}-lambda-throttles-alarm"
  }
}

# ============================================================================
# Outputs
# ============================================================================

output "lambda_function_name" {
  description = "Name of the Lambda function"
  value       = aws_lambda_function.video_processor.function_name
}

output "lambda_function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.video_processor.arn
}

output "lambda_role_arn" {
  description = "ARN of the Lambda IAM role"
  value       = aws_iam_role.lambda_worker.arn
}

output "lambda_log_group" {
  description = "CloudWatch log group for Lambda"
  value       = aws_cloudwatch_log_group.lambda_worker.name
}

