#!/bin/bash
# ===============================================
# Proyecto 1 - Web Server Initialization (ASG)
# ===============================================

set -euo pipefail

# Log all output to /var/log/user-data.log
exec > >(tee /var/log/user-data.log | logger -t user-data -s 2>/dev/console) 2>&1

echo "==================================="
echo "Starting Web Server Initialization"
echo "==================================="

# --- Update packages ---
echo "Updating system packages..."
dnf update -y

# --- Install Docker ---
echo "Installing Docker..."
dnf install -y docker
systemctl enable docker
systemctl start docker
usermod -aG docker ec2-user

# --- Install Docker Compose ---
echo "Installing Docker Compose..."
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose

# --- Install AWS CLI v2 ---
echo "Installing AWS CLI..."
curl -s "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip -q awscliv2.zip
./aws/install
rm -rf aws awscliv2.zip

# --- Create app directory ---
mkdir -p /home/ec2-user/proyecto1
cd /home/ec2-user/proyecto1

# --- Create Nginx config ---
cat > nginx.aws.conf << 'EOF'
events {}
http {
  server {
    listen 80;
    location / {
      proxy_pass http://localhost:8080;
    }
  }
}
EOF

# --- Create Docker Compose file ---
cat > docker-compose.yml << EOF
version: '3.8'

services:
  api:
    image: ${ecr_api_url}:${ecr_image_tag}
    container_name: ${project_name}-api
    restart: unless-stopped
    environment:
      - DB_DRIVER=postgres
      - DB_HOST=${db_host}
      - DB_PORT=${db_port}
      - DB_NAME=${db_name}
      - DB_USER=${db_user}
      - DB_PASSWORD=${db_password}
      - DB_SSL_MODE=require
      - DB_MAX_OPEN_CONNS=25
      - DB_MAX_IDLE_CONNS=5
      - PORT=8080
      - HOST=0.0.0.0
      - GIN_MODE=release
      - JWT_SECRET=${jwt_secret}
      - JWT_ISSUER=Proyecto_1
      - JWT_EXPIRATION=24h
      - APP_NAME=Proyecto_1
      - APP_VERSION=1.0.0
      - APP_ENV=production
      - AWS_DEFAULT_REGION=${aws_region}
      - S3_BUCKET_NAME=${s3_bucket_name}
      - SQS_QUEUE_NAME=${sqs_queue_name}
      - SQS_QUEUE_URL=${sqs_queue_url}
      - FRONTEND_URL=${frontend_url}
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    container_name: ${project_name}-nginx
    restart: unless-stopped
    network_mode: host
    volumes:
      - ./nginx.aws.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - api
EOF

# --- Set permissions ---
chown -R ec2-user:ec2-user /home/ec2-user/proyecto1

# --- Login to ECR ---
ECR_REGISTRY=$(echo "${ecr_api_url}" | cut -d'/' -f1)
echo "Logging in to ECR registry: $ECR_REGISTRY"
aws ecr get-login-password --region ${aws_region} | docker login --username AWS --password-stdin $ECR_REGISTRY || echo "ECR login failed (might be temporary)"

# --- Pull API image and start services ---
echo "Pulling latest API image..."
if docker pull ${ecr_api_url}:${ecr_image_tag}; then
  echo "Starting containers..."
  docker-compose up -d
else
  echo "ECR image not found yet; skipping container startup."
fi

echo "==================================="
echo "Web Server Initialization Complete"
echo "==================================="
