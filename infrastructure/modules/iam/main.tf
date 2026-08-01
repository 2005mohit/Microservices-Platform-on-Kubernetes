# EKS cluster control-plane role
resource "aws_iam_role" "eks_cluster_role" {
  name = "${var.cluster_name}-eks-cluster-role"
  description = "Allows the EKS control plane to manage cluster resources"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action    = "sts:AssumeRole"
        Effect    = "Allow"
        Principal = { Service = "eks.amazonaws.com" }
      }
    ]
  })
}

# EKS worker node role
resource "aws_iam_role" "eks_node_role" {
  name = "${var.cluster_name}-eks-node-role"
  description = "Allows EC2 worker nodes to register and join the EKS cluster"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action    = "sts:AssumeRole"
        Effect    = "Allow"
        Principal = { Service = "ec2.amazonaws.com" }
      }
    ]
  })
}
