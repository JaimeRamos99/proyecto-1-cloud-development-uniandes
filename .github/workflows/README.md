# GitHub Actions Workflows

This directory contains CI/CD workflows for automated testing and deployment.

## Workflows

### 1. CI (`pr-ci.yml`)

**Triggers**: Pull requests and pushes to main/develop  
**Purpose**: Code validation and testing

- ✅ Go tests for API and Worker
- ✅ Code formatting checks (gofmt)
- ✅ Static analysis (go vet)
- ✅ Application builds
- ✅ PostgreSQL service for testing

### 2. Deploy (`deploy.yml`)

**Triggers**:

- Pull requests → Tests + frontend build only
- Main branch → Full deployment

**Purpose**: Application deployment

- ✅ Build and push Docker images to ECR
- ✅ Run database migrations
- ✅ Build and deploy frontend
- ✅ Deploy to Auto Scaling Group instances via SSM
- ✅ Health checks via Application Load Balancer

## Required GitHub Secrets

Add these in **Settings → Secrets and variables → Actions → Repository secrets**:

### AWS Credentials

```
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
```

### Infrastructure Details

```
ECR_REGISTRY                    # e.g., 123456789.dkr.ecr.us-east-1.amazonaws.com
RDS_ENDPOINT                    # e.g., proyecto1-db.xxx.us-east-1.rds.amazonaws.com
DB_PASSWORD                     # Database password from terraform.tfvars
S3_BUCKET_NAME                  # S3 bucket name for videos
SQS_QUEUE_NAME                  # SQS queue name for video processing
JWT_SECRET                      # JWT secret for authentication
ALB_DNS_NAME                    # Application Load Balancer DNS name (REQUIRED)
ASG_NAME                        # Auto Scaling Group name (optional, defaults to proyecto1-web-server-asg)
WORKER_IP                       # Public IP of worker EC2 (still uses SSH)
```

### SSH Access (Worker Only)

```
SSH_PRIVATE_KEY                 # Content of your .pem file (entire file) - Only needed for worker
```

**Note**: Web server now uses AWS Systems Manager (SSM) - no SSH keys needed!

## Getting Values

After `terraform apply`:

```bash
cd terraform
# ECR Registry (extract from repository URL)
terraform output ecr_api_repository_url      # Extract registry part (before /)
terraform output ecr_worker_repository_url  # Extract registry part
terraform output ecr_frontend_repository_url # Extract registry part

# Database
terraform output rds_address                 # RDS endpoint

# Application Load Balancer (REQUIRED for health checks)
terraform output alb_dns_name                # Add as ALB_DNS_NAME secret

# Auto Scaling Group (optional)
terraform output web_server_asg_id           # ASG name for ASG_NAME secret

# Worker (still uses SSH)
terraform output worker_public_ip            # Worker IP
```

For SSH key (worker only):

```bash
cat ~/.ssh/proyecto1-key.pem
```

## Important: After Terraform Apply

**You MUST add the ALB DNS name as a GitHub secret:**

1. Get the ALB DNS:
   ```bash
   terraform output -raw alb_dns_name
   ```

2. Add to GitHub Secrets:
   - Go to: Settings → Secrets and variables → Actions → Repository secrets
   - Add: `ALB_DNS_NAME` = (value from step 1)

## Manual Deployment

If needed, you can deploy manually using AWS Systems Manager:

```bash
# Get instance IDs from ASG
INSTANCE_IDS=$(aws autoscaling describe-auto-scaling-groups \
  --auto-scaling-group-names proyecto1-web-server-asg \
  --region us-east-1 \
  --query 'AutoScalingGroups[0].Instances[*].InstanceId' \
  --output text)

# Deploy to each instance
for INSTANCE_ID in $INSTANCE_IDS; do
  aws ssm send-command \
    --instance-ids "$INSTANCE_ID" \
    --document-name "AWS-RunShellScript" \
    --parameters "commands=['cd /home/ec2-user/proyecto1', 'sudo docker-compose pull', 'sudo docker-compose up -d']" \
    --region us-east-1
done
```
