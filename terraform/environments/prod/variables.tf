# Production Environment - Variables

# GCP Project Configuration
variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "asia-southeast2"
}

variable "zone" {
  description = "GCP zone"
  type        = string
  default     = "asia-southeast2-a"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "prod"
}

# Network Configuration
variable "network_name" {
  description = "Name of the VPC network"
  type        = string
  default     = "go-micro-commerce-prod"
}

variable "subnet_cidr" {
  description = "CIDR range for the primary subnet"
  type        = string
  default     = "10.0.0.0/20"

  validation {
    condition     = can(cidrhost(var.subnet_cidr, 0))
    error_message = "subnet_cidr must be a valid CIDR range (e.g., 10.0.0.0/20)"
  }
}

variable "pods_cidr" {
  description = "Secondary CIDR range for GKE pods"
  type        = string
  default     = "10.4.0.0/14"

  validation {
    condition     = can(cidrhost(var.pods_cidr, 0))
    error_message = "pods_cidr must be a valid CIDR range (e.g., 10.4.0.0/14)"
  }
}

variable "services_cidr" {
  description = "Secondary CIDR range for GKE services"
  type        = string
  default     = "10.8.0.0/20"

  validation {
    condition     = can(cidrhost(var.services_cidr, 0))
    error_message = "services_cidr must be a valid CIDR range (e.g., 10.8.0.0/20)"
  }
}

variable "enable_flow_logs" {
  description = "Enable VPC flow logs"
  type        = bool
  default     = true
}

variable "enable_nat" {
  description = "Enable Cloud NAT"
  type        = bool
  default     = true
}

# GKE Cluster Configuration
variable "cluster_name" {
  description = "Name of the GKE cluster"
  type        = string
  default     = "go-micro-commerce-prod"
}

variable "kubernetes_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.33"

  validation {
    condition     = can(regex("^[0-9]+\\.[0-9]+$", var.kubernetes_version))
    error_message = "kubernetes_version must be in format 'X.Y' (e.g., 1.33)"
  }
}

variable "release_channel" {
  description = "GKE release channel"
  type        = string
  default     = "REGULAR"
}

variable "max_pods_per_node" {
  description = "Maximum pods per node"
  type        = number
  default     = 110
}

variable "enable_workload_identity" {
  description = "Enable Workload Identity"
  type        = bool
  default     = true
}

# Stateful Pool (Databases)
variable "stateful_pool_enabled" {
  description = "Enable stateful node pool"
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
  description = "Disk size for stateful pool"
  type        = number
  default     = 100
}

variable "stateful_pool_disk_type" {
  description = "Disk type for stateful pool"
  type        = string
  default     = "pd-balanced"
}

# Stateless Pool (Microservices - Spot VMs)
variable "stateless_pool_enabled" {
  description = "Enable stateless node pool"
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
  description = "Disk size for stateless pool"
  type        = number
  default     = 30
}

variable "stateless_pool_disk_type" {
  description = "Disk type for stateless pool"
  type        = string
  default     = "pd-balanced"
}

# Monitoring Pool (Observability Stack)
variable "monitoring_pool_enabled" {
  description = "Enable monitoring node pool"
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
  description = "Disk size for monitoring pool"
  type        = number
  default     = 30
}

variable "monitoring_pool_disk_type" {
  description = "Disk type for monitoring pool"
  type        = string
  default     = "pd-balanced"
}

# Infra Pool (Operators, ArgoCD, ESO)
variable "infra_pool_enabled" {
  description = "Enable infra node pool"
  type        = bool
  default     = true
}

variable "infra_pool_min_nodes" {
  description = "Minimum nodes in infra pool"
  type        = number
  default     = 1
}

variable "infra_pool_max_nodes" {
  description = "Maximum nodes in infra pool"
  type        = number
  default     = 2
}

variable "infra_pool_machine_type" {
  description = "Machine type for infra pool"
  type        = string
  default     = "e2-small"
}

variable "infra_pool_disk_size_gb" {
  description = "Disk size for infra pool"
  type        = number
  default     = 30
}

variable "infra_pool_disk_type" {
  description = "Disk type for infra pool"
  type        = string
  default     = "pd-balanced"
}

# Gateway Pool (Traefik, Apollo Router, API Gateway)
variable "gateway_pool_enabled" {
  description = "Enable gateway node pool"
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
  description = "Disk size for gateway pool"
  type        = number
  default     = 30
}

variable "gateway_pool_disk_type" {
  description = "Disk type for gateway pool"
  type        = string
  default     = "pd-balanced"
}

# External Secrets Operator
variable "eso_namespace" {
  description = "Namespace for External Secrets Operator"
  type        = string
  default     = "external-secrets"
}

variable "eso_chart_version" {
  description = "Helm chart version for External Secrets Operator"
  type        = string
  default     = "1.0.0"
}

variable "eso_replicas" {
  description = "Number of External Secrets Operator replicas"
  type        = number
  default     = 2
}

variable "eso_gcp_service_account_name" {
  description = "GCP service account name for External Secrets Operator"
  type        = string
  default     = "external-secrets-operator"
}

variable "eso_k8s_service_account_name" {
  description = "Kubernetes service account name for External Secrets Operator"
  type        = string
  default     = "external-secrets-sa"
}

variable "eso_create_cluster_secret_store" {
  description = "Create ClusterSecretStore for Google Secret Manager"
  type        = bool
  default     = true
}

variable "eso_cluster_secret_store_name" {
  description = "Name of the ClusterSecretStore"
  type        = string
  default     = "gcp-secret-manager"
}

# CloudNative PostgreSQL Operator
variable "cnpg_namespace" {
  description = "Namespace for CloudNative PG operator"
  type        = string
  default     = "cnpg-system"
}

variable "cnpg_chart_version" {
  description = "Helm chart version for CloudNative PG"
  type        = string
  default     = "0.26.1"
}

# Strimzi Kafka Operator
variable "kafka_namespace" {
  description = "Namespace for Strimzi Kafka operator"
  type        = string
  default     = "kafka-system"
}

variable "kafka_chart_version" {
  description = "Helm chart version for Strimzi Kafka"
  type        = string
  default     = "0.48.0"
}

variable "kafka_watch_namespaces" {
  description = "Namespaces to watch for Kafka resources"
  type        = list(string)
  default     = []
}

# Redis Operator
variable "redis_namespace" {
  description = "Namespace for Redis operator"
  type        = string
  default     = "redis-system"
}

variable "redis_chart_version" {
  description = "Helm chart version for Redis operator"
  type        = string
  default     = "0.22.2"
}

variable "redis_watch_namespaces" {
  description = "Namespaces to watch for Redis resources"
  type        = list(string)
  default     = []
}

# Monitoring Stack
variable "monitoring_namespace" {
  description = "Namespace for monitoring stack"
  type        = string
  default     = "monitoring"
}

variable "kube_prometheus_stack_chart_version" {
  description = "Helm chart version for kube-prometheus-stack"
  type        = string
  default     = "79.5.0"
}

variable "grafana_admin_password" {
  description = "Admin password for Grafana"
  type        = string
  sensitive   = true
}

variable "grafana_enable_ingress" {
  description = "Enable Ingress for Grafana web UI"
  type        = bool
  default     = true
}

variable "grafana_domain_name" {
  description = "Domain name for Grafana web UI"
  type        = string
  default     = "grafana.api.discky.com"
}

variable "grafana_tls_issuer" {
  description = "cert-manager ClusterIssuer name for Grafana TLS"
  type        = string
  default     = "letsencrypt-prod"
}

variable "prometheus_retention" {
  description = "Prometheus data retention period"
  type        = string
  default     = "30d"
}

variable "prometheus_storage_size" {
  description = "Storage size for Prometheus"
  type        = string
  default     = "100Gi"
}

variable "loki_chart_version" {
  description = "Helm chart version for Loki"
  type        = string
  default     = "6.46.0"
}

variable "loki_storage_size" {
  description = "Storage size for Loki"
  type        = string
  default     = "50Gi"
}

variable "tempo_chart_version" {
  description = "Helm chart version for Tempo"
  type        = string
  default     = "1.24.0"
}

variable "tempo_storage_size" {
  description = "Storage size for Tempo"
  type        = string
  default     = "30Gi"
}

# ArgoCD
variable "argocd_namespace" {
  description = "Namespace for ArgoCD"
  type        = string
  default     = "argocd"
}

variable "argocd_chart_version" {
  description = "Helm chart version for ArgoCD"
  type        = string
  default     = "9.1.2"
}

variable "argocd_admin_password" {
  description = "Admin password for ArgoCD"
  type        = string
  default     = ""
  sensitive   = true
}

variable "argocd_git_repo_url" {
  description = "Git repository URL for ArgoCD applications"
  type        = string
  default     = ""
}

variable "argocd_git_repo_path" {
  description = "Path in Git repository for applications"
  type        = string
  default     = "deployments/k8s"
}

variable "argocd_enable_bootstrap" {
  description = "Enable bootstrap ApplicationSet"
  type        = bool
  default     = false
}

variable "argocd_enable_ingress" {
  description = "Enable Ingress for ArgoCD web UI"
  type        = bool
  default     = true
}

variable "argocd_domain_name" {
  description = "Domain name for ArgoCD web UI"
  type        = string
  default     = "argocd.api.discky.com"
}

variable "argocd_tls_issuer" {
  description = "cert-manager ClusterIssuer name for ArgoCD TLS"
  type        = string
  default     = "letsencrypt-prod"
}

variable "argocd_git_username" {
  description = "Git username for repository authentication (for HTTPS method)"
  type        = string
  default     = "git"
}

variable "argocd_git_token" {
  description = "Git token for repository authentication (for HTTPS method)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "argocd_git_ssh_private_key" {
  description = "SSH private key for repository authentication (for SSH method)"
  type        = string
  default     = ""
  sensitive   = true
}

# cert-manager
variable "cert_manager_namespace" {
  description = "Namespace for cert-manager"
  type        = string
  default     = "cert-manager"
}

variable "cert_manager_chart_version" {
  description = "Helm chart version for cert-manager"
  type        = string
  default     = "v1.19.1"
}

variable "cert_manager_replicas" {
  description = "Number of cert-manager replicas"
  type        = number
  default     = 1
}

variable "cert_manager_create_cluster_issuers" {
  description = "Create Let's Encrypt ClusterIssuers"
  type        = bool
  default     = true
}

variable "cert_manager_letsencrypt_email" {
  description = "Email address for Let's Encrypt notifications"
  type        = string
}

variable "cert_manager_letsencrypt_staging_issuer_name" {
  description = "Name for Let's Encrypt staging issuer"
  type        = string
  default     = "letsencrypt-staging"
}

variable "cert_manager_letsencrypt_prod_issuer_name" {
  description = "Name for Let's Encrypt production issuer"
  type        = string
  default     = "letsencrypt-prod"
}

# Traefik
variable "traefik_namespace" {
  description = "Namespace for Traefik"
  type        = string
  default     = "traefik"
}

variable "traefik_chart_version" {
  description = "Helm chart version for Traefik"
  type        = string
  default     = "37.3.0"
}

variable "traefik_replicas" {
  description = "Number of Traefik replicas"
  type        = number
  default     = 3
}

variable "traefik_service_type" {
  description = "Service type for Traefik"
  type        = string
  default     = "LoadBalancer"
}

variable "traefik_enable_dashboard" {
  description = "Enable Traefik dashboard"
  type        = bool
  default     = true
}

# ============================================================================
# Cloudflare Configuration
# ============================================================================
# Note: Frontend (go.micro.commerce.discky.com) is deployed via Cloudflare Pages
#       Only backend API DNS records are managed by Terraform

variable "cloudflare_api_token" {
  description = "Cloudflare API token for DNS management"
  type        = string
  sensitive   = true
}

variable "cloudflare_zone_id" {
  description = "Cloudflare Zone ID for the domain"
  type        = string
  sensitive   = true
}

variable "domain_name" {
  description = "Domain name managed in Cloudflare"
  type        = string
  default     = "discky.com"

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9.-]*\\.[a-z]{2,}$", var.domain_name))
    error_message = "domain_name must be a valid domain name (e.g., example.com)"
  }
}

variable "api_subdomain" {
  description = "Subdomain for backend API"
  type        = string
  default     = "api"
}

variable "enable_api_wildcard" {
  description = "Enable wildcard DNS record for API subdomains"
  type        = bool
  default     = false
}

variable "enable_argocd_dns" {
  description = "Enable DNS record for ArgoCD web UI"
  type        = bool
  default     = true
}

variable "argocd_subdomain" {
  description = "Full subdomain for ArgoCD (e.g., argocd.api or just argocd)"
  type        = string
  default     = "argocd.api"
}

variable "enable_grafana_dns" {
  description = "Enable DNS record for Grafana monitoring UI"
  type        = bool
  default     = true
}

variable "grafana_subdomain" {
  description = "Full subdomain for Grafana (e.g., grafana.api or just grafana)"
  type        = string
  default     = "grafana.api"
}

variable "enable_traefik_dns" {
  description = "Enable DNS record for Traefik dashboard UI"
  type        = bool
  default     = true
}

variable "traefik_subdomain" {
  description = "Full subdomain for Traefik dashboard (e.g., traefik)"
  type        = string
  default     = "traefik"
}

# ============================================================================
# Private Cluster Configuration
# ============================================================================

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

  validation {
    condition     = can(cidrhost(var.master_ipv4_cidr_block, 0)) && can(regex("/28$", var.master_ipv4_cidr_block))
    error_message = "master_ipv4_cidr_block must be a valid /28 CIDR range (e.g., 10.13.0.0/28)"
  }
}
