# GKE Cluster Module Outputs

output "cluster_id" {
  description = "The ID of the GKE cluster"
  value       = google_container_cluster.primary.id
}

output "cluster_name" {
  description = "The name of the GKE cluster"
  value       = google_container_cluster.primary.name
}

output "cluster_endpoint" {
  description = "The endpoint for the GKE cluster"
  value       = google_container_cluster.primary.endpoint
  sensitive   = true
}

output "cluster_ca_certificate" {
  description = "The CA certificate for the GKE cluster"
  value       = google_container_cluster.primary.master_auth[0].cluster_ca_certificate
  sensitive   = true
}

output "cluster_location" {
  description = "The location (zone/region) of the GKE cluster"
  value       = google_container_cluster.primary.location
}

output "cluster_master_version" {
  description = "The Kubernetes master version"
  value       = google_container_cluster.primary.master_version
}

output "cluster_self_link" {
  description = "The self link of the GKE cluster"
  value       = google_container_cluster.primary.self_link
}

output "stateful_pool_name" {
  description = "The name of the stateful node pool"
  value       = var.stateful_pool_enabled ? google_container_node_pool.stateful[0].name : null
}

output "stateless_pool_name" {
  description = "The name of the stateless node pool"
  value       = var.stateless_pool_enabled ? google_container_node_pool.stateless[0].name : null
}

output "kubeconfig_command" {
  description = "Command to configure kubectl"
  value       = "gcloud container clusters get-credentials ${google_container_cluster.primary.name} --zone=${google_container_cluster.primary.location} --project=${var.project_id}"
}

output "cost_summary" {
  description = "Estimated monthly cost summary"
  value = {
    stateful_pool  = var.stateful_pool_enabled ? "~$105/month (3 × e2-standard-2 regular VMs)" : "disabled"
    stateless_pool = var.stateless_pool_enabled ? "~$21-105/month (${var.stateless_pool_min_nodes}-${var.stateless_pool_max_nodes} × e2-medium Spot VMs)" : "disabled"
    total          = "~$126-210/month"
    savings        = "~60% vs all regular VMs"
  }
}
