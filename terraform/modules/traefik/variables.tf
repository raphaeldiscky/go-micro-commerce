# Traefik Module Variables

variable "namespace" {
  description = "Kubernetes namespace for Traefik"
  type        = string
  default     = "traefik"
}

variable "create_namespace" {
  description = "Create the namespace if it doesn't exist"
  type        = bool
  default     = true
}

variable "chart_version" {
  description = "Helm chart version for Traefik"
  type        = string
  default     = "33.2.1"
}

variable "replicas" {
  description = "Number of Traefik replicas"
  type        = number
  default     = 2
}

variable "service_type" {
  description = "Service type for Traefik (LoadBalancer, NodePort, ClusterIP)"
  type        = string
  default     = "LoadBalancer"
}

variable "enable_dashboard" {
  description = "Enable Traefik dashboard"
  type        = bool
  default     = true
}

variable "enable_metrics" {
  description = "Enable Prometheus metrics"
  type        = bool
  default     = true
}
