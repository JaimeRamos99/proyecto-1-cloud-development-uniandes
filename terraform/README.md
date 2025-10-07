# Infrastructure as Code

This directory contains Terraform configuration for AWS infrastructure.

## What this creates

- **EC2 Instances**: Web server (API + Nginx) and Worker (video processing)
- **RDS PostgreSQL**: Managed database
- **S3 Bucket**: Video file storage
- **SQS Queue**: Message queue for video processing
- **ECR Repositories**: Docker image storage for API and Worker
- **Security Groups**: Network security rules
- **IAM Roles**: AWS permissions for EC2 instances

## Usage

1. **Configure variables**:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars with your values
   ```

2. **Deploy infrastructure**:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

3. **Save outputs**:
   ```bash
   terraform output > ../deployment-info.txt
   ```

## Required Variables

- `key_pair_name`: Your AWS SSH key pair name
- `db_password`: Strong database password
- `s3_bucket_name`: Globally unique bucket name
- `jwt_secret`: Random secret (generate with `openssl rand -base64 32`)
- `allowed_ssh_cidr`: Your IP address (e.g., "1.2.3.4/32")

## Application Deployment

Application deployment is handled by GitHub Actions CI/CD pipeline.
See `.github/workflows/` for automated deployment.

## Cost Estimate

~$102/month for production setup:
- EC2 Web Server (t3.small): ~$15
- EC2 Worker (c5.large): ~$60  
- RDS (db.t3.micro): ~$15
- S3 + SQS + Data Transfer: ~$12

## Cleanup

To destroy all infrastructure:
```bash
terraform destroy
```

**Warning**: This is irreversible!
