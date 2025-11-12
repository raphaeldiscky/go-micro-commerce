# Redis Operator Module Variables

variable "namespace" {
  description = "Kubernetes namespace for the operator"
  type        = string
  default     = "redis-system"
}

variable "create_namespace" {
  description = "Create the namespace if it doesn't exist"
  type        = bool
  default     = true
}

variable "chart_version" {
  description = "Helm chart version for Redis operator"
  type        = string
  default     = "0.18.0"
}

variable "watch_namespaces" {
  description = "Namespaces to watch for Redis resources (empty = all namespaces)"
  type        = list(string)
  default     = []
}
