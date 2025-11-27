# Production Environment - Main Configuration
# Composes all infrastructure and Kubernetes modules

# GCP Network (VPC, subnet, NAT)
module "gcp_network" {
  source = "../../modules/gcp-network"

  project_id       = var.project_id
  region           = var.region
  environment      = var.environment
  network_name     = var.network_name
  subnet_cidr      = var.subnet_cidr
  pods_cidr        = var.pods_cidr
  services_cidr    = var.services_cidr
  enable_flow_logs = var.enable_flow_logs
  enable_nat       = var.enable_nat
}

# GKE Cluster with stateful and stateless pools
module "gke_cluster" {
  source = "../../modules/gke-cluster"

  project_id           = var.project_id
  region               = var.region
  zone                 = var.zone
  cluster_name         = var.cluster_name
  network_name         = module.gcp_network.network_name
  subnet_name          = module.gcp_network.subnet_name
  pods_range_name      = module.gcp_network.pods_range_name
  services_range_name  = module.gcp_network.services_range_name
  kubernetes_version   = var.kubernetes_version
  release_channel      = var.release_channel
  max_pods_per_node    = var.max_pods_per_node

  # Stateful pool (databases)
  stateful_pool_enabled       = var.stateful_pool_enabled
  stateful_pool_node_count    = var.stateful_pool_node_count
  stateful_pool_machine_type  = var.stateful_pool_machine_type
  stateful_pool_disk_size_gb  = var.stateful_pool_disk_size_gb
  stateful_pool_disk_type     = var.stateful_pool_disk_type

  # Stateless pool (microservices - Spot VMs)
  stateless_pool_enabled      = var.stateless_pool_enabled
  stateless_pool_min_nodes    = var.stateless_pool_min_nodes
  stateless_pool_max_nodes    = var.stateless_pool_max_nodes
  stateless_pool_machine_type = var.stateless_pool_machine_type
  stateless_pool_disk_size_gb = var.stateless_pool_disk_size_gb
  stateless_pool_disk_type    = var.stateless_pool_disk_type

  # Monitoring pool (observability stack)
  monitoring_pool_enabled      = var.monitoring_pool_enabled
  monitoring_pool_min_nodes    = var.monitoring_pool_min_nodes
  monitoring_pool_max_nodes    = var.monitoring_pool_max_nodes
  monitoring_pool_machine_type = var.monitoring_pool_machine_type
  monitoring_pool_disk_size_gb = var.monitoring_pool_disk_size_gb
  monitoring_pool_disk_type    = var.monitoring_pool_disk_type

  # Control plane pool (operators, ArgoCD, ESO)
  control_plane_pool_enabled      = var.control_plane_pool_enabled
  control_plane_pool_min_nodes    = var.control_plane_pool_min_nodes
  control_plane_pool_max_nodes    = var.control_plane_pool_max_nodes
  control_plane_pool_machine_type = var.control_plane_pool_machine_type
  control_plane_pool_disk_size_gb = var.control_plane_pool_disk_size_gb
  control_plane_pool_disk_type    = var.control_plane_pool_disk_type

  # Gateway pool (Traefik, Apollo Router, API Gateway)
  gateway_pool_enabled      = var.gateway_pool_enabled
  gateway_pool_min_nodes    = var.gateway_pool_min_nodes
  gateway_pool_max_nodes    = var.gateway_pool_max_nodes
  gateway_pool_machine_type = var.gateway_pool_machine_type
  gateway_pool_disk_size_gb = var.gateway_pool_disk_size_gb
  gateway_pool_disk_type    = var.gateway_pool_disk_type

  # Private cluster configuration
  enable_private_nodes     = var.enable_private_nodes
  enable_private_endpoint  = var.enable_private_endpoint
  master_ipv4_cidr_block   = var.master_ipv4_cidr_block

  enable_workload_identity = var.enable_workload_identity

  depends_on = [module.gcp_network]
}

# External Secrets Operator (manages secrets from Google Secret Manager)
module "external_secrets_operator" {
  source = "../../modules/external-secrets-operator"

  project_id       = var.project_id
  cluster_name     = module.gke_cluster.cluster_name
  cluster_location = var.zone
  namespace        = var.eso_namespace
  create_namespace = true
  chart_version    = var.eso_chart_version
  replicas         = var.eso_replicas

  gcp_service_account_name   = var.eso_gcp_service_account_name
  k8s_service_account_name   = var.eso_k8s_service_account_name
  create_cluster_secret_store = var.eso_create_cluster_secret_store
  cluster_secret_store_name   = var.eso_cluster_secret_store_name

  enable_monitoring = true

  depends_on = [module.gke_cluster, module.monitoring]
}

# CloudNative PostgreSQL Operator
module "cloudnative_pg_operator" {
  source = "../../modules/cloudnative-pg-operator"

  namespace         = var.cnpg_namespace
  create_namespace  = true
  chart_version     = var.cnpg_chart_version
  enable_monitoring = true

  depends_on = [module.gke_cluster, module.monitoring]
}

# Strimzi Kafka Operator
module "strimzi_kafka_operator" {
  source = "../../modules/strimzi-kafka-operator"

  namespace        = var.kafka_namespace
  create_namespace = true
  chart_version    = var.kafka_chart_version
  watch_namespaces = var.kafka_watch_namespaces
  log_level        = "INFO"

  depends_on = [module.gke_cluster]
}

# Redis Operator
module "redis_operator" {
  source = "../../modules/redis-operator"

  namespace        = var.redis_namespace
  create_namespace = true
  chart_version    = var.redis_chart_version
  watch_namespaces = var.redis_watch_namespaces

  depends_on = [module.gke_cluster]
}

# Monitoring Stack (Prometheus, Grafana, Loki, Tempo)
module "monitoring" {
  source = "../../modules/monitoring"

  namespace                           = var.monitoring_namespace
  create_namespace                    = true
  kube_prometheus_stack_chart_version = var.kube_prometheus_stack_chart_version
  grafana_admin_password              = var.grafana_admin_password
  grafana_enable_ingress              = var.grafana_enable_ingress
  grafana_domain_name                 = var.grafana_domain_name
  grafana_tls_issuer                  = var.grafana_tls_issuer
  prometheus_retention                = var.prometheus_retention
  prometheus_storage_size             = var.prometheus_storage_size
  loki_chart_version                  = var.loki_chart_version
  loki_storage_size                   = var.loki_storage_size
  tempo_chart_version                 = var.tempo_chart_version
  tempo_storage_size                  = var.tempo_storage_size

  depends_on = [module.gke_cluster]
}

# cert-manager for automated TLS certificate management
module "cert_manager" {
  source = "../../modules/cert-manager"

  namespace                       = var.cert_manager_namespace
  create_namespace                = true
  chart_version                   = var.cert_manager_chart_version
  replicas                        = var.cert_manager_replicas
  enable_monitoring               = true
  create_cluster_issuers          = var.cert_manager_create_cluster_issuers
  letsencrypt_email               = var.cert_manager_letsencrypt_email
  letsencrypt_staging_issuer_name = var.cert_manager_letsencrypt_staging_issuer_name
  letsencrypt_prod_issuer_name    = var.cert_manager_letsencrypt_prod_issuer_name

  depends_on = [module.gke_cluster, module.traefik]
}

# ArgoCD for GitOps (manages all applications)
module "argocd" {
  source = "../../modules/argocd"

  namespace           = var.argocd_namespace
  create_namespace    = true
  chart_version       = var.argocd_chart_version
  git_repo_url        = var.argocd_git_repo_url
  git_repo_path       = var.argocd_git_repo_path
  enable_bootstrap    = var.argocd_enable_bootstrap
  enable_ingress      = var.argocd_enable_ingress
  domain_name         = var.argocd_domain_name
  tls_issuer          = var.argocd_tls_issuer
  git_username        = var.argocd_git_username
  git_token           = var.argocd_git_token
  git_ssh_private_key = var.argocd_git_ssh_private_key

  depends_on = [module.monitoring, module.cert_manager]
}

# Traefik Ingress Controller
module "traefik" {
  source = "../../modules/traefik"

  namespace        = var.traefik_namespace
  create_namespace = true
  chart_version    = var.traefik_chart_version
  replicas         = var.traefik_replicas
  service_type     = var.traefik_service_type
  enable_dashboard = var.traefik_enable_dashboard
  enable_metrics   = true

  depends_on = [module.monitoring]
}

# ============================================================================
# Domain and DNS Configuration
# ============================================================================
# Note: Frontend (go.micro.commerce.discky.com) is deployed via Cloudflare Pages
#       Terraform only manages backend API DNS records

# Configure Cloudflare DNS records for backend API and ArgoCD
module "cloudflare_dns" {
  source = "../../modules/cloudflare-dns"

  cloudflare_zone_id  = var.cloudflare_zone_id
  domain_name         = var.domain_name
  api_subdomain       = var.api_subdomain
  enable_api_wildcard = var.enable_api_wildcard
  enable_argocd_dns   = var.enable_argocd_dns
  argocd_subdomain    = var.argocd_subdomain
  enable_grafana_dns  = var.enable_grafana_dns
  grafana_subdomain   = var.grafana_subdomain
  traefik_ip          = module.traefik.load_balancer_ip

  depends_on = [module.traefik]
}
