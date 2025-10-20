# Proyecto_1 - Terraform Variables
# Define all configurable parameters

# General Configuration
variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "production"
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "proyecto1"
}

# EC2 Configuration
variable "key_pair_name" {
  description = "Name of the SSH key pair for EC2 instances"
  type        = string
}

variable "web_server_instance_type" {
  description = "EC2 instance type for web server (API + Nginx)"
  type        = string
  default     = "t3.small"  # 2 vCPU, 2 GiB RAM
}

variable "worker_instance_type" {
  description = "EC2 instance type for worker (video processing)"
  type        = string
  default     = "t3.small"
}

variable "root_volume_size" {
  description = "Root volume size in GB for EC2 instances"
  type        = number
  default     = 30
}

variable "allowed_ssh_cidr" {
  description = "CIDR block allowed to SSH into EC2 instances (use your IP/32)"
  type        = string
  default     = "0.0.0.0/0"  # WARNING: Change this to your IP for security
}

# RDS Configuration
variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "proyecto_1"
}

variable "db_username" {
  description = "Database master username"
  type        = string
  default     = "postgres"
}

variable "db_password" {
  description = "Database master password (use secure value)"
  type        = string
  sensitive   = true
}

variable "db_allocated_storage" {
  description = "Allocated storage for RDS in GB"
  type        = number
  default     = 20
}

variable "db_backup_retention_period" {
  description = "Number of days to retain backups"
  type        = number
  default     = 7
}

# S3 Configuration
variable "s3_bucket_name" {
  description = "S3 bucket name for video storage (must be globally unique)"
  type        = string
}

# SQS Configuration
variable "sqs_queue_name" {
  description = "SQS queue name for video processing"
  type        = string
  default     = "proyecto1-video-processing"
}

variable "sqs_visibility_timeout" {
  description = "SQS message visibility timeout in seconds"
  type        = number
  default     = 900  # 15 minutes (enough for video processing)
}

variable "sqs_message_retention" {
  description = "SQS message retention period in seconds"
  type        = number
  default     = 1209600  # 14 days
}

# Application Configuration
variable "jwt_secret" {
  description = "JWT secret for authentication"
  type        = string
  sensitive   = true
}

variable "ecr_image_tag" {
  description = "Docker image tag to deploy"
  type        = string
  default     = "latest"
}

# Tags
variable "additional_tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}

# Auto Scaling Configuration
variable "asg_min_size" {
  description = "Minimum number of instances in Auto Scaling Group"
  type        = number
  default     = 1
}

variable "asg_max_size" {
  description = "Maximum number of instances in Auto Scaling Group"
  type        = number
  default     = 3
}

variable "asg_desired_capacity" {
  description = "Desired number of instances in Auto Scaling Group"
  type        = number
  default     = 2
}

variable "cpu_scale_up_threshold" {
  description = "CPU percentage to trigger scale up"
  type        = number
  default     = 70
}

variable "cpu_scale_down_threshold" {
  description = "CPU percentage to trigger scale down"
  type        = number
  default     = 30
}

variable "cpu_target_value" {
  description = "Target CPU utilization for target tracking scaling"
  type        = number
  default     = 50
}

# CloudWatch Configuration
variable "log_retention_days" {
  description = "CloudWatch log retention in days"
  type        = number
  default     = 7
}

variable "create_sns_topic" {
  description = "Whether to create SNS topic for alarms"
  type        = bool
  default     = false
}

variable "alarm_email" {
  description = "Email address for alarm notifications"
  type        = string
  default     = ""
}

variable "sns_topic_arn" {
  description = "Existing SNS topic ARN for alarms (if not creating new one)"
  type        = string
  default     = ""
}