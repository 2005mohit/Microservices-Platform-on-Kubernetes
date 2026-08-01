variable "project" {
  type = string
}

variable "environment" {
  type = string
}

variable "cluster_name" {
  type = string
}

variable "eks_version" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "private_subnet_ids" {
  type = list(string)
}

variable "public_subnet_ids" {
  type = list(string)
}

variable "cluster_role_arn" {
  type = string
}

variable "cluster_role_name" {
  type = string
}

variable "node_role_arn" {
  type = string
}

variable "node_role_name" {
  type = string
}

variable "node_instance_type" {
  type = string
}

variable "node_desired_size" {
  type = number
}

variable "node_min_size" {
  type = number
}

variable "node_max_size" {
  type = number
}
