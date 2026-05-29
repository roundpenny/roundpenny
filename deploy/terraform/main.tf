module "vpc" {
  source = "./modules/vpc"

  environment          = var.environment
  cidr                 = var.vpc_cidr
  availability_zones   = var.availability_zones
  public_subnet_cidrs  = var.public_subnet_cidrs
  private_subnet_cidrs = var.private_subnet_cidrs
}

module "eks" {
  source = "./modules/eks"

  environment         = var.environment
  vpc_id              = module.vpc.vpc_id
  private_subnet_ids  = module.vpc.private_subnet_ids
  cluster_version     = var.eks_cluster_version
  node_instance_types = var.eks_node_instance_types
  min_nodes           = var.eks_min_nodes
  max_nodes           = var.eks_max_nodes
  desired_nodes       = var.eks_desired_nodes
}

module "rds" {
  source = "./modules/rds"

  environment           = var.environment
  vpc_id                = module.vpc.vpc_id
  private_subnet_ids    = module.vpc.private_subnet_ids
  eks_security_group_id = module.eks.cluster_security_group_id
  instance_class        = var.rds_instance_class
  allocated_storage     = var.rds_allocated_storage
  multi_az              = var.rds_multi_az
  backup_retention_days = var.rds_backup_retention_days
  username              = var.db_username
  password              = var.db_password
}

module "msk" {
  source = "./modules/msk"

  environment           = var.environment
  vpc_id                = module.vpc.vpc_id
  private_subnet_ids    = module.vpc.private_subnet_ids
  eks_security_group_id = module.eks.cluster_security_group_id
  broker_instance_type  = var.msk_broker_instance_type
  broker_count          = var.msk_broker_count
  ebs_volume_size       = var.msk_ebs_volume_size
}

module "ecr" {
  source = "./modules/ecr"
}

module "alb" {
  source = "./modules/alb"

  environment       = var.environment
  vpc_id            = module.vpc.vpc_id
  public_subnet_ids = module.vpc.public_subnet_ids
  certificate_arn   = var.certificate_arn
}

module "security_groups" {
  source = "./modules/security-groups"

  environment           = var.environment
  vpc_id                = module.vpc.vpc_id
  alb_security_group_id = module.alb.alb_security_group_id
}
