# GCP Network Module
# Creates VPC, subnet with secondary ranges for GKE, and Cloud NAT

# VPC Network
resource "google_compute_network" "vpc" {
  name                    = var.network_name
  project                 = var.project_id
  auto_create_subnetworks = false
  routing_mode            = "REGIONAL"
  description             = "VPC network for ${var.environment} environment"
}

# Subnet with secondary IP ranges for GKE pods and services
resource "google_compute_subnetwork" "gke_subnet" {
  name          = "${var.network_name}-gke-subnet"
  project       = var.project_id
  region        = var.region
  network       = google_compute_network.vpc.id
  ip_cidr_range = var.subnet_cidr
  description   = "Subnet for GKE nodes in ${var.environment}"

  # Secondary IP ranges for GKE
  secondary_ip_range {
    range_name    = "gke-pods"
    ip_cidr_range = var.pods_cidr
  }

  secondary_ip_range {
    range_name    = "gke-services"
    ip_cidr_range = var.services_cidr
  }

  # Enable private Google access for private GKE nodes
  private_ip_google_access = true

  # Enable flow logs for network monitoring (optional)
  dynamic "log_config" {
    for_each = var.enable_flow_logs ? [1] : []
    content {
      aggregation_interval = "INTERVAL_5_SEC"
      flow_sampling        = 0.5
      metadata             = "INCLUDE_ALL_METADATA"
    }
  }
}

# Cloud Router for NAT
resource "google_compute_router" "router" {
  count   = var.enable_nat ? 1 : 0
  name    = "${var.network_name}-router"
  project = var.project_id
  region  = var.region
  network = google_compute_network.vpc.id

  bgp {
    asn = 64512
  }
}

# Cloud NAT for outbound internet access from private GKE nodes
resource "google_compute_router_nat" "nat" {
  count   = var.enable_nat ? 1 : 0
  name    = "${var.network_name}-nat"
  project = var.project_id
  region  = var.region
  router  = google_compute_router.router[0].name

  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}
