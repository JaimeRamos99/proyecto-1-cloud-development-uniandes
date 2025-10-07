# Terraform Quick Start Guide

Get your infrastructure running in AWS in **15 minutes**.

## ⚡ Prerequisites (5 minutes)

1. **Install Terraform**:

   ```bash
   # macOS
   brew install terraform

   # Linux
   wget https://releases.hashicorp.com/terraform/1.6.0/terraform_1.6.0_linux_amd64.zip
   unzip terraform_1.6.0_linux_amd64.zip
   sudo mv terraform /usr/local/bin/
   ```

2. **Configure AWS CLI**:

   ```bash
   aws configure
   # Enter: Access Key, Secret Key, Region (us-east-1), Output (json)
   ```

3. **Create SSH Key Pair** (if you don't have one):
   ```bash
   # In AWS Console: EC2 → Key Pairs → Create Key Pair
   # Save the .pem file to ~/.ssh/
   chmod 400 ~/.ssh/your-key.pem
   ```

## 🚀 Deploy Infrastructure (10 minutes)

### Step 1: Configure Variables (2 minutes)

```bash
cd terraform

# Copy example file
cp terraform.tfvars.example terraform.tfvars

# Edit with your values
nano terraform.tfvars
```

**Minimum required changes:**

```hcl
key_pair_name  = "your-key-pair-name"        # Your AWS key pair
db_password    = "SecurePassword123!"        # Strong password
s3_bucket_name = "proyecto1-videos-12345"    # Globally unique
jwt_secret     = "your-random-secret-here"   # Generate with: openssl rand -base64 32
```

**Optional but recommended:**

```hcl
allowed_ssh_cidr = "YOUR_IP/32"  # Restrict SSH to your IP
```

### Step 2: Initialize Terraform (1 minute)

```bash
terraform init
```

### Step 3: Preview Changes (1 minute)

```bash
terraform plan
```

You should see: `Plan: 30 to add, 0 to change, 0 to destroy`

### Step 4: Create Infrastructure (6 minutes)

```bash
terraform apply
```

Type `yes` when prompted.

**Wait time**: ~6 minutes (RDS takes the longest)

### Step 5: Save Outputs

```bash
terraform output > deployment-info.txt
terraform output deployment_summary
```

## 📦 Post-Deployment (10 minutes)

### 1. Initialize Database (2 minutes)

```bash
export RDS_ENDPOINT=$(terraform output -raw rds_address)
export DB_PASSWORD="your-db-password"

cd ..
PGPASSWORD=$DB_PASSWORD psql -h $RDS_ENDPOINT -U postgres -d proyecto_1 -f db/001_create_user_table.sql
PGPASSWORD=$DB_PASSWORD psql -h $RDS_ENDPOINT -U postgres -d proyecto_1 -f db/002_create_video_table.sql
PGPASSWORD=$DB_PASSWORD psql -h $RDS_ENDPOINT -U postgres -d proyecto_1 -f db/003_add_is_public_to_videos.sql
PGPASSWORD=$DB_PASSWORD psql -h $RDS_ENDPOINT -U postgres -d proyecto_1 -f db/004_create_votes_table.sql
PGPASSWORD=$DB_PASSWORD psql -h $RDS_ENDPOINT -U postgres -d proyecto_1 -f db/005_create_player_rankings_view.sql
```

### 2. Build and Push Docker Images (5 minutes)

```bash
cd terraform
export ECR_REGISTRY=$(terraform output -raw ecr_api_repository_url | cut -d'/' -f1)
export AWS_REGION="us-east-1"

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

### 3. Deploy Application (3 minutes)

```bash
cd ../terraform
export WEB_IP=$(terraform output -raw web_server_public_ip)
export KEY_PAIR_NAME="your-key-pair-name"

# Build frontend
cd ../front
npm ci
npm run build

# Copy files to web server
cd ..
scp -i ~/.ssh/$KEY_PAIR_NAME.pem nginx/nginx.conf ec2-user@$WEB_IP:~/proyecto1/nginx.conf
scp -i ~/.ssh/$KEY_PAIR_NAME.pem -r front/dist/* ec2-user@$WEB_IP:~/proyecto1/frontend-dist/

# Start containers
ssh -i ~/.ssh/$KEY_PAIR_NAME.pem ec2-user@$WEB_IP << 'ENDSSH'
cd ~/proyecto1
docker-compose pull
docker-compose up -d
ENDSSH
```

## ✅ Verify Deployment (1 minute)

```bash
export WEB_IP=$(terraform output -raw web_server_public_ip)

# Test health
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
```

**Success!** Your application is live at: `http://$WEB_IP`

## 🎯 Common Commands

```bash
# View outputs
terraform output
terraform output web_server_public_ip

# SSH to instances
ssh -i ~/.ssh/$KEY_PAIR_NAME.pem ec2-user@$(terraform output -raw web_server_public_ip)

# View logs
ssh -i ~/.ssh/$KEY_PAIR_NAME.pem ec2-user@$(terraform output -raw web_server_public_ip) \
  "docker logs -f proyecto1-api-aws"

# Update infrastructure
terraform plan   # Preview changes
terraform apply  # Apply changes

# Destroy everything
terraform destroy
```

## 💰 Cost

**~$102/month** for:

- EC2 Web Server (t3.small): $15
- EC2 Worker (c5.large): $60
- RDS (db.t3.micro): $15
- S3 + SQS + Data Transfer: $12

## 🆘 Troubleshooting

### "Error: Invalid key pair"

```bash
# Check your key pair exists
aws ec2 describe-key-pairs --key-names your-key-pair-name
```

### "Error: Bucket already exists"

```bash
# Change s3_bucket_name in terraform.tfvars to a unique value
s3_bucket_name = "proyecto1-videos-YOUR_UNIQUE_SUFFIX"
```

### "Can't connect to RDS"

```bash
# Wait for RDS to be fully available (takes ~5 minutes after terraform apply)
aws rds describe-db-instances --db-instance-identifier proyecto1-db
```

### "Docker images not found"

```bash
# Make sure you pushed images to ECR
aws ecr list-images --repository-name proyecto1-api
```

---

**Total Time**: ~25 minutes from zero to running application

**Questions?** Check the [Full README](README.md) or open an issue.
