# Proyecto_1 - Terraform Outputs
# Export important information about the infrastructure

# EC2 Instances
output "web_server_public_ip" {
  description = "Public IP address of the web server"
  value       = aws_eip.web_server.public_ip
}

output "web_server_private_ip" {
  description = "Private IP address of the web server"
  value       = aws_instance.web_server.private_ip
}

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

# Security Groups
output "web_server_security_group_id" {
  description = "Security group ID for web server"
  value       = aws_security_group.web_server.id
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
output "web_server_iam_role_arn" {
  description = "IAM role ARN for web server"
  value       = aws_iam_role.web_server.arn
}

output "worker_iam_role_arn" {
  description = "IAM role ARN for worker"
  value       = aws_iam_role.worker.arn
}

# Region
output "aws_region" {
  description = "AWS region"
  value       = var.aws_region
}

# SSH Commands
output "ssh_web_server" {
  description = "SSH command for web server"
  value       = "ssh -i ~/.ssh/${var.key_pair_name}.pem ec2-user@${aws_eip.web_server.public_ip}"
}

output "ssh_worker" {
  description = "SSH command for worker"
  value       = "ssh -i ~/.ssh/${var.key_pair_name}.pem ec2-user@${aws_instance.worker.public_ip}"
}

# Application URL
output "application_url" {
  description = "Application URL"
  value       = "http://${aws_eip.web_server.public_ip}"
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
    
    🌐 Web Server
    ├─ Public IP:  ${aws_eip.web_server.public_ip}
    ├─ Private IP: ${aws_instance.web_server.private_ip}
    ├─ Instance:   ${aws_instance.web_server.instance_type}
    └─ SSH:        ssh -i ~/.ssh/${var.key_pair_name}.pem ec2-user@${aws_eip.web_server.public_ip}
    
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
    
    📦 S3 Bucket
    └─ Name:       ${aws_s3_bucket.videos.id}
    
    📨 SQS Queue
    ├─ Name:       ${aws_sqs_queue.video_processing.name}
    └─ URL:        ${aws_sqs_queue.video_processing.url}
    
    🐳 ECR Repositories
    ├─ API:        ${aws_ecr_repository.api.repository_url}
    └─ Worker:     ${aws_ecr_repository.worker.repository_url}
    
    🌐 Application
    └─ URL:        http://${aws_eip.web_server.public_ip}
    
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
       curl http://${aws_eip.web_server.public_ip}/api/health
    
    ============================================
  EOT
}

