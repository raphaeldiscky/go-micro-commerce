# External Secrets Operator Module Variables

# GCP Configuration
variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "cluster_name" {
  description = "GKE cluster name for Workload Identity"
  type        = string
}

variable "cluster_location" {
  description = "GKE cluster location (zone or region)"
  type        = string
}

# Kubernetes Configuration
variable "namespace" {
  description = "Kubernetes namespace for External Secrets Operator"
  type        = string
  default     = "external-secrets"
}

variable "create_namespace" {
  description = "Create the namespace if it doesn't exist"
  type        = bool
  default     = true
}

# Helm Chart Configuration
variable "chart_version" {
  description = "Helm chart version for External Secrets Operator"
  type        = string
  default     = "1.0.0"
}

variable "replicas" {
  description = "Number of External Secrets Operator replicas"
  type        = number
  default     = 2
}

variable "enable_monitoring" {
  description = "Enable Prometheus monitoring"
  type        = bool
  default     = true
}

# Service Account Configuration
variable "gcp_service_account_name" {
  description = "Name for the GCP service account"
  type        = string
  default     = "external-secrets-operator"
}

variable "k8s_service_account_name" {
  description = "Name for the Kubernetes service account"
  type        = string
  default     = "external-secrets-sa"
}

# ClusterSecretStore Configuration
variable "create_cluster_secret_store" {
  description = "Create a ClusterSecretStore for Google Secret Manager"
  type        = bool
  default     = true
}

variable "cluster_secret_store_name" {
  description = "Name for the ClusterSecretStore"
  type        = string
  default     = "gcp-secret-manager"
}
