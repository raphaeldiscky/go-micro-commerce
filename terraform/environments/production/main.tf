# Production environment Talos cluster configuration
# This file uses the talos-cluster module to deploy a 3-node cluster on GCP VMs

terraform {
  required_version = ">= 1.12.2"

  required_providers {
    talos = {
      source  = "siderolabs/talos"
      version = "~> 0.9.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 3.1.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.38.0"
    }
  }
}

# Configure providers
provider "talos" {}

# Kubernetes and Helm providers configured to use generated kubeconfig
# This requires a two-phase apply:
#   1. terraform apply -target=module.talos_cluster  (creates cluster and kubeconfig)
#   2. terraform apply                                (deploys Kubernetes resources)
provider "kubernetes" {
  config_path = module.talos_cluster.kubeconfig_file
}

provider "helm" {
  kubernetes = {
    config_path = module.talos_cluster.kubeconfig_file
  }
}

# Deploy Talos cluster
module "talos_cluster" {
  source = "../../modules/talos-cluster"

  environment        = var.environment
  cluster_name       = var.cluster_name
  kubernetes_version = var.kubernetes_version
  talos_version      = var.talos_version

  # VM configuration (IP addresses from GCP VMs)
  control_plane_nodes = var.control_plane_nodes
  worker_nodes        = var.worker_nodes

  # External IPs for certificate SANs
  control_plane_external_ips = var.control_plane_external_ips
  worker_external_ips        = var.worker_external_ips

  # Cluster networking
  cluster_endpoint = var.cluster_endpoint
  pod_cidr         = var.pod_cidr
  service_cidr     = var.service_cidr

  # Disk configuration
  install_disk  = var.install_disk
  install_image = var.install_image

  # Longhorn storage
  enable_longhorn    = var.enable_longhorn
  longhorn_disk_path = var.longhorn_disk_path

  # Resource configuration
  control_plane_resources = var.control_plane_resources
  worker_resources        = var.worker_resources

  # GCP configuration
  gcp_project_id  = var.gcp_project_id
  gcp_region      = var.gcp_region
  gcp_zone        = var.gcp_zone
  gcp_sa_key_path = var.gcp_sa_key_path

  # External Secrets
  enable_external_secrets = var.enable_external_secrets

  # Tags
  tags = var.tags
}
