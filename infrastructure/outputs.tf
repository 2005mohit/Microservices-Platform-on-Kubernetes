output "cluster_name" {
  value = module.eks.cluster_name
}

output "cluster_endpoint" {
  value = module.eks.cluster_endpoint
}

output "cluster_oidc_issuer" {
  value = module.eks.cluster_oidc_issuer
}

output "kubeconfig_command" {
  value = "aws eks update-kubeconfig --region ${var.region} --name ${var.cluster_name}"
}

output "rds_endpoint" {
  value = module.rds.endpoint
}

output "vpc_id" {
  value = module.vpc.vpc_id
}
