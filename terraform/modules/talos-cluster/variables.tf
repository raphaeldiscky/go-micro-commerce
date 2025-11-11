# Environment configuration
variable "environment" {
  description = "Environment name (dev, staging, production)"
  type        = string
  validation {
    condition     = contains(["dev", "staging", "production"], var.environment)
    error_message = "Environment must be dev, staging, or production."
  }
}

variable "cluster_name" {
  description = "Name of the Kubernetes cluster"
  type        = string
  default     = "go-micro-commerce"
}

variable "kubernetes_version" {
  description = "Kubernetes version to deploy"
  type        = string
  default     = "1.34.1"
}

# Talos configuration
variable "talos_version" {
  description = "Talos Linux version"
  type        = string
  default     = "v1.11.5"
}

variable "install_disk" {
  description = "Disk device for Talos installation (GCP typically uses /dev/sda)"
  type        = string
  default     = "/dev/sda"
}

variable "install_image" {
  description = "Talos installation image (factory.talos.dev URL with extensions)"
  type        = string
  default     = ""
}

# Control plane configuration
variable "control_plane_nodes" {
  description = "List of control plane node configurations"
  type = list(object({
    name = string
    ip   = string
  }))
  validation {
    condition     = length(var.control_plane_nodes) >= 1
    error_message = "At least one control plane node is required."
  }
}

# Worker configuration
variable "worker_nodes" {
  description = "List of worker node configurations"
  type = list(object({
    name = string
    ip   = string
  }))
  validation {
    condition     = length(var.worker_nodes) >= 1
    error_message = "At least one worker node is required."
  }
}

# External IPs for certificate SANs
variable "control_plane_external_ips" {
  description = "External IP addresses for control plane nodes (for Talos API certificate SANs)"
  type        = list(string)
  default     = []
}

variable "worker_external_ips" {
  description = "External IP addresses for worker nodes (for Talos API certificate SANs)"
  type        = list(string)
  default     = []
}

# Cluster networking
variable "cluster_endpoint" {
  description = "Kubernetes API endpoint URL"
  type        = string
}

variable "pod_cidr" {
  description = "CIDR block for pod network"
  type        = string
  default     = "10.244.0.0/16"
}

variable "service_cidr" {
  description = "CIDR block for service network"
  type        = string
  default     = "10.96.0.0/12"
}

# Longhorn configuration
variable "enable_longhorn" {
  description = "Enable Longhorn storage configuration in Talos"
  type        = bool
  default     = true
}

variable "longhorn_disk_path" {
  description = "Path for Longhorn storage on worker nodes"
  type        = string
  default     = "/var/lib/longhorn"
}

# Additional machine config patches
variable "control_plane_patches" {
  description = "Additional config patches for control plane nodes (YAML list)"
  type        = list(string)
  default     = []
}

variable "worker_patches" {
  description = "Additional config patches for worker nodes (YAML list)"
  type        = list(string)
  default     = []
}

# Resource limits
variable "control_plane_resources" {
  description = "Resource limits for control plane"
  type = object({
    cpu_reserved    = string
    memory_reserved = string
  })
  default = {
    cpu_reserved    = "500m"
    memory_reserved = "1Gi"
  }
}

variable "worker_resources" {
  description = "Resource limits for worker nodes"
  type = object({
    cpu_reserved    = string
    memory_reserved = string
  })
  default = {
    cpu_reserved    = "250m"
    memory_reserved = "512Mi"
  }
}

# Tags
variable "tags" {
  description = "Tags to apply as labels on nodes"
  type        = map(string)
  default     = {}
}

# GCP Configuration
variable "gcp_project_id" {
  description = "GCP Project ID for External Secrets Operator and Secret Manager"
  type        = string
  default     = ""
}

variable "gcp_region" {
  description = "GCP region for deployment"
  type        = string
  default     = "asia-southeast2"
}

variable "gcp_zone" {
  description = "GCP zone for deployment"
  type        = string
  default     = "asia-southeast2-a"
}

variable "gcp_sa_key_path" {
  description = "Path to GCP service account key JSON file for External Secrets Operator"
  type        = string
  default     = ""
  sensitive   = true
}

# External Secrets Configuration
variable "enable_external_secrets" {
  description = "Enable External Secrets Operator for GCP Secret Manager integration"
  type        = bool
  default     = false
}
