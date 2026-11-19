locals {
  azs = slice(data.aws_availability_zones.available.names, 0, 2)
}

data "aws_availability_zones" "available" {
  state = "available"
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "6.7.3"

  name = "${var.cluster_name}-vpc"
  cidr = "10.40.0.0/16"
  azs  = local.azs

  private_subnets = ["10.40.1.0/24", "10.40.2.0/24"]
  public_subnets  = ["10.40.101.0/24", "10.40.102.0/24"]

  enable_nat_gateway = true
  single_nat_gateway = true

  public_subnet_tags = { "kubernetes.io/role/elb" = "1" }
  private_subnet_tags = {
    "kubernetes.io/role/internal-elb"       = "1"
    "kubernetes.io/cluster/${var.cluster_name}" = "shared"
  }

  tags = { Project = "FleetOpsLab", ManagedBy = "Terraform" }
}

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "21.29.0"

  name               = var.cluster_name
  kubernetes_version = var.kubernetes_version

  enable_irsa                              = true
  enable_cluster_creator_admin_permissions = true
  endpoint_private_access                 = true
  endpoint_public_access                  = true
  endpoint_public_access_cidrs             = var.endpoint_public_access_cidrs

  vpc_id                   = module.vpc.vpc_id
  subnet_ids               = module.vpc.private_subnets
  control_plane_subnet_ids = module.vpc.public_subnets

  addons = {
    coredns                = { most_recent = true }
    eks-pod-identity-agent = { most_recent = true, before_compute = true }
    kube-proxy             = { most_recent = true }
    vpc-cni                = { most_recent = true, before_compute = true }
  }

  eks_managed_node_groups = {
    lab = {
      instance_types = [var.instance_type]
      min_size       = 2
      max_size       = 2
      desired_size   = 2
      capacity_type  = "ON_DEMAND"
    }
  }

  tags = { Project = "FleetOpsLab", ManagedBy = "Terraform" }
}
