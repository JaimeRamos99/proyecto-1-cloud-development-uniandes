# autoscaling-alb.tf - Auto Scaling Group and Application Load Balancer for API

# Application Load Balancer
resource "aws_lb" "api" {
  name               = "${var.project_name}-api-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets           = data.aws_subnets.default.ids

  enable_deletion_protection = false
  enable_http2              = true
  enable_cross_zone_load_balancing = true

  tags = {
    Name        = "${var.project_name}-api-alb"
    Environment = var.environment
  }
}

# ALB Target Group
resource "aws_lb_target_group" "api" {
  name     = "${var.project_name}-api-tg"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = data.aws_vpc.default.id

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 5
    interval            = 30
    path                = "/health"  # Make sure your API has this endpoint
    matcher             = "200"
  }

  deregistration_delay = 30

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 86400
    enabled         = true
  }

  tags = {
    Name = "${var.project_name}-api-tg"
  }
}

# ALB Listener
resource "aws_lb_listener" "api" {
  load_balancer_arn = aws_lb.api.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.api.arn
  }
}

data "aws_subnets" "all" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

data "aws_subnet" "details" {
  for_each = toset(data.aws_subnets.all.ids)
  id       = each.value
}

locals {
  # Exclude AZs that lack t3.small capacity (e.g., us-east-1e)
  valid_subnets = [
    for s in data.aws_subnets.all.ids : s
    if contains(
      ["us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1f"],
      data.aws_subnet.details[s].availability_zone
    )
  ]
}


# Launch Template for Auto Scaling
resource "aws_launch_template" "api" {
  name_prefix   = "${var.project_name}-api-"
  image_id      = data.aws_ami.amazon_linux_2023.id
  instance_type = var.web_server_instance_type
  key_name      = var.key_pair_name

  vpc_security_group_ids = [aws_security_group.api_instances.id]
  iam_instance_profile {
    arn = aws_iam_instance_profile.web_server.arn
  }

  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      volume_size           = var.root_volume_size
      volume_type           = "gp3"
      delete_on_termination = true
      encrypted            = true
    }
  }

  user_data = base64encode(templatefile("${path.module}/user-data-web-server.sh", {
    aws_region          = var.aws_region
    ecr_api_url        = aws_ecr_repository.api.repository_url
    ecr_image_tag      = var.ecr_image_tag
    db_host            = aws_db_instance.main.address
    db_port            = aws_db_instance.main.port
    db_name            = var.db_name
    db_user            = var.db_username
    db_password        = var.db_password
    jwt_secret         = var.jwt_secret
    s3_bucket_name     = aws_s3_bucket.videos.id
    sqs_queue_url      = aws_sqs_queue.video_processing.url
    sqs_queue_name     = var.sqs_queue_name
    project_name       = var.project_name
    frontend_url       = "https://${aws_cloudfront_distribution.frontend.domain_name}"
  }))

  tag_specifications {
    resource_type = "instance"
    tags = {
      Name        = "${var.project_name}-api-asg-instance"
      Environment = var.environment
      Role        = "api-server"
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Auto Scaling Group
resource "aws_autoscaling_group" "api" {
  name               = "${var.project_name}-api-asg"
  min_size           = var.asg_min_size
  max_size           = var.asg_max_size
  desired_capacity   = var.asg_desired_capacity
  health_check_grace_period = 300
  health_check_type  = "ELB"
  force_delete       = true
  
  launch_template {
    id      = aws_launch_template.api.id
    version = "$Latest"
  }

  vpc_zone_identifier = local.valid_subnets
  target_group_arns   = [aws_lb_target_group.api.arn]

  enabled_metrics = [
    "GroupMinSize",
    "GroupMaxSize",
    "GroupDesiredCapacity",
    "GroupInServiceInstances",
    "GroupPendingInstances",
    "GroupStandbyInstances",
    "GroupTerminatingInstances",
    "GroupTotalInstances"
  ]

  tag {
    key                 = "Name"
    value               = "${var.project_name}-api-asg"
    propagate_at_launch = false
  }

  depends_on = [aws_db_instance.main]
}

# Auto Scaling Policy - Scale Up
resource "aws_autoscaling_policy" "scale_up" {
  name                   = "${var.project_name}-scale-up"
  scaling_adjustment     = 1
  adjustment_type        = "ChangeInCapacity"
  cooldown              = 300
  autoscaling_group_name = aws_autoscaling_group.api.name
}

# Auto Scaling Policy - Scale Down
resource "aws_autoscaling_policy" "scale_down" {
  name                   = "${var.project_name}-scale-down"
  scaling_adjustment     = -1
  adjustment_type        = "ChangeInCapacity"
  cooldown              = 300
  autoscaling_group_name = aws_autoscaling_group.api.name
}

# CloudWatch Alarm - High CPU
resource "aws_cloudwatch_metric_alarm" "cpu_high" {
  alarm_name          = "${var.project_name}-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name        = "CPUUtilization"
  namespace          = "AWS/EC2"
  period             = 300
  statistic          = "Average"
  threshold          = var.cpu_scale_up_threshold
  alarm_description  = "This metric monitors ec2 cpu utilization"
  alarm_actions      = [aws_autoscaling_policy.scale_up.arn]

  dimensions = {
    AutoScalingGroupName = aws_autoscaling_group.api.name
  }
}

# CloudWatch Alarm - Low CPU
resource "aws_cloudwatch_metric_alarm" "cpu_low" {
  alarm_name          = "${var.project_name}-cpu-low"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 2
  metric_name        = "CPUUtilization"
  namespace          = "AWS/EC2"
  period             = 300
  statistic          = "Average"
  threshold          = var.cpu_scale_down_threshold
  alarm_description  = "This metric monitors ec2 cpu utilization"
  alarm_actions      = [aws_autoscaling_policy.scale_down.arn]

  dimensions = {
    AutoScalingGroupName = aws_autoscaling_group.api.name
  }
}

# Target Tracking Scaling Policy (Alternative/Additional)
resource "aws_autoscaling_policy" "target_tracking" {
  name                   = "${var.project_name}-target-tracking"
  autoscaling_group_name = aws_autoscaling_group.api.name
  policy_type           = "TargetTrackingScaling"

  target_tracking_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ASGAverageCPUUtilization"
    }
    target_value = var.cpu_target_value
  }
}

# Security Group for ALB
resource "aws_security_group" "alb" {
  name        = "${var.project_name}-alb-sg"
  description = "Security group for Application Load Balancer"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTP from anywhere"
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS from anywhere"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound traffic"
  }

  tags = {
    Name = "${var.project_name}-alb-sg"
  }
}

# Security Group for API Instances
resource "aws_security_group" "api_instances" {
  name        = "${var.project_name}-api-instances-sg"
  description = "Security group for API instances in Auto Scaling Group"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
    description     = "API port from ALB"
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.allowed_ssh_cidr]
    description = "SSH access"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound traffic"
  }

  tags = {
    Name = "${var.project_name}-api-instances-sg"
  }
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}
