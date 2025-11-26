# GKE Cluster Module
# Creates cost-efficient GKE cluster with stateful and stateless node pools

# GKE Cluster (minimal default pool, will be removed)
resource "google_container_cluster" "primary" {
  name     = var.cluster_name
  project  = var.project_id
  location = var.region

  # Network configuration
  network    = var.network_name
  subnetwork = var.subnet_name

  # Remove default node pool immediately
  remove_default_node_pool = true
  initial_node_count       = 1

  # Kubernetes version and release channel
  min_master_version = var.kubernetes_version
  release_channel {
    channel = var.release_channel
  }

  # IP allocation policy for VPC-native cluster
  ip_allocation_policy {
    cluster_secondary_range_name  = var.pods_range_name
    services_secondary_range_name = var.services_range_name
  }

  # Workload Identity
  dynamic "workload_identity_config" {
    for_each = var.enable_workload_identity ? [1] : []
    content {
      workload_pool = "${var.project_id}.svc.id.goog"
    }
  }

  # Networking
  networking_mode = "VPC_NATIVE"

  default_max_pods_per_node = var.max_pods_per_node

  # Addons
  addons_config {
    horizontal_pod_autoscaling {
      disabled = false
    }
    http_load_balancing {
      disabled = false
    }
    gce_persistent_disk_csi_driver_config {
      enabled = true
    }
  }

  # Logging and monitoring
  logging_config {
    enable_components = ["SYSTEM_COMPONENTS", "WORKLOADS"]
  }

  monitoring_config {
    enable_components = ["SYSTEM_COMPONENTS", "STORAGE", "POD", "DEPLOYMENT", "STATEFULSET", "DAEMONSET", "HPA"]

    managed_prometheus {
      enabled = true
    }
  }

  # Security
  security_posture_config {
    mode = "BASIC"
  }

  # Maintenance policy
  maintenance_policy {
    daily_maintenance_window {
      start_time = "03:00"
    }
  }

  # Binary authorization
  binary_authorization {
    evaluation_mode = "DISABLED"
  }

  # Private cluster configuration
  private_cluster_config {
    enable_private_nodes    = var.enable_private_nodes    # Nodes get only private IPs (no external IPs)
    enable_private_endpoint = var.enable_private_endpoint # Keep master public for kubectl access
    master_ipv4_cidr_block  = var.master_ipv4_cidr_block  # /28 CIDR for control plane
  }

  # Shielded nodes
  enable_shielded_nodes = true
}

# Stateful Node Pool (for databases - PostgreSQL, Kafka, Redis)
resource "google_container_node_pool" "stateful" {
  count      = var.stateful_pool_enabled ? 1 : 0
  name       = "stateful-pool"
  project    = var.project_id
  location   = var.region
  node_locations = [var.zone] 
  cluster    = google_container_cluster.primary.name
  node_count = var.stateful_pool_node_count

  node_config {
    machine_type = var.stateful_pool_machine_type
    disk_size_gb = var.stateful_pool_disk_size_gb
    disk_type    = var.stateful_pool_disk_type
    image_type   = "COS_CONTAINERD"
    spot         = false # Regular VMs for stateful reliability

    # Labels for workload scheduling
    labels = {
      workload-type = "stateful"
      pool-type     = "regular"
    }

    # Taint to ensure only stateful workloads run here
    taint {
      key    = "workload-type"
      value  = "stateful"
      effect = "NO_SCHEDULE"
    }

    # OAuth scopes
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    # Metadata
    metadata = {
      disable-legacy-endpoints = "true"
    }

    # Workload Identity
    dynamic "workload_metadata_config" {
      for_each = var.enable_workload_identity ? [1] : []
      content {
        mode = "GKE_METADATA"
      }
    }

    # Shielded instance config
    shielded_instance_config {
      enable_secure_boot          = false
      enable_integrity_monitoring = true
    }
  }

  # Management
  management {
    auto_repair  = true
    auto_upgrade = true
  }

  # Upgrade settings
  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0  
  }
}

# Stateless Node Pool (for microservices - Spot VMs with autoscaling)
resource "google_container_node_pool" "stateless" {
  count    = var.stateless_pool_enabled ? 1 : 0
  name     = "stateless-pool"
  project  = var.project_id
  location = var.region
  node_locations = [var.zone]  # Keep all nodes in zone-a to fit CPU quota
  cluster  = google_container_cluster.primary.name

  # Autoscaling configuration
  autoscaling {
    min_node_count = var.stateless_pool_min_nodes
    max_node_count = var.stateless_pool_max_nodes
  }

  initial_node_count = var.stateless_pool_min_nodes

  node_config {
    machine_type = var.stateless_pool_machine_type
    disk_size_gb = var.stateless_pool_disk_size_gb
    disk_type    = var.stateless_pool_disk_type
    image_type   = "COS_CONTAINERD"
    spot         = true # Spot VMs for 60-91% cost savings!

    # Labels for workload scheduling
    labels = {
      workload-type = "stateless"
      pool-type     = "spot"
    }

    # OAuth scopes
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    # Metadata
    metadata = {
      disable-legacy-endpoints = "true"
    }

    # Workload Identity
    dynamic "workload_metadata_config" {
      for_each = var.enable_workload_identity ? [1] : []
      content {
        mode = "GKE_METADATA"
      }
    }

    # Shielded instance config
    shielded_instance_config {
      enable_secure_boot          = false
      enable_integrity_monitoring = true
    }
  }

  # Management
  management {
    auto_repair  = true
    auto_upgrade = true
  }

  # Upgrade settings
  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }
}

# Monitoring Node Pool (for observability stack - Prometheus, Grafana, Loki, Tempo, Alloy)
resource "google_container_node_pool" "monitoring" {
  count    = var.monitoring_pool_enabled ? 1 : 0
  name     = "monitoring-pool"
  project  = var.project_id
  location = var.region
  node_locations = [var.zone]  # Keep all nodes in zone-a to fit CPU quota
  cluster  = google_container_cluster.primary.name

  # Autoscaling configuration
  autoscaling {
    min_node_count = var.monitoring_pool_min_nodes
    max_node_count = var.monitoring_pool_max_nodes
  }

  initial_node_count = var.monitoring_pool_min_nodes

  node_config {
    machine_type = var.monitoring_pool_machine_type
    disk_size_gb = var.monitoring_pool_disk_size_gb
    disk_type    = var.monitoring_pool_disk_type
    image_type   = "COS_CONTAINERD"
    spot         = false # Regular VMs for monitoring reliability

    # Labels for workload scheduling
    labels = {
      workload-type = "monitoring"
      pool-type     = "regular"
    }

    # Taint to ensure only monitoring workloads run here
    taint {
      key    = "workload-type"
      value  = "monitoring"
      effect = "NO_SCHEDULE"
    }

    # OAuth scopes
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    # Metadata
    metadata = {
      disable-legacy-endpoints = "true"
    }

    # Workload Identity
    dynamic "workload_metadata_config" {
      for_each = var.enable_workload_identity ? [1] : []
      content {
        mode = "GKE_METADATA"
      }
    }

    # Shielded instance config
    shielded_instance_config {
      enable_secure_boot          = false
      enable_integrity_monitoring = true
    }
  }

  # Management
  management {
    auto_repair  = true
    auto_upgrade = true
  }

  # Upgrade settings
  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }
}

# Control Plane Node Pool (for operators, ArgoCD, ESO)
resource "google_container_node_pool" "control_plane" {
  count    = var.control_plane_pool_enabled ? 1 : 0
  name     = "control-plane-pool"
  project  = var.project_id
  location = var.region
  node_locations = [var.zone]  # Keep all nodes in zone-a to fit CPU quota
  cluster  = google_container_cluster.primary.name

  # Autoscaling configuration
  autoscaling {
    min_node_count = var.control_plane_pool_min_nodes
    max_node_count = var.control_plane_pool_max_nodes
  }

  initial_node_count = var.control_plane_pool_min_nodes

  node_config {
    machine_type = var.control_plane_pool_machine_type
    disk_size_gb = var.control_plane_pool_disk_size_gb
    disk_type    = var.control_plane_pool_disk_type
    image_type   = "COS_CONTAINERD"
    spot         = false # Regular VMs for control plane reliability

    # Labels for workload scheduling
    labels = {
      workload-type = "control-plane"
      pool-type     = "regular"
    }

    # Taint to ensure only control plane workloads run here
    taint {
      key    = "workload-type"
      value  = "control-plane"
      effect = "NO_SCHEDULE"
    }

    # OAuth scopes
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    # Metadata
    metadata = {
      disable-legacy-endpoints = "true"
    }

    # Workload Identity
    dynamic "workload_metadata_config" {
      for_each = var.enable_workload_identity ? [1] : []
      content {
        mode = "GKE_METADATA"
      }
    }

    # Shielded instance config
    shielded_instance_config {
      enable_secure_boot          = false
      enable_integrity_monitoring = true
    }
  }

  # Management
  management {
    auto_repair  = true
    auto_upgrade = true
  }

  # Upgrade settings
  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }
}

# Gateway Node Pool (for Traefik, Apollo Router, API Gateway)
resource "google_container_node_pool" "gateway" {
  count    = var.gateway_pool_enabled ? 1 : 0
  name     = "gateway-pool"
  project  = var.project_id
  location = var.region
  node_locations = [var.zone]  # Keep all nodes in zone-a to fit CPU quota
  cluster  = google_container_cluster.primary.name

  # Autoscaling configuration
  autoscaling {
    min_node_count = var.gateway_pool_min_nodes
    max_node_count = var.gateway_pool_max_nodes
  }

  initial_node_count = var.gateway_pool_min_nodes

  node_config {
    machine_type = var.gateway_pool_machine_type
    disk_size_gb = var.gateway_pool_disk_size_gb
    disk_type    = var.gateway_pool_disk_type
    image_type   = "COS_CONTAINERD"
    spot         = false # Regular VMs for gateway reliability

    # Labels for workload scheduling
    labels = {
      workload-type = "gateway"
      pool-type     = "regular"
    }

    # Taint to ensure only gateway workloads run here
    taint {
      key    = "workload-type"
      value  = "gateway"
      effect = "NO_SCHEDULE"
    }

    # OAuth scopes
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    # Metadata
    metadata = {
      disable-legacy-endpoints = "true"
    }

    # Workload Identity
    dynamic "workload_metadata_config" {
      for_each = var.enable_workload_identity ? [1] : []
      content {
        mode = "GKE_METADATA"
      }
    }

    # Shielded instance config
    shielded_instance_config {
      enable_secure_boot          = false
      enable_integrity_monitoring = true
    }
  }

  # Management
  management {
    auto_repair  = true
    auto_upgrade = true
  }

  # Upgrade settings
  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }
}
