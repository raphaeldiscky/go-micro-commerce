# Production Environment - Outputs

# ============================================================================
# Network Outputs
# ============================================================================

output "network_name" {
  description = "VPC network name"
  value       = module.gcp_network.network_name
}

output "network_self_link" {
  description = "VPC network self link"
  value       = module.gcp_network.network_self_link
}

output "subnet_name" {
  description = "GKE subnet name"
  value       = module.gcp_network.subnet_name
}

output "nat_enabled" {
  description = "Whether Cloud NAT is enabled"
  value       = var.enable_nat
}

# ============================================================================
# GKE Cluster Outputs
# ============================================================================

output "cluster_name" {
  description = "GKE cluster name"
  value       = module.gke_cluster.cluster_name
}

output "cluster_location" {
  description = "GKE cluster location (zone/region)"
  value       = module.gke_cluster.cluster_location
}

output "cluster_endpoint" {
  description = "GKE cluster endpoint"
  value       = module.gke_cluster.cluster_endpoint
  sensitive   = true
}

output "cluster_ca_certificate" {
  description = "GKE cluster CA certificate"
  value       = module.gke_cluster.cluster_ca_certificate
  sensitive   = true
}

output "cluster_master_version" {
  description = "Kubernetes master version"
  value       = module.gke_cluster.cluster_master_version
}

output "kubeconfig_command" {
  description = "Command to configure kubectl"
  value       = module.gke_cluster.kubeconfig_command
}

output "stateful_pool_name" {
  description = "Stateful node pool name (databases)"
  value       = module.gke_cluster.stateful_pool_name
}

output "stateless_pool_name" {
  description = "Stateless node pool name (microservices)"
  value       = module.gke_cluster.stateless_pool_name
}

output "monitoring_pool_name" {
  description = "Monitoring node pool name (observability stack)"
  value       = module.gke_cluster.monitoring_pool_name
}

output "control_plane_pool_name" {
  description = "Control plane node pool name (operators, ArgoCD, ESO)"
  value       = module.gke_cluster.control_plane_pool_name
}

output "gateway_pool_name" {
  description = "Gateway node pool name (Traefik, Apollo Router, API Gateway)"
  value       = module.gke_cluster.gateway_pool_name
}

output "cost_summary" {
  description = "Estimated monthly cost breakdown"
  value       = module.gke_cluster.cost_summary
}

# ============================================================================
# Kubernetes Operators
# ============================================================================

# External Secrets Operator Outputs
output "eso_namespace" {
  description = "External Secrets Operator namespace"
  value       = module.external_secrets_operator.namespace
}

output "eso_status" {
  description = "External Secrets Operator status"
  value       = module.external_secrets_operator.status
}

output "eso_gcp_service_account_email" {
  description = "GCP service account email for Secret Manager access"
  value       = module.external_secrets_operator.gcp_service_account_email
}

output "eso_cluster_secret_store_name" {
  description = "ClusterSecretStore name for ExternalSecrets"
  value       = module.external_secrets_operator.cluster_secret_store_name
}

# CloudNative PostgreSQL Operator Outputs
output "cnpg_namespace" {
  description = "CloudNative PG operator namespace"
  value       = module.cloudnative_pg_operator.namespace
}

output "cnpg_status" {
  description = "CloudNative PG operator status"
  value       = module.cloudnative_pg_operator.status
}

# Strimzi Kafka Operator Outputs
output "kafka_namespace" {
  description = "Strimzi Kafka operator namespace"
  value       = module.strimzi_kafka_operator.namespace
}

output "kafka_status" {
  description = "Strimzi Kafka operator status"
  value       = module.strimzi_kafka_operator.status
}

# Redis Operator Outputs
output "redis_namespace" {
  description = "Redis operator namespace"
  value       = module.redis_operator.namespace
}

output "redis_status" {
  description = "Redis operator status"
  value       = module.redis_operator.status
}

# Monitoring Outputs
output "monitoring_namespace" {
  description = "Monitoring stack namespace"
  value       = module.monitoring.namespace
}

output "grafana_url" {
  description = "URL to access Grafana"
  value       = module.monitoring.grafana_url
}

output "grafana_admin_user" {
  description = "Grafana admin username"
  value       = module.monitoring.grafana_admin_user
}

output "prometheus_url" {
  description = "URL to access Prometheus"
  value       = module.monitoring.prometheus_url
}

# ArgoCD Outputs
output "argocd_namespace" {
  description = "ArgoCD namespace"
  value       = module.argocd.namespace
}

output "argocd_url" {
  description = "URL to access ArgoCD"
  value       = module.argocd.argocd_url
}

output "argocd_admin_user" {
  description = "ArgoCD admin username"
  value       = module.argocd.argocd_admin_user
}

output "argocd_password_command" {
  description = "Command to get ArgoCD admin password"
  value       = module.argocd.argocd_password_command
}

# cert-manager Outputs
output "cert_manager_namespace" {
  description = "cert-manager namespace"
  value       = module.cert_manager.namespace
}

output "cert_manager_letsencrypt_prod_issuer" {
  description = "Let's Encrypt production ClusterIssuer name"
  value       = module.cert_manager.letsencrypt_prod_issuer
}

output "cert_manager_letsencrypt_staging_issuer" {
  description = "Let's Encrypt staging ClusterIssuer name"
  value       = module.cert_manager.letsencrypt_staging_issuer
}

output "cert_manager_status" {
  description = "cert-manager deployment status"
  value       = module.cert_manager.status
}

# Traefik Outputs
output "traefik_namespace" {
  description = "Traefik namespace"
  value       = module.traefik.namespace
}

output "traefik_dashboard_url" {
  description = "URL to access Traefik dashboard"
  value       = module.traefik.dashboard_url
}

output "traefik_service_type" {
  description = "Traefik service type"
  value       = module.traefik.service_type
}

# ============================================================================
# Domain and DNS Outputs
# ============================================================================

output "api_url" {
  description = "Backend API base URL"
  value       = "https://${var.api_subdomain}.${var.domain_name}"
}

output "traefik_load_balancer_ip" {
  description = "Traefik LoadBalancer IP address (for DNS configuration)"
  value       = module.traefik.load_balancer_ip
}

output "cloudflare_dns_records" {
  description = "Cloudflare DNS records created by Terraform"
  value       = module.cloudflare_dns.dns_records
}

output "argocd_public_url" {
  description = "Public HTTPS URL for ArgoCD (via Cloudflare DNS + Traefik)"
  value       = module.cloudflare_dns.argocd_url
}

output "grafana_public_url" {
  description = "Public HTTPS URL for Grafana (via Cloudflare DNS + Traefik)"
  value       = module.cloudflare_dns.grafana_url
}

output "frontend_deployment_note" {
  description = "Note about frontend deployment"
  value       = "Frontend (https://go.micro.commerce.${var.domain_name}) is deployed via Cloudflare Pages (not managed by Terraform). See terraform/CLOUDFLARE_PAGES_SETUP.md for setup instructions."
}

# ============================================================================
# Deployment Summary
# ============================================================================

output "deployment_summary" {
  description = "Summary of deployed components"
  value = {
    infrastructure = {
      cluster_name     = module.gke_cluster.cluster_name
      cluster_location = module.gke_cluster.cluster_location
      network_name     = module.gcp_network.network_name
      cost_estimate    = module.gke_cluster.cost_summary
    }
    operators = {
      external_secrets = module.external_secrets_operator.status
      cloudnative_pg   = module.cloudnative_pg_operator.status
      strimzi_kafka    = module.strimzi_kafka_operator.status
      redis            = module.redis_operator.status
    }
    platform = {
      monitoring = module.monitoring.namespace
      argocd     = module.argocd.status
      traefik    = module.traefik.status
    }
    dns = {
      api_url     = "https://${var.api_subdomain}.${var.domain_name}"
      traefik_ip  = module.traefik.load_balancer_ip
      note        = "Frontend deployed via Cloudflare Pages (see CLOUDFLARE_PAGES_SETUP.md)"
    }
  }
}
