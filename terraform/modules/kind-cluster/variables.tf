variable "name" {
  description = "Name of the kind cluster."
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9-]*$", var.name))
    error_message = "Cluster names must contain lowercase letters, digits, or hyphens and start with a letter or digit."
  }
}

variable "node_image" {
  description = "Pinned kind node image used for every cluster node."
  type        = string
  default     = "kindest/node:v1.36.6"
}

variable "worker_count" {
  description = "Number of worker nodes in addition to the control plane."
  type        = number
  default     = 2

  validation {
    condition     = var.worker_count >= 0 && var.worker_count == floor(var.worker_count)
    error_message = "worker_count must be a non-negative integer."
  }
}
