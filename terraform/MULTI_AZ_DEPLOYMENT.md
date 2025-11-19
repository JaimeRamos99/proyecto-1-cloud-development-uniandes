# Multi-AZ Deployment Configuration

## Overview

This document describes the changes made to deploy web server instances across multiple Availability Zones (AZs) to increase availability and fault tolerance.

## Changes Made

### 1. **main.tf** - Subnet Selection
- **Added**: New data source `data.aws_subnets.multi_az` that explicitly selects subnets from multiple AZs
- **Purpose**: Ensures instances are distributed across different availability zones
- **Configuration**: Uses the `number_of_azs` variable to control how many AZs to use (default: 2)

```hcl
data "aws_subnets" "multi_az" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }

  filter {
    name   = "availability-zone"
    values = slice(data.aws_availability_zones.available.names, 0, var.number_of_azs)
  }
}
```

### 2. **variables.tf** - Configuration Variables
- **Added**: `number_of_azs` variable with validation (must be between 2 and 3)
- **Updated**: Default ASG configuration for multi-AZ deployment:
  - `web_server_asg_min_size`: Changed from 1 to 2
  - `web_server_asg_desired_capacity`: Changed from 1 to 2
  - `web_server_asg_max_size`: Changed from 3 to 4

### 3. **web-server.tf** - Auto Scaling Group & Load Balancer
- **Updated**: Application Load Balancer to use multi-AZ subnets
- **Updated**: Auto Scaling Group to deploy instances across multiple AZs
- **Added**: Comments indicating multi-AZ deployment

### 4. **lambda-worker.tf** - Lambda Function
- **Updated**: Lambda VPC configuration to use multi-AZ subnets
- **Purpose**: Ensures Lambda functions can access resources in all AZs

### 5. **outputs.tf** - Infrastructure Outputs
- **Added**: Output for availability zones being used
- **Added**: Output for subnet IDs across multiple AZs
- **Updated**: Deployment summary to include multi-AZ information

## Benefits

### High Availability
- **Fault Tolerance**: If one AZ fails, instances in other AZs continue to serve traffic
- **Zero Downtime**: The load balancer automatically routes traffic away from failed instances
- **Automatic Recovery**: Auto Scaling Group launches new instances in healthy AZs

### Improved Performance
- **Lower Latency**: Users are served from the nearest AZ
- **Better Distribution**: Traffic is spread across multiple data centers
- **Increased Capacity**: Resources are distributed across multiple physical locations

### Compliance & Best Practices
- **AWS Recommended**: Multi-AZ is an AWS Well-Architected Framework best practice
- **Production Ready**: Suitable for production workloads requiring high availability
- **Disaster Recovery**: Protects against AZ-level failures

## Configuration

### Default Configuration
By default, the infrastructure will deploy across **2 Availability Zones** with:
- **Minimum Instances**: 2 (one per AZ)
- **Desired Instances**: 2
- **Maximum Instances**: 4

### Customization
You can customize the number of AZs by setting the `number_of_azs` variable in your `terraform.tfvars`:

```hcl
number_of_azs = 2  # Deploy across 2 AZs (default)
# or
number_of_azs = 3  # Deploy across 3 AZs
```

### Scaling Behavior
The Auto Scaling Group will automatically:
- Distribute instances evenly across the configured AZs
- Launch new instances in the AZ with the fewest instances
- Maintain balance across AZs during scale-out and scale-in events

## Deployment Architecture

```
┌─────────────────────────────────────────────────────────┐
│           Application Load Balancer (ALB)               │
│              (spans all Availability Zones)             │
└────────────────┬──────────────┬─────────────────────────┘
                 │              │
        ┌────────┴─────┐   ┌────┴─────────┐
        │   AZ-1a      │   │   AZ-1b      │
        ├──────────────┤   ├──────────────┤
        │ Web Server 1 │   │ Web Server 2 │
        │              │   │              │
        │ - API        │   │ - API        │
        │ - Nginx      │   │ - Nginx      │
        └──────────────┘   └──────────────┘
                 │              │
        └────────┴──────────────┴───────────────────┐
                                                     │
                                        ┌────────────┴──────────┐
                                        │   RDS PostgreSQL      │
                                        │   (can be multi-AZ)   │
                                        └───────────────────────┘
```

## Verification

After deployment, verify multi-AZ configuration:

### 1. Check Terraform Outputs
```bash
cd terraform
terraform output availability_zones
terraform output subnet_ids
```

### 2. Verify in AWS Console
- **EC2 > Auto Scaling Groups**: Check the "Instances" tab to see instances distributed across AZs
- **EC2 > Load Balancers**: Verify the ALB spans multiple AZs
- **CloudWatch**: Monitor metrics across all AZs

### 3. Test Failover
Simulate an AZ failure by terminating an instance:
```bash
# Get instance ID from AWS Console
aws ec2 terminate-instances --instance-ids i-xxxxxxxxx

# Auto Scaling Group will automatically launch a replacement
```

## Cost Considerations

### Multi-AZ vs Single-AZ
- **Increased Costs**: Running instances in multiple AZs increases costs
- **Data Transfer**: Cross-AZ data transfer incurs small charges
- **Worth It**: The reliability and availability benefits typically outweigh the additional costs

### Cost Optimization Tips
1. **Right-Size Instances**: Use appropriate instance types for your workload
2. **Use Spot Instances**: Consider using Spot instances for non-critical workloads
3. **Enable Auto Scaling**: Let ASG scale down during low traffic periods
4. **Monitor Usage**: Use CloudWatch to track resource utilization

## Maintenance

### Rolling Updates
When updating instances:
1. Terraform will create new instances before terminating old ones
2. Load balancer health checks ensure traffic only goes to healthy instances
3. Zero downtime during updates

### Scaling Events
- **Scale Out**: New instances are distributed across AZs
- **Scale In**: Instances are terminated from the AZ with the most instances
- **AZ Rebalancing**: ASG automatically rebalances instances across AZs

## Troubleshooting

### Issue: Uneven Distribution
**Cause**: Auto Scaling may temporarily have uneven distribution
**Solution**: Wait for ASG to automatically rebalance, or trigger manual rebalancing

### Issue: Insufficient Capacity
**Cause**: One AZ may have insufficient capacity for the requested instance type
**Solution**: ASG will automatically launch instances in other AZs

### Issue: High Data Transfer Costs
**Cause**: Excessive cross-AZ traffic
**Solution**: 
- Minimize cross-AZ data transfer
- Use CloudFront for static content
- Cache frequently accessed data

## Additional Recommendations

### Enable RDS Multi-AZ
For complete high availability, consider enabling Multi-AZ for RDS:

```hcl
# In rds.tf
resource "aws_db_instance" "main" {
  # ... other configuration ...
  multi_az = true  # Enable automatic failover
}
```

**Note**: This increases RDS costs but provides automatic failover.

### Monitor AZ Health
Set up CloudWatch alarms to monitor:
- Instance health per AZ
- ALB target health per AZ
- Uneven instance distribution alerts

### Disaster Recovery Plan
1. **Regular Backups**: RDS automated backups are enabled (7-day retention)
2. **AMI Snapshots**: Consider periodic AMI snapshots of web servers
3. **Documented Procedures**: Maintain runbooks for common failure scenarios

## Summary

Your infrastructure is now configured for **high availability** across multiple Availability Zones. The Auto Scaling Group will automatically:
- ✅ Distribute instances across 2+ AZs
- ✅ Replace failed instances
- ✅ Balance load across healthy instances
- ✅ Maintain availability during AZ failures

This configuration provides production-grade reliability and follows AWS best practices.

