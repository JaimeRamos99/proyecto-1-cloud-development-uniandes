#!/bin/bash
# Proyecto_1 - Terraform Deployment Helper Script
# This script helps with common deployment tasks

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Proyecto_1 Terraform Deployment${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if terraform is installed
if ! command -v terraform &> /dev/null; then
    print_error "Terraform is not installed!"
    print_info "Install from: https://www.terraform.io/downloads"
    exit 1
fi

# Check if terraform.tfvars exists
if [ ! -f "terraform.tfvars" ]; then
    print_warning "terraform.tfvars not found!"
    print_info "Creating from example..."
    cp terraform.tfvars.example terraform.tfvars
    print_warning "Please edit terraform.tfvars with your values"
    print_info "File location: $(pwd)/terraform.tfvars"
    exit 1
fi

# Menu
echo "Select action:"
echo "  1) Initialize Terraform"
echo "  2) Plan infrastructure"
echo "  3) Apply infrastructure"
echo "  4) Show outputs"
echo "  5) Initialize database schema"
echo "  6) Build and push Docker images"
echo "  7) Deploy to web server"
echo "  8) Deploy to worker"
echo "  9) Full deployment (all steps)"
echo "  10) Destroy infrastructure"
echo "  0) Exit"
echo ""
read -p "Enter your choice [0-10]: " choice

case $choice in
    1)
        print_info "Initializing Terraform..."
        terraform init
        print_success "Terraform initialized!"
        ;;
        
    2)
        print_info "Planning infrastructure..."
        terraform plan
        ;;
        
    3)
        print_info "Applying infrastructure..."
        terraform apply
        print_success "Infrastructure created!"
        print_info "Run 'terraform output' to see details"
        ;;
        
    4)
        print_info "Terraform outputs:"
        terraform output deployment_summary
        ;;
        
    5)
        print_info "Initializing database schema..."
        
        if ! command -v psql &> /dev/null; then
            print_error "psql is not installed!"
            exit 1
        fi
        
        RDS_ENDPOINT=$(terraform output -raw rds_address)
        read -sp "Enter database password: " DB_PASSWORD
        echo ""
        
        print_info "Running migrations..."
        cd ..
        PGPASSWORD=$DB_PASSWORD psql -h $RDS_ENDPOINT -U postgres -d proyecto_1 -f db/001_create_user_table.sql
        PGPASSWORD=$DB_PASSWORD psql -h $RDS_ENDPOINT -U postgres -d proyecto_1 -f db/002_create_video_table.sql
        PGPASSWORD=$DB_PASSWORD psql -h $RDS_ENDPOINT -U postgres -d proyecto_1 -f db/003_add_is_public_to_videos.sql
        PGPASSWORD=$DB_PASSWORD psql -h $RDS_ENDPOINT -U postgres -d proyecto_1 -f db/004_create_votes_table.sql
        PGPASSWORD=$DB_PASSWORD psql -h $RDS_ENDPOINT -U postgres -d proyecto_1 -f db/005_create_player_rankings_view.sql
        cd terraform
        
        print_success "Database initialized!"
        ;;
        
    6)
        print_info "Building and pushing Docker images..."
        
        ECR_API_URL=$(terraform output -raw ecr_api_repository_url)
        ECR_WORKER_URL=$(terraform output -raw ecr_worker_repository_url)
        ECR_REGISTRY=$(echo $ECR_API_URL | cut -d'/' -f1)
        AWS_REGION=$(terraform output -raw aws_region 2>/dev/null || echo "us-east-1")
        
        print_info "Logging in to ECR..."
        aws ecr get-login-password --region $AWS_REGION | \
            docker login --username AWS --password-stdin $ECR_REGISTRY
        
        print_info "Building API image..."
        cd ../api
        docker build -t $ECR_API_URL:latest .
        print_info "Pushing API image..."
        docker push $ECR_API_URL:latest
        
        print_info "Building Worker image..."
        cd ../worker
        docker build -t $ECR_WORKER_URL:latest .
        print_info "Pushing Worker image..."
        docker push $ECR_WORKER_URL:latest
        
        cd ../terraform
        print_success "Docker images pushed!"
        ;;
        
    7)
        print_info "Deploying to web server..."
        
        WEB_IP=$(terraform output -raw web_server_public_ip)
        read -p "Enter SSH key pair name: " KEY_PAIR_NAME
        
        print_info "Building frontend..."
        cd ../front
        npm ci
        npm run build
        
        print_info "Copying files to web server..."
        cd ..
        scp -i ~/.ssh/$KEY_PAIR_NAME.pem nginx/nginx.conf ec2-user@$WEB_IP:~/proyecto1/nginx.conf
        scp -i ~/.ssh/$KEY_PAIR_NAME.pem -r front/dist/* ec2-user@$WEB_IP:~/proyecto1/frontend-dist/
        
        print_info "Starting containers..."
        ssh -i ~/.ssh/$KEY_PAIR_NAME.pem ec2-user@$WEB_IP << 'ENDSSH'
cd ~/proyecto1
docker-compose pull
docker-compose up -d
docker ps
ENDSSH
        
        cd terraform
        print_success "Web server deployed!"
        print_info "Application URL: http://$WEB_IP"
        ;;
        
    8)
        print_info "Worker is automatically deployed by user-data script"
        print_info "To check status:"
        
        WORKER_IP=$(terraform output -raw worker_public_ip)
        read -p "Enter SSH key pair name: " KEY_PAIR_NAME
        
        ssh -i ~/.ssh/$KEY_PAIR_NAME.pem ec2-user@$WORKER_IP "docker ps"
        ;;
        
    9)
        print_info "Starting full deployment..."
        print_warning "This will run all steps sequentially"
        read -p "Continue? (yes/no): " confirm
        
        if [ "$confirm" != "yes" ]; then
            print_info "Cancelled"
            exit 0
        fi
        
        # Run all steps
        $0  # Will show menu for each step
        ;;
        
    10)
        print_warning "This will DESTROY all infrastructure!"
        print_warning "This action is IRREVERSIBLE!"
        read -p "Type 'destroy' to confirm: " confirm
        
        if [ "$confirm" != "destroy" ]; then
            print_info "Cancelled"
            exit 0
        fi
        
        print_info "Destroying infrastructure..."
        terraform destroy
        print_success "Infrastructure destroyed"
        ;;
        
    0)
        print_info "Exiting..."
        exit 0
        ;;
        
    *)
        print_error "Invalid choice"
        exit 1
        ;;
esac

echo ""
print_success "Script completed!"

