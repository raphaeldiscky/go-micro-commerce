# Traefik Module
# Installs Traefik ingress controller for load balancing and routing

# Create namespace for Traefik
resource "kubernetes_namespace" "traefik" {
  count = var.create_namespace ? 1 : 0

  metadata {
    name = var.namespace

    labels = {
      "app.kubernetes.io/name"       = "traefik"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

# Install Traefik via Helm
resource "helm_release" "traefik" {
  name       = "traefik"
  namespace  = var.namespace
  repository = "https://traefik.github.io/charts"
  chart      = "traefik"
  version    = var.chart_version

  create_namespace = false

  values = [
    yamlencode({
      # Deployment configuration
      deployment = {
        replicas = var.replicas
      }

      # Service configuration
      service = {
        type = var.service_type
      }

      # Ports configuration
      ports = {
        web = {
          port     = 80
          exposedPort = 80
        }
        websecure = {
          port     = 443
          exposedPort = 443
          tls = {
            enabled = true
          }
        }
      }

      # Ingress routes
      ingressRoute = {
        dashboard = {
          enabled = var.enable_dashboard
        }
      }

      # Resource limits
      resources = {
        limits = {
          cpu    = "1000m"
          memory = "1Gi"
        }
        requests = {
          cpu    = "200m"
          memory = "256Mi"
        }
      }

      # Prometheus metrics
      metrics = {
        prometheus = {
          enabled = var.enable_metrics
          serviceMonitor = {
            enabled = var.enable_metrics
          }
        }
      }

      # Enable access logs
      logs = {
        general = {
          level = "INFO"
        }
        access = {
          enabled = true
        }
      }

      # Additional configuration
      additionalArguments = [
        "--providers.kubernetesingress.ingressclass=traefik",
        "--providers.kubernetescrd",
        "--providers.kubernetesgateway",
      ]

      # Security
      securityContext = {
        capabilities = {
          drop = ["ALL"]
        }
        readOnlyRootFilesystem = true
        runAsNonRoot = true
        runAsUser = 65532
      }

      # Node affinity for gateway pool
      nodeSelector = {
        workload-type = "gateway"
      }

      tolerations = [
        {
          key      = "workload-type"
          operator = "Equal"
          value    = "gateway"
          effect   = "NoSchedule"
        }
      ]
    })
  ]

  depends_on = [kubernetes_namespace.traefik]
}

# Data source to get Traefik service details (including LoadBalancer IP)
data "kubernetes_service" "traefik" {
  metadata {
    name      = helm_release.traefik.name
    namespace = var.namespace
  }

  depends_on = [helm_release.traefik]
}
