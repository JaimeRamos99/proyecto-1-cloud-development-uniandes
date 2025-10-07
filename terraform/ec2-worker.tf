# Proyecto_1 - Worker EC2 Instance
# Handles asynchronous video processing

resource "aws_instance" "worker" {
  ami                    = data.aws_ami.amazon_linux_2023.id
  instance_type          = var.worker_instance_type
  key_name              = var.key_pair_name
  vpc_security_group_ids = [aws_security_group.worker.id]
  iam_instance_profile   = aws_iam_instance_profile.worker.name

  root_block_device {
    volume_size           = var.root_volume_size
    volume_type           = "gp3"
    delete_on_termination = true
    encrypted             = true
  }

  user_data = templatefile("${path.module}/user-data-worker.sh", {
    aws_region          = var.aws_region
    ecr_worker_url      = aws_ecr_repository.worker.repository_url
    ecr_image_tag       = var.ecr_image_tag
    db_host             = aws_db_instance.main.address
    db_port             = aws_db_instance.main.port
    db_name             = var.db_name
    db_user             = var.db_username
    db_password         = var.db_password
    s3_bucket_name      = aws_s3_bucket.videos.id
    sqs_queue_url       = aws_sqs_queue.video_processing.url
    sqs_queue_name      = var.sqs_queue_name
    project_name        = var.project_name
  })

  # Wait for RDS and SQS to be available
  depends_on = [
    aws_db_instance.main,
    aws_sqs_queue.video_processing
  ]

  tags = {
    Name = "${var.project_name}-worker"
    Role = "worker"
  }
}

