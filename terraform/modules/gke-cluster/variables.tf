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
  default     = 100
}

variable "stateful_pool_disk_type" {
  description = "Disk type for stateful pool nodes"
  type        = string
  default     = "pd-ssd"
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
  default     = 10
}

variable "stateless_pool_machine_type" {
  description = "Machine type for stateless pool"
  type        = string
  default     = "e2-medium"
}

variable "stateless_pool_disk_size_gb" {
  description = "Disk size in GB for stateless pool nodes"
  type        = number
  default     = 30
}

variable "stateless_pool_disk_type" {
  description = "Disk type for stateless pool nodes"
  type        = string
  default     = "pd-balanced"
}

variable "enable_workload_identity" {
  description = "Enable Workload Identity for secure pod authentication"
  type        = bool
  default     = true
}
