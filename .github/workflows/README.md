# GitHub Actions Setup

This document explains how to configure GitHub Actions for automated deployment.

## Required GitHub Secrets

Add these secrets in your GitHub repository:
**Settings → Secrets and variables → Actions → Repository secrets**

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
SSH_PRIVATE_KEY                 # Content of your .pem file (entire file content)
```

## How to get these values

### 1. AWS Credentials
Use the same credentials you configured with `aws configure`

### 2. Infrastructure Details
After running `terraform apply`, get these from terraform outputs:

```bash
cd terraform
terraform output ecr_api_repository_url    # Extract registry part
terraform output rds_address               # RDS endpoint
terraform output web_server_public_ip     # Web server IP
terraform output worker_public_ip         # Worker IP
```

### 3. SSH Private Key
Copy the entire content of your `.pem` file:

```bash
cat ~/.ssh/proyecto1-key.pem
```

Copy everything including:
```
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
...
-----END RSA PRIVATE KEY-----
```

## Workflow Behavior

### On Pull Request:
- ✅ Runs tests (API + Worker)
- ✅ Builds frontend
- ❌ Does NOT deploy

### On Push to Main:
- ✅ Runs tests
- ✅ Builds and pushes Docker images to ECR
- ✅ Runs database migrations
- ✅ Builds frontend
- ✅ Deploys to EC2 instances
- ✅ Runs health checks

## Manual Deployment

If you need to deploy manually:

```bash
# Build and push images
cd api && docker build -t $ECR_REGISTRY/proyecto1-api:latest .
docker push $ECR_REGISTRY/proyecto1-api:latest

cd ../worker && docker build -t $ECR_REGISTRY/proyecto1-worker:latest .
docker push $ECR_REGISTRY/proyecto1-worker:latest

# Deploy to web server
scp -i ~/.ssh/proyecto1-key.pem nginx/nginx.conf ec2-user@$WEB_IP:~/proyecto1/
scp -i ~/.ssh/proyecto1-key.pem -r front/dist/* ec2-user@$WEB_IP:~/proyecto1/frontend-dist/

ssh -i ~/.ssh/proyecto1-key.pem ec2-user@$WEB_IP << 'ENDSSH'
cd ~/proyecto1
docker-compose pull
docker-compose up -d
ENDSSH
```

## Troubleshooting

### Common Issues:

1. **SSH connection fails**: Check SSH_PRIVATE_KEY secret format
2. **ECR push fails**: Verify AWS credentials and ECR_REGISTRY
3. **Database migration fails**: Check RDS_ENDPOINT and DB_PASSWORD
4. **Health check fails**: Wait longer or check container logs

### View logs:
```bash
ssh -i ~/.ssh/proyecto1-key.pem ec2-user@$WEB_IP "docker logs -f proyecto1-api-aws"
```
