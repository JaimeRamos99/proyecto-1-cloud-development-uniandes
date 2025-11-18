# Proyecto_1 - VPC Endpoints
# Gateway endpoints for Lambda to access AWS services without internet
# Note: VPC and subnets data sources are defined in security-groups.tf

# Get route tables for the VPC
data "aws_route_tables" "default" {
  vpc_id = data.aws_vpc.default.id
}

# ============================================================================
# S3 Gateway Endpoint
# ============================================================================
# Allows Lambda in VPC to access S3 without NAT Gateway (free, faster)

resource "aws_vpc_endpoint" "s3" {
  vpc_id       = data.aws_vpc.default.id
  service_name = "com.amazonaws.${var.aws_region}.s3"

  # Gateway endpoint (no cost, better performance than interface endpoint)
  vpc_endpoint_type = "Gateway"

  # Associate with all route tables in the VPC
  route_table_ids = data.aws_route_tables.default.ids

  # Allow Lambda to access S3
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:*"
        Resource  = "*"
      }
    ]
  })

  tags = {
    Name = "${var.project_name}-s3-endpoint"
    Type = "Gateway"
  }
}

# ============================================================================
# SQS Interface Endpoint (Optional - for better performance)
# ============================================================================
# Allows Lambda to access SQS through private network
# Note: Interface endpoints have a cost (~$0.01/hour = $7/month)

# Uncomment if you want SQS to go through VPC endpoint instead of internet
# resource "aws_vpc_endpoint" "sqs" {
#   vpc_id            = data.aws_vpc.default.id
#   service_name      = "com.amazonaws.${var.aws_region}.sqs"
#   vpc_endpoint_type = "Interface"
#
#   subnet_ids         = data.aws_subnets.default.ids
#   security_group_ids = [aws_security_group.lambda_worker.id]
#
#   private_dns_enabled = true
#
#   tags = {
#     Name = "${var.project_name}-sqs-endpoint"
#     Type = "Interface"
#   }
# }

# ============================================================================
# Outputs
# ============================================================================

output "s3_vpc_endpoint_id" {
  description = "ID of the S3 VPC endpoint"
  value       = aws_vpc_endpoint.s3.id
}

output "s3_vpc_endpoint_state" {
  description = "State of the S3 VPC endpoint"
  value       = aws_vpc_endpoint.s3.state
}

