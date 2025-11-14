# Proyecto_1 - Web Server Launch Template
# Launch template for Auto Scaling Group (API + Nginx + Frontend)

resource "aws_launch_template" "web_server" {
  name_prefix   = "${var.project_name}-web-server-"
  image_id      = data.aws_ami.amazon_linux_2023.id
  instance_type = var.web_server_instance_type
  key_name      = var.key_pair_name

  vpc_security_group_ids = [aws_security_group.web_server.id]

  iam_instance_profile {
    name = aws_iam_instance_profile.web_server.name
  }

  block_device_mappings {
    device_name = "/dev/xvda"

    ebs {
      volume_size           = var.root_volume_size
      volume_type           = "gp3"
      delete_on_termination = true
      encrypted             = true
    }
  }

  user_data = base64encode(templatefile("${path.module}/user-data-web-server.sh", {
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
  }))

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = "${var.project_name}-web-server"
      Role = "web-server"
    }
  }

  # Wait for RDS to be available before launching
  depends_on = [aws_db_instance.main]

  lifecycle {
    create_before_destroy = true
  }
}

# Application Load Balancer
resource "aws_lb" "web_server" {
  name               = "${var.project_name}-web-server-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = data.aws_subnets.default.ids

  enable_deletion_protection = false

  tags = {
    Name = "${var.project_name}-web-server-alb"
  }
}

# Target Group for web server instances
resource "aws_lb_target_group" "web_server" {
  name     = "${var.project_name}-web-server-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = data.aws_vpc.default.id

  # Graceful shutdown - wait for connections to drain
  deregistration_delay = 30

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 5
    interval            = 30
    path                = "/api/health"
    protocol            = "HTTP"
    port                = "traffic-port"
    matcher             = "200"
  }

  tags = {
    Name = "${var.project_name}-web-server-tg"
  }
}

# ALB Listener
resource "aws_lb_listener" "web_server" {
  load_balancer_arn = aws_lb.web_server.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.web_server.arn
  }
}

# Auto Scaling Group
resource "aws_autoscaling_group" "web_server" {
  name                = "${var.project_name}-web-server-asg"
  vpc_zone_identifier = data.aws_subnets.default.ids
  target_group_arns   = [aws_lb_target_group.web_server.arn]
  health_check_type   = "ELB"
  health_check_grace_period = 300

  min_size         = var.web_server_asg_min_size
  max_size         = var.web_server_asg_max_size
  desired_capacity = var.web_server_asg_desired_capacity

  launch_template {
    id      = aws_launch_template.web_server.id
    version = "$Latest"
  }

  tag {
    key                 = "Name"
    value               = "${var.project_name}-web-server"
    propagate_at_launch = true
  }

  tag {
    key                 = "Role"
    value               = "web-server"
    propagate_at_launch = true
  }

  depends_on = [
    aws_db_instance.main,
    aws_lb_target_group.web_server
  ]
}

# Automatic Scaling Policy - CPU Utilization
resource "aws_autoscaling_policy" "web_server_cpu" {
  name                   = "${var.project_name}-web-server-cpu-scaling"
  autoscaling_group_name = aws_autoscaling_group.web_server.name
  policy_type            = "TargetTrackingScaling"

  target_tracking_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ASGAverageCPUUtilization"
    }
    target_value = 70.0
  }
}
