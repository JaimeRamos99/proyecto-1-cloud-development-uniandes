# cloudwatch.tf - CloudWatch Monitoring and Dashboards

# CloudWatch Dashboard
resource "aws_cloudwatch_dashboard" "main" {
  dashboard_name = "${var.project_name}-dashboard"

  dashboard_body = jsonencode({
    widgets = [
      # EC2 Metrics Widget
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/EC2", "CPUUtilization", { stat = "Average" }],
            [".", "NetworkIn", { stat = "Sum", yAxis = "right" }],
            [".", "NetworkOut", { stat = "Sum", yAxis = "right" }]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "EC2 Instance Metrics"
          yAxis = {
            left = {
              label = "CPU %"
              min   = 0
              max   = 100
            }
            right = {
              label = "Network (Bytes)"
            }
          }
        }
        width  = 12
        height = 6
      },
      # Auto Scaling Metrics Widget
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/AutoScaling", "GroupDesiredCapacity", "AutoScalingGroupName", aws_autoscaling_group.api.name],
            [".", "GroupInServiceInstances", ".", "."],
            [".", "GroupMinSize", ".", "."],
            [".", "GroupMaxSize", ".", "."]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "Auto Scaling Group Metrics"
        }
        width  = 12
        height = 6
      },
      # ALB Metrics Widget
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/ApplicationELB", "TargetResponseTime", "LoadBalancer", aws_lb.api.arn_suffix],
            [".", "RequestCount", ".", ".", { stat = "Sum", yAxis = "right" }],
            [".", "HTTPCode_Target_2XX_Count", ".", ".", { stat = "Sum", yAxis = "right" }],
            [".", "HTTPCode_Target_4XX_Count", ".", ".", { stat = "Sum", yAxis = "right" }],
            [".", "HTTPCode_Target_5XX_Count", ".", ".", { stat = "Sum", yAxis = "right" }]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "Load Balancer Metrics"
        }
        width  = 12
        height = 6
      },
      # RDS Metrics Widget
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/RDS", "CPUUtilization", "DBInstanceIdentifier", aws_db_instance.main.id],
            [".", "DatabaseConnections", ".", "."],
            [".", "FreeableMemory", ".", ".", { yAxis = "right" }]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "RDS Database Metrics"
        }
        width  = 12
        height = 6
      },
      # S3 Metrics Widget
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/S3", "BucketSizeBytes", "BucketName", aws_s3_bucket.videos.id, "StorageType", "StandardStorage", { stat = "Average" }],
            [".", "NumberOfObjects", ".", ".", ".", "AllStorageTypes", { stat = "Average", yAxis = "right" }]
          ]
          period = 86400
          stat   = "Average"
          region = var.aws_region
          title  = "S3 Video Storage Metrics"
        }
        width  = 12
        height = 6
      },
      # SQS Metrics Widget
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/SQS", "NumberOfMessagesSent", "QueueName", var.sqs_queue_name, { stat = "Sum" }],
            [".", "NumberOfMessagesReceived", ".", ".", { stat = "Sum" }],
            [".", "ApproximateNumberOfMessagesVisible", ".", "."],
            [".", "ApproximateNumberOfMessagesDelayed", ".", "."]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "SQS Queue Metrics"
        }
        width  = 12
        height = 6
      }
    ]
  })
}

# CloudWatch Log Groups
resource "aws_cloudwatch_log_group" "api" {
  name              = "/aws/ec2/${var.project_name}/api"
  retention_in_days = var.log_retention_days

  tags = {
    Name        = "${var.project_name}-api-logs"
    Environment = var.environment
  }
}

resource "aws_cloudwatch_log_group" "worker" {
  name              = "/aws/ec2/${var.project_name}/worker"
  retention_in_days = var.log_retention_days

  tags = {
    Name        = "${var.project_name}-worker-logs"
    Environment = var.environment
  }
}

# CloudWatch Alarms for RDS
resource "aws_cloudwatch_metric_alarm" "rds_cpu" {
  alarm_name          = "${var.project_name}-rds-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name        = "CPUUtilization"
  namespace          = "AWS/RDS"
  period             = 300
  statistic          = "Average"
  threshold          = 80
  alarm_description  = "RDS instance high CPU"
  alarm_actions      = var.sns_topic_arn != "" ? [var.sns_topic_arn] : []

  dimensions = {
    DBInstanceIdentifier = aws_db_instance.main.id
  }
}

resource "aws_cloudwatch_metric_alarm" "rds_storage" {
  alarm_name          = "${var.project_name}-rds-storage-low"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 1
  metric_name        = "FreeStorageSpace"
  namespace          = "AWS/RDS"
  period             = 300
  statistic          = "Average"
  threshold          = 2147483648  # 2GB in bytes
  alarm_description  = "RDS instance low storage"
  alarm_actions      = var.sns_topic_arn != "" ? [var.sns_topic_arn] : []

  dimensions = {
    DBInstanceIdentifier = aws_db_instance.main.id
  }
}

# CloudWatch Alarms for ALB
resource "aws_cloudwatch_metric_alarm" "alb_unhealthy_hosts" {
  alarm_name          = "${var.project_name}-alb-unhealthy-hosts"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name        = "UnHealthyHostCount"
  namespace          = "AWS/ApplicationELB"
  period             = 60
  statistic          = "Average"
  threshold          = 0
  alarm_description  = "Alert when we have any unhealthy hosts"
  alarm_actions      = var.sns_topic_arn != "" ? [var.sns_topic_arn] : []

  dimensions = {
    TargetGroup  = aws_lb_target_group.api.arn_suffix
    LoadBalancer = aws_lb.api.arn_suffix
  }
}

resource "aws_cloudwatch_metric_alarm" "alb_response_time" {
  alarm_name          = "${var.project_name}-alb-response-time"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name        = "TargetResponseTime"
  namespace          = "AWS/ApplicationELB"
  period             = 300
  statistic          = "Average"
  threshold          = 2
  alarm_description  = "Alert when response time is high"
  alarm_actions      = var.sns_topic_arn != "" ? [var.sns_topic_arn] : []

  dimensions = {
    LoadBalancer = aws_lb.api.arn_suffix
  }
}

# CloudWatch Alarms for SQS
resource "aws_cloudwatch_metric_alarm" "sqs_messages_high" {
  alarm_name          = "${var.project_name}-sqs-messages-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name        = "ApproximateNumberOfMessagesVisible"
  namespace          = "AWS/SQS"
  period             = 300
  statistic          = "Average"
  threshold          = 100
  alarm_description  = "Alert when too many messages in queue"
  alarm_actions      = var.sns_topic_arn != "" ? [var.sns_topic_arn] : []

  dimensions = {
    QueueName = var.sqs_queue_name
  }
}

# SNS Topic for Alarms (Optional)
resource "aws_sns_topic" "alarms" {
  count = var.create_sns_topic ? 1 : 0
  name  = "${var.project_name}-alarms"

  tags = {
    Name        = "${var.project_name}-alarms"
    Environment = var.environment
  }
}

resource "aws_sns_topic_subscription" "alarm_email" {
  count     = var.create_sns_topic ? 1 : 0
  topic_arn = aws_sns_topic.alarms[0].arn
  protocol  = "email"
  endpoint  = var.alarm_email
}

# Outputs
output "cloudwatch_dashboard_url" {
  value       = "https://console.aws.amazon.com/cloudwatch/home?region=${var.aws_region}#dashboards:name=${aws_cloudwatch_dashboard.main.dashboard_name}"
  description = "URL to CloudWatch Dashboard"
}

output "sns_topic_arn" {
  value       = var.create_sns_topic ? aws_sns_topic.alarms[0].arn : ""
  description = "SNS Topic ARN for alarms"
}