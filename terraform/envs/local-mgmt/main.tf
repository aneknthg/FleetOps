terraform {
  required_version = ">= 1.6.0"

  required_providers {
    kind = {
      source  = "tehcyx/kind"
      version = "~> 0.11.0"
    }
  }
}

provider "kind" {}

module "cluster" {
  source       = "../../modules/kind-cluster"
  name         = "mgmt"
  worker_count = 2
}

output "name" {
  value = module.cluster.name
}

output "endpoint" {
  value = module.cluster.endpoint
}

output "kubeconfig_path" {
  value = module.cluster.kubeconfig_path
}
