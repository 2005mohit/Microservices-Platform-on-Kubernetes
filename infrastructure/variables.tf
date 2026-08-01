variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "project" {
  description = "Project name (used in resource names)"
  type        = string
  default     = "devplatform"
}

variable "cluster_name" {
  description = "EKS cluster name"
  type        = string
  default     = "devplatform"
}

variable "eks_version" {
  description = "EKS Kubernetes version (check supported versions with: aws eks describe-cluster-versions)"
  type        = string
  default     = "1.35"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "Availability zones used for subnets"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b", "us-east-1c"]
}

variable "node_instance_type" {
  description = "EKS managed node group instance type (t3.small gives 9 pod IPs/node; t3.micro is limited to 2)"
  type        = string
  default     = "t3.small"
}

variable "node_desired_size" {
  description = "Desired number of worker nodes"
  type        = number
  default     = 3
}

variable "node_min_size" {
  description = "Minimum number of worker nodes"
  type        = number
  default     = 3
}

variable "node_max_size" {
  description = "Maximum number of worker nodes (for cluster autoscaler)"
  type        = number
  default     = 4
}

variable "db_name" {
  description = "RDS database name"
  type        = string
  default     = "devplatform"
}

variable "db_username" {
  description = "RDS database master username"
  type        = string
  default     = "devplatform"
}

variable "db_password" {
  description = "RDS database master password. Provide via terraform.tfvars (never commit it)."
  type        = string
  sensitive   = true
}

variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t4g.micro"
}
