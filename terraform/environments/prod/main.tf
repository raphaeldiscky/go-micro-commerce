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

  depends_on = [module.gke_cluster]
}

# CloudNative PostgreSQL Operator
module "cloudnative_pg_operator" {
  source = "../../modules/cloudnative-pg-operator"

  namespace         = var.cnpg_namespace
  create_namespace  = true
  chart_version     = var.cnpg_chart_version
  enable_monitoring = true

  depends_on = [module.gke_cluster]
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

  namespace                     = var.monitoring_namespace
  create_namespace              = true
  kube_prometheus_stack_version = var.kube_prometheus_stack_version
  grafana_admin_password        = var.grafana_admin_password
  prometheus_retention          = var.prometheus_retention
  prometheus_storage_size       = var.prometheus_storage_size
  loki_version                  = var.loki_version
  loki_storage_size             = var.loki_storage_size
  tempo_version                 = var.tempo_version
  tempo_storage_size            = var.tempo_storage_size

  depends_on = [
    module.cloudnative_pg_operator,
    module.strimzi_kafka_operator,
    module.redis_operator
  ]
}

# ArgoCD for GitOps (manages all applications)
module "argocd" {
  source = "../../modules/argocd"

  namespace        = var.argocd_namespace
  create_namespace = true
  chart_version    = var.argocd_chart_version
  admin_password   = var.argocd_admin_password
  git_repo_url     = var.argocd_git_repo_url
  git_repo_path    = var.argocd_git_repo_path
  enable_bootstrap = var.argocd_enable_bootstrap

  depends_on = [module.monitoring]
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

# Configure Cloudflare DNS records for backend API
module "cloudflare_dns" {
  source = "../../modules/cloudflare-dns"

  domain_name         = var.domain_name
  api_subdomain       = var.api_subdomain
  enable_api_wildcard = var.enable_api_wildcard
  traefik_ip          = module.traefik.load_balancer_ip

  depends_on = [module.traefik]
}
