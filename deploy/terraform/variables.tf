variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b", "us-east-1c"]
}

variable "private_subnet_cidrs" {
  description = "CIDRs for private subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
}

variable "public_subnet_cidrs" {
  description = "CIDRs for public subnets"
  type        = list(string)
  default     = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
}

variable "eks_cluster_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.28"
}

variable "eks_node_instance_types" {
  description = "EC2 instance types for EKS node group"
  type        = list(string)
  default     = ["t3.medium"]
}

variable "eks_min_nodes" {
  description = "Minimum nodes in EKS node group"
  type        = number
  default     = 3
}

variable "eks_max_nodes" {
  description = "Maximum nodes in EKS node group"
  type        = number
  default     = 10
}

variable "eks_desired_nodes" {
  description = "Desired nodes in EKS node group"
  type        = number
  default     = 3
}

variable "rds_instance_class" {
  description = "RDS instance type"
  type        = string
  default     = "db.t3.medium"
}

variable "rds_allocated_storage" {
  description = "RDS storage in GB"
  type        = number
  default     = 100
}

variable "rds_multi_az" {
  description = "Enable multi-AZ for RDS"
  type        = bool
  default     = true
}

variable "rds_backup_retention_days" {
  description = "RDS backup retention in days"
  type        = number
  default     = 30
}

variable "msk_broker_instance_type" {
  description = "MSK broker instance type"
  type        = string
  default     = "kafka.t3.small"
}

variable "msk_broker_count" {
  description = "Number of MSK brokers"
  type        = number
  default     = 3
}

variable "msk_ebs_volume_size" {
  description = "MSK broker EBS volume size in GB"
  type        = number
  default     = 100
}

variable "domain_name" {
  description = "Domain name for the platform"
  type        = string
  default     = "api.roundup-platform.com"
}

variable "certificate_arn" {
  description = "ACM certificate ARN for the domain"
  type        = string
  default     = ""
}

variable "db_username" {
  description = "Database master username"
  type        = string
  sensitive   = true
  default     = "roundup_admin"
}

variable "db_password" {
  description = "Database master password"
  type        = string
  sensitive   = true
}
