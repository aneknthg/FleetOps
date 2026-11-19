terraform {
  required_version = ">= 1.6.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.59"
    }
  }
}

variable "cluster_name" {
  description = "Name of the EKS cluster."
  type        = string
}

variable "kubernetes_version" {
  description = "EKS Kubernetes minor version. Upgrade by one minor through a reviewed change."
  type        = string
  default     = "1.34"
}

variable "endpoint_public_access_cidrs" {
  description = "Operator IPv4 CIDRs allowed to reach the public EKS API endpoint; set to your current IP/32."
  type        = list(string)
}

variable "instance_type" {
  description = "EC2 instance type for the small managed node group."
  type        = string
  default     = "t3.small"
}
