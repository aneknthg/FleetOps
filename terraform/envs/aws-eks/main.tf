terraform {
  required_version = ">= 1.6.0"

  backend "s3" {}

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.59"
    }
  }
}

variable "region" {
  description = "AWS region for the EKS cluster."
  type        = string
  default     = "us-east-1"
}

variable "cluster_name" {
  description = "EKS cluster name."
  type        = string
  default     = "fleetops-lab"
}

variable "kubernetes_version" {
  description = "EKS Kubernetes minor version."
  type        = string
  default     = "1.34"
}

variable "endpoint_public_access_cidrs" {
  description = "Set to your current public IP as a /32 before planning."
  type        = list(string)
}

provider "aws" {
  region = var.region
}

module "eks_cluster" {
  source = "../../modules/eks-cluster"

  cluster_name                 = var.cluster_name
  kubernetes_version           = var.kubernetes_version
  endpoint_public_access_cidrs = var.endpoint_public_access_cidrs
}

output "cluster_name" {
  value = module.eks_cluster.cluster_name
}

output "cluster_endpoint" {
  value = module.eks_cluster.cluster_endpoint
}
