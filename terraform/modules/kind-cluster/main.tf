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

resource "kind_cluster" "this" {
  name       = var.name
  node_image = var.node_image

  kind_config {
    kind        = "Cluster"
    api_version = "kind.x-k8s.io/v1alpha4"

    dynamic "node" {
      for_each = range(var.worker_count + 1)
      content {
        role = node.value == 0 ? "control-plane" : "worker"
      }
    }
  }
}
