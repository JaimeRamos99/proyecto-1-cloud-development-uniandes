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
- ✅ Deploy to EC2 instances
- ✅ Health checks

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
WEB_SERVER_IP                   # Public IP of web server EC2
WORKER_IP                       # Public IP of worker EC2
DB_PASSWORD                     # Database password from terraform.tfvars
```

### SSH Access

```
SSH_PRIVATE_KEY                 # Content of your .pem file (entire file)
```

## Getting Values

After `terraform apply`:

```bash
cd terraform
terraform output ecr_api_repository_url      # Extract registry part
terraform output ecr_worker_repository_url  # Extract registry part
terraform output ecr_frontend_repository_url # Extract registry part
terraform output rds_address               # RDS endpoint
terraform output web_server_public_ip     # Web server IP
terraform output worker_public_ip         # Worker IP
```

For SSH key:

```bash
cat ~/.ssh/proyecto1-key.pem
```

## Manual Deployment

If needed:

```bash
# Build and push images
cd api && docker build -t $ECR_REGISTRY/proyecto1-api:latest .
docker push $ECR_REGISTRY/proyecto1-api:latest

cd ../worker && docker build -t $ECR_REGISTRY/proyecto1-worker:latest .
docker push $ECR_REGISTRY/proyecto1-worker:latest

cd ../front && docker build -t $ECR_REGISTRY/proyecto1-frontend:latest .
docker push $ECR_REGISTRY/proyecto1-frontend:latest

# Deploy to web server
scp -i ~/.ssh/proyecto1-key.pem nginx/nginx.aws.conf ec2-user@$WEB_IP:~/proyecto1/
scp -i ~/.ssh/proyecto1-key.pem -r front/dist/* ec2-user@$WEB_IP:~/proyecto1/frontend-dist/

ssh -i ~/.ssh/proyecto1-key.pem ec2-user@$WEB_IP << 'ENDSSH'
cd ~/proyecto1
docker-compose pull
docker-compose up -d
ENDSSH
```
