# CloudNative PostgreSQL Operator Module Variables

variable "namespace" {
  description = "Kubernetes namespace for the operator"
  type        = string
  default     = "cnpg-system"
}

variable "create_namespace" {
  description = "Create the namespace if it doesn't exist"
  type        = bool
  default     = true
}

variable "chart_version" {
  description = "Helm chart version for CloudNative PG operator"
  type        = string
  default     = "0.26.1"
}

variable "replicas" {
  description = "Number of operator replicas"
  type        = number
  default     = 1
}

variable "enable_monitoring" {
  description = "Enable Prometheus monitoring"
  type        = bool
  default     = true
}
