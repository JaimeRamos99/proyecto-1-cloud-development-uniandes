# Proyecto_1 - Terraform Outputs
# Export important information about the infrastructure

# Load Balancer and Auto Scaling
output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = aws_lb.api.dns_name
}

output "api_endpoint" {
  description = "API endpoint URL through Load Balancer"
  value       = "http://${aws_lb.api.dns_name}"
}

output "autoscaling_group_name" {
  description = "Name of the Auto Scaling Group"
  value       = aws_autoscaling_group.api.name
}

output "autoscaling_min_size" {
  description = "Minimum size of Auto Scaling Group"
  value       = aws_autoscaling_group.api.min_size
}

output "autoscaling_max_size" {
  description = "Maximum size of Auto Scaling Group"
  value       = aws_autoscaling_group.api.max_size
}

# Frontend (S3 + CloudFront)
output "frontend_s3_website_url" {
  description = "S3 static website URL"
  value       = "http://${aws_s3_bucket_website_configuration.frontend.website_endpoint}"
}

output "frontend_cloudfront_url" {
  description = "CloudFront distribution URL for frontend"
  value       = "https://${aws_cloudfront_distribution.frontend.domain_name}"
}

output "frontend_s3_bucket" {
  description = "S3 bucket name for frontend"
  value       = aws_s3_bucket.frontend.id
}

# Worker Instance
output "worker_public_ip" {
  description = "Public IP address of the worker"
  value       = aws_instance.worker.public_ip
}

output "worker_private_ip" {
  description = "Private IP address of the worker"
  value       = aws_instance.worker.private_ip
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

# S3 Bucket (Videos)
output "s3_videos_bucket_name" {
  description = "S3 bucket name for video storage"
  value       = aws_s3_bucket.videos.id
}

output "s3_videos_bucket_arn" {
  description = "S3 bucket ARN for videos"
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

# Security Groups
output "alb_security_group_id" {
  description = "Security group ID for ALB"
  value       = aws_security_group.alb.id
}

output "api_instances_security_group_id" {
  description = "Security group ID for API instances"
  value       = aws_security_group.api_instances.id
}

output "worker_security_group_id" {
  description = "Security group ID for worker"
  value       = aws_security_group.worker.id
}

output "rds_security_group_id" {
  description = "Security group ID for RDS"
  value       = aws_security_group.rds.id
}

# IAM Roles
output "api_iam_role_arn" {
  description = "IAM role ARN for API instances"
  value       = aws_iam_role.web_server.arn
}

output "worker_iam_role_arn" {
  description = "IAM role ARN for worker"
  value       = aws_iam_role.worker.arn
}

# CloudWatch
output "cloudwatch_dashboard_url" {
  description = "URL to CloudWatch Dashboard"
  value       = "https://console.aws.amazon.com/cloudwatch/home?region=${var.aws_region}#dashboards:name=${aws_cloudwatch_dashboard.main.dashboard_name}"
}

# Region
output "aws_region" {
  description = "AWS region"
  value       = var.aws_region
}

# SSH Commands
output "ssh_worker" {
  description = "SSH command for worker"
  value       = "ssh -i ~/.ssh/${var.key_pair_name}.pem ec2-user@${aws_instance.worker.public_ip}"
}

output "ssh_api_instances_note" {
  description = "Note about SSH to API instances"
  value       = "API instances are managed by Auto Scaling. Use AWS Systems Manager Session Manager or get instance IPs from EC2 console."
}

# Deployment Summary
output "deployment_summary" {
  description = "Summary of deployed infrastructure"
  value = <<-EOT
    
    ============================================
    🎉 Deployment Complete!
    ============================================
    
    📍 Region: ${var.aws_region}
    🏷️  Environment: ${var.environment}
    
    🌐 Frontend
    ├─ CloudFront: ${aws_cloudfront_distribution.frontend.domain_name}
    ├─ S3 Website: ${aws_s3_bucket_website_configuration.frontend.website_endpoint}
    └─ S3 Bucket:  ${aws_s3_bucket.frontend.id}
    
    🔄 Load Balancer & Auto Scaling
    ├─ ALB DNS:    ${aws_lb.api.dns_name}
    ├─ API URL:    http://${aws_lb.api.dns_name}
    ├─ Min Size:   ${aws_autoscaling_group.api.min_size} instances
    ├─ Max Size:   ${aws_autoscaling_group.api.max_size} instances
    └─ Desired:    ${aws_autoscaling_group.api.desired_capacity} instances
    
    ⚙️  Worker
    ├─ Public IP:  ${aws_instance.worker.public_ip}
    ├─ Private IP: ${aws_instance.worker.private_ip}
    ├─ Instance:   ${aws_instance.worker.instance_type}
    └─ SSH:        ssh -i ~/.ssh/${var.key_pair_name}.pem ec2-user@${aws_instance.worker.public_ip}
    
    🗄️  Database (RDS)
    ├─ Endpoint:   ${aws_db_instance.main.endpoint}
    ├─ Address:    ${aws_db_instance.main.address}
    ├─ Port:       ${aws_db_instance.main.port}
    ├─ Database:   ${var.db_name}
    └─ Username:   ${var.db_username}
    
    📦 S3 Buckets
    ├─ Videos:     ${aws_s3_bucket.videos.id}
    └─ Frontend:   ${aws_s3_bucket.frontend.id}
    
    📨 SQS Queue
    ├─ Name:       ${aws_sqs_queue.video_processing.name}
    └─ URL:        ${aws_sqs_queue.video_processing.url}
    
    🐳 ECR Repositories
    ├─ API:        ${aws_ecr_repository.api.repository_url}
    └─ Worker:     ${aws_ecr_repository.worker.repository_url}
    
    📊 Monitoring
    └─ Dashboard:  https://console.aws.amazon.com/cloudwatch/home?region=${var.aws_region}#dashboards:name=${aws_cloudwatch_dashboard.main.dashboard_name}
    
    🌐 Application URLs
    ├─ Frontend:   https://${aws_cloudfront_distribution.frontend.domain_name}
    └─ API:        http://${aws_lb.api.dns_name}
    
    ============================================
    📋 Next Steps:
    ============================================
    
    1. Initialize database:
       cd .. && ./terraform/scripts/init-db.sh
    
    2. Build and push Docker images:
       cd .. && ./terraform/scripts/push-images.sh
    
    3. Deploy frontend to S3:
       aws s3 sync ./frontend/build s3://${aws_s3_bucket.frontend.id} --delete
       aws cloudfront create-invalidation --distribution-id ${aws_cloudfront_distribution.frontend.id} --paths "/*"
    
    4. Verify deployment:
       curl http://${aws_lb.api.dns_name}/health
       curl https://${aws_cloudfront_distribution.frontend.domain_name}
    
    ============================================
  EOT
}