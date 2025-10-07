# Proyecto_1 - Terraform Infrastructure

Complete AWS infrastructure as code using Terraform.

## 📋 Prerequisites

- ✅ [Terraform](https://www.terraform.io/downloads) >= 1.0 installed
- ✅ [AWS CLI](https://aws.amazon.com/cli/) installed and configured
- ✅ AWS account with appropriate permissions
- ✅ SSH key pair created in AWS Console
- ✅ Docker installed locally (for building images)

## 🏗️ Infrastructure Overview

This Terraform configuration creates:

- **2 EC2 Instances**:
  - `web-server` (t3.small): API + Nginx + Frontend
  - `worker` (c5.large): Video processing
- **RDS PostgreSQL** (db.t3.micro): Database
- **S3 Bucket**: Video storage
- **SQS Queue**: Message queue for video processing
- **ECR Repositories**: Docker image storage
- **IAM Roles**: EC2 permissions for AWS services
- **Security Groups**: Network security
- **Elastic IP**: Static IP for web server

## 🚀 Quick Start

### Step 1: Configure Variables

```bash
cd terraform

# Copy example variables
cp terraform.tfvars.example terraform.tfvars

# Edit with your values
nano terraform.tfvars
```

**Required variables to change:**

- `key_pair_name`: Your AWS SSH key pair name
- `db_password`: Strong database password
- `s3_bucket_name`: Globally unique bucket name
- `jwt_secret`: Random secret (generate with `openssl rand -base64 32`)
- `allowed_ssh_cidr`: Your IP address (e.g., "1.2.3.4/32")

### Step 2: Initialize Terraform

```bash
terraform init
```

This downloads the AWS provider and initializes the backend.

### Step 3: Plan Infrastructure

```bash
terraform plan
```

Review what will be created. You should see ~30 resources to be created.

### Step 4: Apply Infrastructure

```bash
terraform apply
```

Type `yes` when prompted. This will take ~10-15 minutes (RDS takes the longest).

### Step 5: Save Outputs

```bash
terraform output > ../deployment-info.txt
```

This saves important information like IP addresses and endpoints.

## 📦 Post-Deployment Steps

After Terraform completes, you need to:

### 1. Initialize Database Schema

```bash
# Get RDS endpoint from outputs
export RDS_ENDPOINT=$(terraform output -raw rds_address)
export DB_PASSWORD="your-db-password"

# Run migrations
cd ..
psql -h $RDS_ENDPOINT -U postgres -d proyecto_1 -f db/001_create_user_table.sql
psql -h $RDS_ENDPOINT -U postgres -d proyecto_1 -f db/002_create_video_table.sql
psql -h $RDS_ENDPOINT -U postgres -d proyecto_1 -f db/003_add_is_public_to_videos.sql
psql -h $RDS_ENDPOINT -U postgres -d proyecto_1 -f db/004_create_votes_table.sql
psql -h $RDS_ENDPOINT -U postgres -d proyecto_1 -f db/005_create_player_rankings_view.sql
```

### 2. Build and Push Docker Images

```bash
# Get ECR registry URL
export ECR_REGISTRY=$(terraform output -raw ecr_api_repository_url | cut -d'/' -f1)
export AWS_REGION=$(terraform output -raw aws_region || echo "us-east-1")

# Login to ECR
aws ecr get-login-password --region $AWS_REGION | \
  docker login --username AWS --password-stdin $ECR_REGISTRY

# Build and push API
cd ../api
docker build -t $ECR_REGISTRY/proyecto1-api:latest .
docker push $ECR_REGISTRY/proyecto1-api:latest

# Build and push Worker
cd ../worker
docker build -t $ECR_REGISTRY/proyecto1-worker:latest .
docker push $ECR_REGISTRY/proyecto1-worker:latest
```

### 3. Build Frontend

```bash
cd ../front
npm ci
npm run build
```

### 4. Deploy to Web Server

```bash
cd ../terraform

# Get web server IP
export WEB_IP=$(terraform output -raw web_server_public_ip)
export KEY_PAIR_NAME="your-key-pair-name"

# Wait for EC2 to finish initialization (check user-data logs)
ssh -i ~/.ssh/$KEY_PAIR_NAME.pem ec2-user@$WEB_IP "tail -f /var/log/cloud-init-output.log"
# Press Ctrl+C when you see "Cloud-init finished"

# Copy nginx config
scp -i ~/.ssh/$KEY_PAIR_NAME.pem ../nginx/nginx.conf ec2-user@$WEB_IP:~/proyecto1/nginx.conf

# Copy frontend files
scp -i ~/.ssh/$KEY_PAIR_NAME.pem -r ../front/dist/* ec2-user@$WEB_IP:~/proyecto1/frontend-dist/

# Start containers
ssh -i ~/.ssh/$KEY_PAIR_NAME.pem ec2-user@$WEB_IP << 'ENDSSH'
cd ~/proyecto1
docker-compose pull
docker-compose up -d
docker ps
ENDSSH
```

### 5. Verify Deployment

```bash
export WEB_IP=$(terraform output -raw web_server_public_ip)

# Test health endpoints
curl http://$WEB_IP/nginx-health
curl http://$WEB_IP/api/health

# Test signup
curl -X POST http://$WEB_IP/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!",
    "username": "testuser",
    "city": "Bogota"
  }'

echo ""
echo "🎉 Application live at: http://$WEB_IP"
```

## 🔄 Managing Infrastructure

### View Current State

```bash
terraform show
```

### Update Infrastructure

```bash
# Edit terraform.tfvars or .tf files
nano terraform.tfvars

# Preview changes
terraform plan

# Apply changes
terraform apply
```

### View Outputs Again

```bash
terraform output
terraform output web_server_public_ip
terraform output deployment_summary
```

### SSH into Instances

```bash
# Web server
terraform output -raw ssh_web_server | bash

# Worker
terraform output -raw ssh_worker | bash
```

### View Logs

```bash
# Web server logs
ssh -i ~/.ssh/$KEY_PAIR_NAME.pem ec2-user@$(terraform output -raw web_server_public_ip) \
  "docker logs -f proyecto1-api-aws"

# Worker logs
ssh -i ~/.ssh/$KEY_PAIR_NAME.pem ec2-user@$(terraform output -raw worker_public_ip) \
  "docker logs -f proyecto1-worker-aws"
```

## 🧹 Cleanup

To destroy all infrastructure:

```bash
terraform destroy
```

Type `yes` when prompted. This will delete all resources created by Terraform.

**Warning**: This is irreversible! Make sure to backup any important data first.

## 💰 Cost Estimate

| Resource       | Configuration | Monthly Cost    |
| -------------- | ------------- | --------------- |
| EC2 Web Server | t3.small      | ~$15            |
| EC2 Worker     | c5.large      | ~$60            |
| RDS PostgreSQL | db.t3.micro   | ~$15            |
| S3 Storage     | 100GB         | ~$2             |
| SQS            | 1M requests   | ~$0.40          |
| Data Transfer  | 100GB         | ~$9             |
| ECR Storage    | 10GB          | ~$1             |
| **Total**      |               | **~$102/month** |

## 📊 Instance Types

### Web Server: t3.small

- **vCPU**: 2
- **RAM**: 2 GiB
- **Storage**: 30 GiB (configurable)
- **Cost**: ~$15/month
- **Use case**: API + Nginx (general purpose)

### Worker: c5.large

- **vCPU**: 2
- **RAM**: 4 GiB
- **Storage**: 30 GiB (configurable)
- **Cost**: ~$60/month
- **Use case**: Video processing (compute optimized)

You can change instance types in `terraform.tfvars`:

```hcl
web_server_instance_type = "t3.micro"   # Cheaper: 2 vCPU, 1 GiB RAM
worker_instance_type     = "t3.medium"  # Cheaper: 2 vCPU, 4 GiB RAM
```

## 🔐 Security Best Practices

1. **Restrict SSH access**:

   ```hcl
   allowed_ssh_cidr = "YOUR_IP/32"  # Not 0.0.0.0/0
   ```

2. **Use strong passwords**:

   ```bash
   # Generate secure password
   openssl rand -base64 32
   ```

3. **Enable MFA** on your AWS account

4. **Use IAM roles** instead of access keys (already configured)

5. **Enable RDS encryption** (already enabled)

6. **Regular backups** (RDS backups enabled for 7 days)

## 🔧 Troubleshooting

### Terraform init fails

```bash
# Clear cache and reinitialize
rm -rf .terraform .terraform.lock.hcl
terraform init
```

### Apply fails with "already exists"

```bash
# Import existing resource
terraform import aws_instance.web_server i-1234567890abcdef0

# Or destroy and recreate
terraform destroy -target=aws_instance.web_server
terraform apply
```

### Can't SSH to instances

```bash
# Check security group
terraform output web_server_public_ip
aws ec2 describe-security-groups --group-ids $(terraform output -raw web_server_security_group_id)

# Verify key permissions
chmod 400 ~/.ssh/$KEY_PAIR_NAME.pem
```

### RDS connection fails

```bash
# Check RDS is available
aws rds describe-db-instances --db-instance-identifier proyecto1-db

# Test connection
psql -h $(terraform output -raw rds_address) -U postgres -d proyecto_1
```

## 📚 Terraform Commands Cheat Sheet

```bash
# Initialize
terraform init

# Validate configuration
terraform validate

# Format code
terraform fmt -recursive

# Plan changes
terraform plan
terraform plan -out=plan.tfplan

# Apply changes
terraform apply
terraform apply plan.tfplan
terraform apply -auto-approve

# Destroy resources
terraform destroy
terraform destroy -target=aws_instance.worker

# View state
terraform show
terraform state list

# View outputs
terraform output
terraform output -json

# Refresh state
terraform refresh

# Import existing resource
terraform import aws_instance.web_server i-1234567890abcdef0
```

## 📝 File Structure

```
terraform/
├── main.tf                    # Main configuration & providers
├── variables.tf               # Input variables
├── outputs.tf                 # Output values
├── terraform.tfvars.example   # Example variables
├── security-groups.tf         # Security group definitions
├── iam.tf                     # IAM roles and policies
├── rds.tf                     # RDS database
├── s3.tf                      # S3 bucket
├── sqs.tf                     # SQS queue
├── ecr.tf                     # ECR repositories
├── ec2-web-server.tf          # Web server EC2
├── ec2-worker.tf              # Worker EC2
├── user-data-web-server.sh    # Web server initialization
├── user-data-worker.sh        # Worker initialization
├── .gitignore                 # Git ignore rules
└── README.md                  # This file
```

---

**🎉 Happy Terraforming!**

For questions or issues, refer to the main project README or AWS documentation.
