# ---------------------------------------------------------------
# IAM policy attachments live here (not in the iam module) so that
# Terraform can guarantee they exist BEFORE the cluster / node group.
# ---------------------------------------------------------------
resource "aws_iam_role_policy_attachment" "cluster_eks_policy" {
  role       = var.cluster_role_name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
}

resource "aws_iam_role_policy_attachment" "node_worker_policy" {
  role       = var.node_role_name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
}

resource "aws_iam_role_policy_attachment" "node_cni_policy" {
  role       = var.node_role_name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
}

resource "aws_iam_role_policy_attachment" "node_ecr_readonly" {
  role       = var.node_role_name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

# ---------------------------------------------------------------
# EKS control plane
# ---------------------------------------------------------------
resource "aws_eks_cluster" "this" {
  name     = var.cluster_name
  role_arn = var.cluster_role_arn
  version  = var.eks_version

  vpc_config {
    subnet_ids              = concat(var.private_subnet_ids, var.public_subnet_ids)
    endpoint_private_access = true
    endpoint_public_access  = true
    public_access_cidrs     = ["0.0.0.0/0"]
  }

  enabled_cluster_log_types = ["api", "audit", "authenticator"]

  depends_on = [
    aws_iam_role_policy_attachment.cluster_eks_policy
  ]

  tags = {
    Name = var.cluster_name
  }
}

# ---------------------------------------------------------------
# EKS addons (default versions chosen by AWS)
# ---------------------------------------------------------------
resource "aws_eks_addon" "vpc_cni" {
  cluster_name = aws_eks_cluster.this.name
  addon_name   = "vpc-cni"
}

resource "aws_eks_addon" "coredns" {
  cluster_name = aws_eks_cluster.this.name
  addon_name   = "coredns"
}

resource "aws_eks_addon" "kube_proxy" {
  cluster_name = aws_eks_cluster.this.name
  addon_name   = "kube-proxy"
}

# ---------------------------------------------------------------
# OIDC provider (used later for IRSA: cluster-autoscaler,
# AWS Load Balancer Controller, fluent-bit -> CloudWatch)
# ---------------------------------------------------------------
data "tls_certificate" "cluster" {
  url = aws_eks_cluster.this.identity[0].oidc[0].issuer
}

resource "aws_iam_openid_connect_provider" "cluster" {
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.cluster.certificates[0].sha1_fingerprint]
  url             = aws_eks_cluster.this.identity[0].oidc[0].issuer
}

# ---------------------------------------------------------------
# Managed node group (runs in the private subnets)
# t3.small ENI limits: 3 ENIs x 4 IPv4 = 9 usable pod IPs per node.
# The NodeConfig pins kubelet maxPods to match the CNI's 9 IPs so the
# scheduler and the VPC CNI agree on pod density per node.
# ---------------------------------------------------------------
resource "aws_launch_template" "node" {
  name          = "${var.cluster_name}-nodes-lt"
  instance_type = var.node_instance_type

  user_data = base64encode(<<-EOT
    MIME-Version: 1.0
    Content-Type: multipart/mixed; boundary="//"

    --//
    Content-Type: application/node.eks.aws

    ---
    apiVersion: node.eks.aws/v1alpha1
    kind: NodeConfig
    spec:
      kubelet:
        config:
          maxPods: 9
          clusterDNS:
          - 172.20.0.10

    --//--
  EOT
  )

  tags = {
    Name = "${var.cluster_name}-nodes"
  }
}

resource "aws_eks_node_group" "main" {
  cluster_name    = aws_eks_cluster.this.name
  node_group_name = "main"
  node_role_arn   = var.node_role_arn
  subnet_ids      = var.private_subnet_ids
  capacity_type   = "ON_DEMAND"

  launch_template {
    id      = aws_launch_template.node.id
    version = aws_launch_template.node.latest_version
  }

  scaling_config {
    desired_size = var.node_desired_size
    min_size     = var.node_min_size
    max_size     = var.node_max_size
  }

  update_config {
    max_unavailable = 1
  }

  depends_on = [
    aws_iam_role_policy_attachment.node_worker_policy,
    aws_iam_role_policy_attachment.node_cni_policy,
    aws_iam_role_policy_attachment.node_ecr_readonly
  ]

  tags = {
    Name = "${var.cluster_name}-nodes"
  }
}
