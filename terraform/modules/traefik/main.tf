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

      # Ports configuration - use standard ports 80/443
      ports = {
        web = {
          port = 80
        }
        websecure = {
          port = 443
          tls = {
            enabled = true
          }
        }
      }

      # Disable Helm chart's automatic Gateway creation
      # We manage Gateway resources via ArgoCD/Kustomize
      gateway = {
        enabled = false
      }

      # Disable automatic GatewayClass creation
      # We manage GatewayClass via ArgoCD/Kustomize
      gatewayClass = {
        enabled = false
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

      # Providers configuration
      providers = {
        kubernetesIngress = {
          enabled = true
        }
        kubernetesCRD = {
          enabled = true
        }
        kubernetesGateway = {
          enabled = true
        }
      }

      # RBAC configuration (required for Gateway API)
      rbac = {
        enabled    = true
        namespaced = false
      }

      # Additional configuration
      additionalArguments = [
        "--providers.kubernetesingress.ingressclass=traefik",
        "--providers.kubernetescrd",
        "--providers.kubernetesgateway",
      ]

      # Security - add NET_BIND_SERVICE to bind to ports 80/443 as non-root
      securityContext = {
        capabilities = {
          drop = ["ALL"]
          add  = ["NET_BIND_SERVICE"]
        }
        readOnlyRootFilesystem = true
        runAsNonRoot           = true
        runAsUser              = 65532
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

# Additional ClusterRole rules for Gateway API
# This patches the Traefik ClusterRole to add Gateway API permissions
resource "kubernetes_cluster_role_v1" "traefik_gateway_api" {
  metadata {
    name = "traefik-gateway-api-extension"
    labels = {
      "app.kubernetes.io/name"       = "traefik"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }

  rule {
    api_groups = ["gateway.networking.k8s.io"]
    resources = [
      "gatewayclasses",
      "gateways",
      "httproutes",
      "grpcroutes",
      "tcproutes",
      "tlsroutes",
      "referencegrants",
      "backendtlspolicies"
    ]
    verbs = ["get", "list", "watch"]
  }

  rule {
    api_groups = ["gateway.networking.k8s.io"]
    resources = [
      "gatewayclasses/status",
      "gateways/status",
      "httproutes/status",
      "grpcroutes/status",
      "tcproutes/status",
      "tlsroutes/status",
      "backendtlspolicies/status"
    ]
    verbs = ["update"]
  }

  depends_on = [helm_release.traefik]
}

# Bind the Gateway API permissions to Traefik service account
resource "kubernetes_cluster_role_binding_v1" "traefik_gateway_api" {
  metadata {
    name = "traefik-gateway-api-extension"
    labels = {
      "app.kubernetes.io/name"       = "traefik"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role_v1.traefik_gateway_api.metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = "traefik"
    namespace = var.namespace
  }

  depends_on = [helm_release.traefik]
}

# Data source to get Traefik service details (including LoadBalancer IP)
data "kubernetes_service" "traefik" {
  metadata {
    name      = helm_release.traefik.name
    namespace = var.namespace
  }

  depends_on = [helm_release.traefik]
}
