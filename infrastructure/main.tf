module "vpc" {
  source             = "./modules/vpc"
  project            = var.project
  environment        = var.environment
  cluster_name       = var.cluster_name
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
}

module "iam" {
  source       = "./modules/iam"
  project      = var.project
  cluster_name = var.cluster_name
}

module "eks" {
  source             = "./modules/eks"
  project            = var.project
  environment        = var.environment
  cluster_name       = var.cluster_name
  eks_version        = var.eks_version
  vpc_id             = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids
  public_subnet_ids  = module.vpc.public_subnet_ids
  cluster_role_arn   = module.iam.cluster_role_arn
  cluster_role_name  = module.iam.cluster_role_name
  node_role_arn      = module.iam.node_role_arn
  node_role_name     = module.iam.node_role_name
  node_instance_type = var.node_instance_type
  node_desired_size  = var.node_desired_size
  node_min_size      = var.node_min_size
  node_max_size      = var.node_max_size
}

module "rds" {
  source             = "./modules/rds"
  project            = var.project
  environment        = var.environment
  vpc_id             = module.vpc.vpc_id
  vpc_cidr           = var.vpc_cidr
  private_subnet_ids = module.vpc.private_subnet_ids
  db_name            = var.db_name
  db_username        = var.db_username
  db_password        = var.db_password
  db_instance_class  = var.db_instance_class
}
