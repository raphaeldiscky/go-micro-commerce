# GKE Cluster Module Variables

variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "GCP region for the cluster"
  type        = string
}

variable "zone" {
  description = "GCP zone for the cluster"
  type        = string
}

variable "cluster_name" {
  description = "Name of the GKE cluster"
  type        = string
}

variable "network_name" {
  description = "Name of the VPC network"
  type        = string
}

variable "subnet_name" {
  description = "Name of the subnet"
  type        = string
}

variable "pods_range_name" {
  description = "Name of the secondary IP range for pods"
  type        = string
  default     = "gke-pods"
}

variable "services_range_name" {
  description = "Name of the secondary IP range for services"
  type        = string
  default     = "gke-services"
}

variable "kubernetes_version" {
  description = "Kubernetes version for the cluster"
  type        = string
  default     = "1.33"
}

variable "release_channel" {
  description = "Release channel for GKE cluster"
  type        = string
  default     = "REGULAR"
}

variable "max_pods_per_node" {
  description = "Maximum number of pods per node"
  type        = number
  default     = 110
}

# Stateful Pool Variables (for databases)
variable "stateful_pool_enabled" {
  description = "Enable stateful node pool for databases"
  type        = bool
  default     = true
}

variable "stateful_pool_node_count" {
  description = "Number of nodes in stateful pool"
  type        = number
  default     = 3
}

variable "stateful_pool_machine_type" {
  description = "Machine type for stateful pool"
  type        = string
  default     = "e2-standard-2"
}

variable "stateful_pool_disk_size_gb" {
  description = "Disk size in GB for stateful pool nodes"
  type        = number
  default     = 25
}

variable "stateful_pool_disk_type" {
  description = "Disk type for stateful pool nodes"
  type        = string
  default     = "pd-balanced"
}

# Stateless Pool Variables (for microservices - Spot VMs)
variable "stateless_pool_enabled" {
  description = "Enable stateless node pool for microservices"
  type        = bool
  default     = true
}

variable "stateless_pool_min_nodes" {
  description = "Minimum nodes in stateless pool"
  type        = number
  default     = 2
}

variable "stateless_pool_max_nodes" {
  description = "Maximum nodes in stateless pool"
  type        = number
  default     = 25
}

variable "stateless_pool_machine_type" {
  description = "Machine type for stateless pool"
  type        = string
  default     = "e2-medium"
}

variable "stateless_pool_disk_size_gb" {
  description = "Disk size in GB for stateless pool nodes"
  type        = number
  default     = 25
}

variable "stateless_pool_disk_type" {
  description = "Disk type for stateless pool nodes"
  type        = string
  default     = "pd-balanced"
}

# Monitoring Pool Variables (for observability stack - Prometheus, Grafana, Loki, Tempo, Alloy)
variable "monitoring_pool_enabled" {
  description = "Enable monitoring node pool for observability stack"
  type        = bool
  default     = true
}

variable "monitoring_pool_min_nodes" {
  description = "Minimum nodes in monitoring pool"
  type        = number
  default     = 1
}

variable "monitoring_pool_max_nodes" {
  description = "Maximum nodes in monitoring pool"
  type        = number
  default     = 3
}

variable "monitoring_pool_machine_type" {
  description = "Machine type for monitoring pool"
  type        = string
  default     = "e2-medium"
}

variable "monitoring_pool_disk_size_gb" {
  description = "Disk size in GB for monitoring pool nodes"
  type        = number
  default     = 25
}

variable "monitoring_pool_disk_type" {
  description = "Disk type for monitoring pool nodes"
  type        = string
  default     = "pd-balanced"
}

variable "enable_workload_identity" {
  description = "Enable Workload Identity for secure pod authentication"
  type        = bool
  default     = true
}

# Control Plane Pool Variables (for operators, ArgoCD, ESO)
variable "control_plane_pool_enabled" {
  description = "Enable control plane node pool for operators and control plane components"
  type        = bool
  default     = true
}

variable "control_plane_pool_min_nodes" {
  description = "Minimum nodes in control plane pool"
  type        = number
  default     = 1
}

variable "control_plane_pool_max_nodes" {
  description = "Maximum nodes in control plane pool"
  type        = number
  default     = 2
}

variable "control_plane_pool_machine_type" {
  description = "Machine type for control plane pool"
  type        = string
  default     = "e2-small"
}

variable "control_plane_pool_disk_size_gb" {
  description = "Disk size in GB for control plane pool nodes"
  type        = number
  default     = 20
}

variable "control_plane_pool_disk_type" {
  description = "Disk type for control plane pool nodes"
  type        = string
  default     = "pd-balanced"
}

# Gateway Pool Variables (for Traefik, Apollo Router, API Gateway)
variable "gateway_pool_enabled" {
  description = "Enable gateway node pool for ingress and API gateways"
  type        = bool
  default     = true
}

variable "gateway_pool_min_nodes" {
  description = "Minimum nodes in gateway pool"
  type        = number
  default     = 1
}

variable "gateway_pool_max_nodes" {
  description = "Maximum nodes in gateway pool"
  type        = number
  default     = 3
}

variable "gateway_pool_machine_type" {
  description = "Machine type for gateway pool"
  type        = string
  default     = "e2-medium"
}

variable "gateway_pool_disk_size_gb" {
  description = "Disk size in GB for gateway pool nodes"
  type        = number
  default     = 20
}

variable "gateway_pool_disk_type" {
  description = "Disk type for gateway pool nodes"
  type        = string
  default     = "pd-balanced"
}

variable "enable_private_nodes" {
  description = "Enable private nodes (nodes without public IPs, use Cloud NAT for internet)"
  type        = bool
  default     = true
}

variable "enable_private_endpoint" {
  description = "Enable private endpoint (restrict master access to private network only)"
  type        = bool
  default     = false
}

variable "master_ipv4_cidr_block" {
  description = "CIDR block for the Kubernetes master (must be /28, cannot overlap with VPC ranges)"
  type        = string
  default     = "10.13.0.0/28"
}
