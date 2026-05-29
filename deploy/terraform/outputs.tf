output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = module.vpc.private_subnet_ids
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = module.vpc.public_subnet_ids
}

output "eks_cluster_endpoint" {
  description = "EKS cluster endpoint"
  value       = module.eks.cluster_endpoint
}

output "eks_cluster_name" {
  description = "EKS cluster name"
  value       = module.eks.cluster_name
}

output "rds_endpoint" {
  description = "RDS endpoint"
  value       = module.rds.endpoint
}

output "rds_reader_endpoint" {
  description = "RDS reader endpoint (read replica)"
  value       = module.rds.reader_endpoint
}

output "msk_bootstrap_brokers" {
  description = "MSK bootstrap brokers"
  value       = module.msk.bootstrap_brokers
}

output "msk_bootstrap_brokers_tls" {
  description = "MSK TLS bootstrap brokers"
  value       = module.msk.bootstrap_brokers_tls
}

output "ecr_repository_urls" {
  description = "ECR repository URLs per service"
  value       = module.ecr.repository_urls
}

output "kong_alb_dns" {
  description = "Kong ALB DNS name"
  value       = module.alb.kong_alb_dns
}
