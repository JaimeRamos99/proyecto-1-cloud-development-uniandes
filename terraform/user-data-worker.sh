#!/bin/bash
# Proyecto_1 - Worker Initialization Script
# This script runs on first boot to set up the worker

set -e

# Log everything
exec > >(tee /var/log/user-data.log)
exec 2>&1

echo "==================================="
echo "Starting Worker Initialization"
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

# Install AWS CLI v2
echo "Installing AWS CLI..."
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip -q awscliv2.zip
./aws/install
rm -rf aws awscliv2.zip

# Install FFmpeg (required for video processing)
echo "Installing FFmpeg..."
# Try multiple methods to install FFmpeg
if ! dnf install -y ffmpeg; then
    echo "FFmpeg not available in default repos, trying alternative methods..."
    # Install EPEL first
    dnf install -y epel-release || echo "EPEL installation failed, continuing..."
    # Try installing from RPM Fusion
    dnf install -y https://mirrors.rpmfusion.org/free/el/rpmfusion-free-release-8.noarch.rpm || echo "RPM Fusion installation failed, continuing..."
    # Try installing FFmpeg again
    dnf install -y ffmpeg || echo "FFmpeg installation failed, but continuing with deployment..."
fi

# Create project directory
echo "Creating project directory..."
mkdir -p /home/ec2-user/proyecto1
cd /home/ec2-user/proyecto1

# Create environment file
cat > .env << EOF
DB_DRIVER=postgres
DB_HOST=${db_host}
DB_PORT=${db_port}
DB_NAME=${db_name}
DB_USER=${db_user}
DB_PASSWORD=${db_password}
DB_SSL_MODE=require
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
AWS_DEFAULT_REGION=${aws_region}
AWS_ENDPOINT_URL=
S3_BUCKET_NAME=${s3_bucket_name}
SQS_QUEUE_NAME=${sqs_queue_name}
EOF

# Set proper permissions
chown -R ec2-user:ec2-user /home/ec2-user/proyecto1

# Login to ECR
echo "Logging in to ECR..."
ECR_REGISTRY=$(echo "${ecr_worker_url}" | cut -d"/" -f1)
aws ecr get-login-password --region ${aws_region} | docker login --username AWS --password-stdin $ECR_REGISTRY || echo "ECR login failed, but continuing..."

# Pull and run worker container
echo "Pulling worker image..."
docker pull ${ecr_worker_url}:${ecr_image_tag} || echo "Image not yet available, will need to push first"

# Create systemd service for worker
cat > /etc/systemd/system/${project_name}-worker.service << EOF
[Unit]
Description=Proyecto 1 Worker
After=docker.service
Requires=docker.service

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/home/ec2-user/proyecto1
ExecStartPre=-/usr/bin/docker stop ${project_name}-worker-aws
ExecStartPre=-/usr/bin/docker rm ${project_name}-worker-aws
ExecStartPre=/usr/bin/docker pull ${ecr_worker_url}:${ecr_image_tag}
ExecStart=/usr/bin/docker run --name ${project_name}-worker-aws \\
  --network host \\
  --env-file /home/ec2-user/proyecto1/.env \\
  ${ecr_worker_url}:${ecr_image_tag}
ExecStop=/usr/bin/docker stop ${project_name}-worker-aws
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Enable and start the service (if image is available)
systemctl daemon-reload || echo "Failed to reload systemd, continuing..."
systemctl enable ${project_name}-worker.service || echo "Failed to enable service, continuing..."

# Try to start the service (will fail if image not pushed yet)
systemctl start ${project_name}-worker.service || echo "Worker service will start once image is pushed"

echo "==================================="
echo "Worker Initialization Complete"
echo "==================================="
echo "Next steps:"
echo "1. Push Docker image to ECR"
echo "2. Restart service: sudo systemctl restart ${project_name}-worker.service"

