# Proyecto_1 - Terraform Outputs
# Export important information about the infrastructure

# Application Load Balancer
output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = aws_lb.web_server.dns_name
}

output "alb_zone_id" {
  description = "Zone ID of the Application Load Balancer"
  value       = aws_lb.web_server.zone_id
}

# Auto Scaling Group
output "web_server_asg_id" {
  description = "ID of the Auto Scaling Group"
  value       = aws_autoscaling_group.web_server.id
}

# RDS Database
output "rds_endpoint" {
  description = "Full RDS endpoint (includes port)"
  value       = aws_db_instance.main.endpoint
}

output "rds_address" {
  description = "RDS address (without port)"
  value       = aws_db_instance.main.address
}

output "rds_port" {
  description = "RDS port"
  value       = aws_db_instance.main.port
}

# S3 Bucket
output "s3_bucket_name" {
  description = "S3 bucket name"
  value       = aws_s3_bucket.videos.id
}

output "s3_bucket_arn" {
  description = "S3 bucket ARN"
  value       = aws_s3_bucket.videos.arn
}

# SQS Queue
output "sqs_queue_url" {
  description = "SQS queue URL"
  value       = aws_sqs_queue.video_processing.url
}

output "sqs_queue_arn" {
  description = "SQS queue ARN"
  value       = aws_sqs_queue.video_processing.arn
}

# ECR Repositories
output "ecr_api_repository_url" {
  description = "ECR repository URL for API"
  value       = aws_ecr_repository.api.repository_url
}

output "ecr_worker_repository_url" {
  description = "ECR repository URL for Worker"
  value       = aws_ecr_repository.worker.repository_url
}

output "ecr_frontend_repository_url" {
  description = "ECR repository URL for Frontend"
  value       = aws_ecr_repository.frontend.repository_url
}

# CloudFront Distribution
output "cloudfront_domain" {
  description = "CloudFront distribution domain name"
  value       = aws_cloudfront_distribution.frontend.domain_name
}

output "cloudfront_distribution_id" {
  description = "CloudFront distribution ID"
  value       = aws_cloudfront_distribution.frontend.id
}

# Frontend S3 Bucket
output "frontend_s3_bucket" {
  description = "Frontend S3 bucket name"
  value       = aws_s3_bucket.frontend.id
}

# Security Groups
output "web_server_security_group_id" {
  description = "Security group ID for web server"
  value       = aws_security_group.web_server.id
}

output "rds_security_group_id" {
  description = "Security group ID for RDS"
  value       = aws_security_group.rds.id
}

# IAM Roles
output "web_server_iam_role_arn" {
  description = "IAM role ARN for web server"
  value       = aws_iam_role.web_server.arn
}

# Region
output "aws_region" {
  description = "AWS region"
  value       = var.aws_region
}



# Application URL
output "application_url" {
  description = "Application URL (via Application Load Balancer)"
  value       = "http://${aws_lb.web_server.dns_name}"
}

# Deployment Summary
output "deployment_summary" {
  description = "Summary of deployed infrastructure"
  value       = <<-EOT
    
    ============================================
    🎉 Deployment Complete!
    ============================================
    
    📍 Region: ${var.aws_region}
    🏷️  Environment: ${var.environment}
    
    🌐 Web Server (Auto Scaling Group)
    ├─ ALB DNS:    ${aws_lb.web_server.dns_name}
    ├─ ASG ID:     ${aws_autoscaling_group.web_server.id}
    ├─ Min Size:   ${var.web_server_asg_min_size}
    ├─ Max Size:   ${var.web_server_asg_max_size}
    ├─ Desired:    ${var.web_server_asg_desired_capacity}
    └─ Access:     Use AWS Systems Manager Session Manager
    
    ⚙️  Lambda Worker
    ├─ Function:   ${aws_lambda_function.video_processor.function_name}
    ├─ Runtime:    ${aws_lambda_function.video_processor.package_type}
    ├─ Memory:     ${aws_lambda_function.video_processor.memory_size} MB
    ├─ Timeout:    ${aws_lambda_function.video_processor.timeout}s
    ├─ Log Group:  ${aws_cloudwatch_log_group.lambda_worker.name}
    └─ Max Concurrency: ${aws_lambda_function.video_processor.reserved_concurrent_executions}
    
    🗄️  Database (RDS)
    ├─ Endpoint:   ${aws_db_instance.main.endpoint}
    ├─ Address:    ${aws_db_instance.main.address}
    ├─ Port:       ${aws_db_instance.main.port}
    ├─ Database:   ${var.db_name}
    └─ Username:   ${var.db_username}
    
    📦 S3 Bucket
    └─ Name:       ${aws_s3_bucket.videos.id}
    
    📨 SQS Queue
    ├─ Name:       ${aws_sqs_queue.video_processing.name}
    └─ URL:        ${aws_sqs_queue.video_processing.url}
    
    🐳 ECR Repositories
    ├─ API:        ${aws_ecr_repository.api.repository_url}
    └─ Worker:     ${aws_ecr_repository.worker.repository_url}
    
    🌐 Application
    └─ URL:        http://${aws_lb.web_server.dns_name}
    
    ============================================
    📋 Next Steps:
    ============================================
    
    1. Initialize database:
       cd .. && ./terraform/scripts/init-db.sh
    
    2. Build and push Docker images:
       cd .. && ./terraform/scripts/push-images.sh
    
    3. Deploy application:
       cd .. && ./terraform/scripts/deploy-app.sh
    
    4. Verify deployment:
       curl http://${aws_lb.web_server.dns_name}/api/health
    
    ============================================
  EOT
}

