#!/bin/bash
# Proyecto_1 - Web Server Initialization Script
# This script runs on first boot to set up the web server

set -e

# Log everything
exec > >(tee /var/log/user-data.log)
exec 2>&1

echo "==================================="
echo "Starting Web Server Initialization"
echo "==================================="

# Update system
echo "Updating system packages..."
dnf update -y

# Install Docker
echo "Installing Docker..."
dnf install -y docker
systemctl start docker
systemctl enable docker
usermod -aG docker ec2-user

# Install Docker Compose
echo "Installing Docker Compose..."
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose

# Install AWS CLI v2
echo "Installing AWS CLI..."
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip -q awscliv2.zip
./aws/install
rm -rf aws awscliv2.zip

# Create project directory
echo "Creating project directory..."
mkdir -p /home/ec2-user/proyecto1
cd /home/ec2-user/proyecto1

# Create docker-compose.yml
echo "Creating Docker Compose configuration..."
cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  api:
    image: ${ecr_api_url}:${ecr_image_tag}
    container_name: ${project_name}-api-aws
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
      - AWS_ENDPOINT_URL=
      - S3_BUCKET_NAME=${s3_bucket_name}
      - SQS_QUEUE_NAME=${sqs_queue_name}
    ports:
      - "8080:8080"
    networks:
      - proyecto1-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    container_name: ${project_name}-nginx-aws
    restart: unless-stopped
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./frontend-dist:/usr/share/nginx/html:ro
    depends_on:
      - api
    networks:
      - proyecto1-network
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost/nginx-health"]
      interval: 30s
      timeout: 10s
      retries: 3

networks:
  proyecto1-network:
    driver: bridge
EOF

# Create placeholder directories
mkdir -p frontend-dist

# Create placeholder nginx config
cat > nginx.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    upstream api {
        server api:8080;
    }

    server {
        listen 80;
        server_name _;

        # Frontend
        location / {
            root /usr/share/nginx/html;
            try_files $uri $uri/ /index.html;
        }

        # API proxy
        location /api/ {
            proxy_pass http://api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Health check
        location /nginx-health {
            access_log off;
            return 200 "OK\n";
            add_header Content-Type text/plain;
        }
    }
}
EOF

# Set proper permissions
chown -R ec2-user:ec2-user /home/ec2-user/proyecto1

# Login to ECR
echo "Logging in to ECR..."
ECR_REGISTRY=$(echo "${ecr_api_url}" | cut -d"/" -f1)
aws ecr get-login-password --region ${aws_region} | docker login --username AWS --password-stdin $ECR_REGISTRY

# Note: Images need to be built and pushed before this will work
# The containers will be started manually after deployment

echo "==================================="
echo "Web Server Initialization Complete"
echo "==================================="
echo "Next steps:"
echo "1. Push Docker images to ECR"
echo "2. Copy nginx config and frontend files"
echo "3. Run: docker-compose up -d"

