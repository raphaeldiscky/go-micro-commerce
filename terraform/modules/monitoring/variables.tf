# Monitoring Module Variables

variable "namespace" {
  description = "Kubernetes namespace for monitoring stack"
  type        = string
  default     = "monitoring"
}

variable "create_namespace" {
  description = "Create the namespace if it doesn't exist"
  type        = bool
  default     = true
}

# Prometheus & Grafana (kube-prometheus-stack)
variable "kube_prometheus_stack_version" {
  description = "Helm chart version for kube-prometheus-stack"
  type        = string
  default     = "79.5.0"
}

variable "grafana_admin_password" {
  description = "Admin password for Grafana"
  type        = string
  default     = "admin"
  sensitive   = true
}

variable "prometheus_retention" {
  description = "Prometheus data retention period"
  type        = string
  default     = "15d"
}

variable "prometheus_storage_size" {
  description = "Storage size for Prometheus"
  type        = string
  default     = "50Gi"
}

# Loki
variable "loki_version" {
  description = "Helm chart version for Loki"
  type        = string
  default     = "6.46.0"
}

variable "loki_storage_size" {
  description = "Storage size for Loki"
  type        = string
  default     = "30Gi"
}

# Tempo
variable "tempo_version" {
  description = "Helm chart version for Tempo"
  type        = string
  default     = "1.24.0"
}

variable "tempo_storage_size" {
  description = "Storage size for Tempo"
  type        = string
  default     = "20Gi"
}
