# CloudNative PostgreSQL Operator Module
# Installs CloudNative PG operator for managing PostgreSQL clusters

# Create namespace for the operator
resource "kubernetes_namespace" "cnpg" {
  count = var.create_namespace ? 1 : 0

  metadata {
    name = var.namespace

    labels = {
      "app.kubernetes.io/name"       = "cloudnative-pg"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

# Install CloudNative PG operator via Helm
resource "helm_release" "cloudnative_pg" {
  name       = "cloudnative-pg"
  namespace  = var.namespace
  repository = "https://cloudnative-pg.github.io/charts"
  chart      = "cloudnative-pg"
  version    = var.chart_version

  create_namespace = false # Already created above

  values = [
    yamlencode({
      replicaCount = var.replicas

      # Monitoring configuration
      monitoring = {
        podMonitorEnabled = var.enable_monitoring
      }

      # Resource limits
      resources = {
        limits = {
          cpu    = "500m"
          memory = "512Mi"
        }
        requests = {
          cpu    = "100m"
          memory = "128Mi"
        }
      }

      # Node affinity for infra pool
      nodeSelector = {
        workload-type = "infra"
      }

      tolerations = [
        {
          key      = "workload-type"
          operator = "Equal"
          value    = "infra"
          effect   = "NoSchedule"
        }
      ]

      # Security context
      securityContext = {
        runAsNonRoot = true
        runAsUser    = 1000
        fsGroup      = 1000
      }
    })
  ]

  depends_on = [kubernetes_namespace.cnpg]
}
