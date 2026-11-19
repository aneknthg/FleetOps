output "name" {
  description = "Name of the created kind cluster."
  value       = kind_cluster.this.name
}

output "endpoint" {
  description = "Kubernetes API server endpoint."
  value       = kind_cluster.this.endpoint
}

output "kubeconfig_path" {
  description = "Path to the kubeconfig written by the kind provider."
  value       = kind_cluster.this.kubeconfig_path
}
