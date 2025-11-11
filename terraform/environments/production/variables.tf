# Production environment variables

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "cluster_name" {
  description = "Name of the Kubernetes cluster"
  type        = string
  default     = "go-micro-commerce-prod"
}

variable "kubernetes_version" {
  description = "Kubernetes version to deploy"
  type        = string
  default     = "1.34.1"
}

variable "talos_version" {
  description = "Talos Linux version"
  type        = string
  default     = "v1.11.5"
}

# VM Configuration
variable "control_plane_nodes" {
  description = "Control plane node configurations with IP addresses from GCP"
  type = list(object({
    name = string
    ip   = string
  }))
  # Default placeholder - update with actual IPs in terraform.tfvars
  default = [
    {
      name = "control-plane-1"
      ip   = "10.0.1.10"
    }
  ]
}

variable "worker_nodes" {
  description = "Worker node configurations with IP addresses from GCP"
  type = list(object({
    name = string
    ip   = string
  }))
  # Default placeholder - update with actual IPs in terraform.tfvars
  default = [
    {
      name = "worker-1"
      ip   = "10.0.1.11"
    },
    {
      name = "worker-2"
      ip   = "10.0.1.12"
    }
  ]
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

# Networking
variable "cluster_endpoint" {
  description = "Kubernetes API endpoint (control plane IP or load balancer)"
  type        = string
  default     = "https://10.0.1.10:6443"
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

# Disk Configuration
variable "install_disk" {
  description = "Disk device for Talos installation (GCP VMs typically use /dev/sda)"
  type        = string
  default     = "/dev/sda"
}

variable "install_image" {
  description = "Talos installation image from Image Factory with extensions"
  type        = string
  # Use Image Factory URL with system extensions for Longhorn
  # Example: factory.talos.dev/installer/613e1592b2da41ae5e265e8789429f22e121aab91cb4deb6bc3c0b6262961245:v1.11.5
  # Generate at: https://factory.talos.dev/
  # Extensions needed: siderolabs/iscsi-tools, siderolabs/util-linux-tools
  default = ""
}

# Longhorn Storage
variable "enable_longhorn" {
  description = "Enable Longhorn storage configuration"
  type        = bool
  default     = true
}

variable "longhorn_disk_path" {
  description = "Path for Longhorn storage on worker nodes"
  type        = string
  default     = "/var/lib/longhorn"
}

# Resource Configuration
variable "control_plane_resources" {
  description = "Resource reservations for control plane (2 vCPU, 4GB RAM total)"
  type = object({
    cpu_reserved    = string
    memory_reserved = string
  })
  default = {
    cpu_reserved    = "500m"    # Reserve 0.5 CPU for system
    memory_reserved = "1Gi"     # Reserve 1GB for system
  }
}

variable "worker_resources" {
  description = "Resource reservations for worker nodes (2 vCPU, 8GB RAM each)"
  type = object({
    cpu_reserved    = string
    memory_reserved = string
  })
  default = {
    cpu_reserved    = "250m"    # Reserve 0.25 CPU for system
    memory_reserved = "512Mi"   # Reserve 512MB for system
  }
}

# Tags
variable "tags" {
  description = "Tags to apply to nodes"
  type        = map(string)
  default = {
    environment    = "production"
    managed-by     = "terraform"
    project        = "go-micro-commerce"
    cloud-provider = "gcp"
    region         = "asia-southeast2"
  }
}

# GCP Configuration
variable "gcp_project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "gcp_region" {
  description = "GCP region"
  type        = string
  default     = "asia-southeast2"
}

variable "gcp_zone" {
  description = "GCP zone"
  type        = string
  default     = "asia-southeast2-a"
}

variable "gcp_sa_key_path" {
  description = "Path to GCP service account key JSON file (DO NOT commit this file)"
  type        = string
  sensitive   = true
}

# External Secrets Configuration
variable "enable_external_secrets" {
  description = "Enable External Secrets Operator for GCP Secret Manager"
  type        = bool
  default     = true
}
