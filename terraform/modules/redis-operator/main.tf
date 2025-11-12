# Redis Operator Module
# Installs Redis operator for managing Redis clusters

# Create namespace for the operator
resource "kubernetes_namespace" "redis" {
  count = var.create_namespace ? 1 : 0

  metadata {
    name = var.namespace

    labels = {
      "app.kubernetes.io/name"       = "redis-operator"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

# Install Redis operator via Helm
resource "helm_release" "redis_operator" {
  name       = "redis-operator"
  namespace  = var.namespace
  repository = "https://ot-container-kit.github.io/helm-charts/"
  chart      = "redis-operator"
  version    = var.chart_version

  create_namespace = false # Already created above

  values = [
    yamlencode({
      # Operator replicas
      replicaCount = 1

      # Watch specific namespaces or all
      watchNamespace = length(var.watch_namespaces) > 0 ? join(",", var.watch_namespaces) : ""

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

      # Security context
      securityContext = {
        runAsNonRoot = true
        runAsUser    = 1000
      }

      # Service account
      serviceAccount = {
        create = true
        name   = "redis-operator"
      }
    })
  ]

  depends_on = [kubernetes_namespace.redis]
}
