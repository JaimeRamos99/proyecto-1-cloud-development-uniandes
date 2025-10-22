# Proyecto_1 - Web Server EC2 Instance
# Hosts API + Nginx + Frontend

resource "aws_instance" "web_server" {
  ami                    = data.aws_ami.amazon_linux_2023.id
  instance_type          = var.web_server_instance_type
  key_name              = var.key_pair_name
  vpc_security_group_ids = [aws_security_group.web_server.id]
  iam_instance_profile   = aws_iam_instance_profile.web_server.name

  root_block_device {
    volume_size           = var.root_volume_size
    volume_type           = "gp3"
    delete_on_termination = true
    encrypted             = true
  }

  user_data = templatefile("${path.module}/user-data-web-server.sh", {
    aws_region          = var.aws_region
    ecr_api_url         = aws_ecr_repository.api.repository_url
    ecr_image_tag       = var.ecr_image_tag
    db_host             = aws_db_instance.main.address
    db_port             = aws_db_instance.main.port
    db_name             = var.db_name
    db_user             = var.db_username
    db_password         = var.db_password
    jwt_secret          = var.jwt_secret
    s3_bucket_name      = aws_s3_bucket.videos.id
    sqs_queue_url       = aws_sqs_queue.video_processing.url
    sqs_queue_name      = var.sqs_queue_name
    project_name        = var.project_name
    frontend_url        = "https://${aws_cloudfront_distribution.frontend.domain_name}"
  })

  # Wait for RDS to be available before launching
  depends_on = [aws_db_instance.main]

  tags = {
    Name = "${var.project_name}-web-server"
    Role = "web-server"
  }
}

# Elastic IP for web server (static IP)
resource "aws_eip" "web_server" {
  instance = aws_instance.web_server.id
  domain   = "vpc"

  tags = {
    Name = "${var.project_name}-web-server-eip"
  }
}

