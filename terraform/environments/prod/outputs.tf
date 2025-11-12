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

output "cost_summary" {
  description = "Estimated monthly cost breakdown"
  value       = module.gke_cluster.cost_summary
}

# ============================================================================
# Kubernetes Operators
# ============================================================================

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
      cloudnative_pg = module.cloudnative_pg_operator.status
      strimzi_kafka  = module.strimzi_kafka_operator.status
      redis          = module.redis_operator.status
    }
    platform = {
      monitoring = module.monitoring.namespace
      argocd     = module.argocd.status
      traefik    = module.traefik.status
    }
  }
}
