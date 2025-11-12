# GCP Network Module Outputs

output "network_id" {
  description = "The ID of the VPC network"
  value       = google_compute_network.vpc.id
}

output "network_name" {
  description = "The name of the VPC network"
  value       = google_compute_network.vpc.name
}

output "network_self_link" {
  description = "The self link of the VPC network"
  value       = google_compute_network.vpc.self_link
}

output "subnet_id" {
  description = "The ID of the GKE subnet"
  value       = google_compute_subnetwork.gke_subnet.id
}

output "subnet_name" {
  description = "The name of the GKE subnet"
  value       = google_compute_subnetwork.gke_subnet.name
}

output "subnet_self_link" {
  description = "The self link of the GKE subnet"
  value       = google_compute_subnetwork.gke_subnet.self_link
}

output "pods_range_name" {
  description = "The name of the secondary IP range for pods"
  value       = "gke-pods"
}

output "services_range_name" {
  description = "The name of the secondary IP range for services"
  value       = "gke-services"
}

output "router_name" {
  description = "The name of the Cloud Router (if NAT is enabled)"
  value       = var.enable_nat ? google_compute_router.router[0].name : null
}

output "nat_name" {
  description = "The name of the Cloud NAT (if enabled)"
  value       = var.enable_nat ? google_compute_router_nat.nat[0].name : null
}
