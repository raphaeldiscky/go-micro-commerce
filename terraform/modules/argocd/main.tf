# ArgoCD Module
# Installs ArgoCD with bootstrap configuration for microservices

# Create namespace for ArgoCD
resource "kubernetes_namespace" "argocd" {
  count = var.create_namespace ? 1 : 0

  metadata {
    name = var.namespace

    labels = {
      "app.kubernetes.io/name"       = "argocd"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

# Install ArgoCD via Helm
resource "helm_release" "argocd" {
  name       = "argocd"
  namespace  = var.namespace
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  version    = var.chart_version

  create_namespace = false

  values = [
    yamlencode({
      # Global configuration
      global = {
        domain = "argocd.local" # Update this for production
      }

      # Server configuration
      server = {
        replicas = 1
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

        # Service configuration
        service = {
          type = "ClusterIP"
        }

        # Metrics
        metrics = {
          enabled = true
          serviceMonitor = {
            enabled = true
          }
        }
      }

      # Controller configuration
      controller = {
        replicas = 1
        resources = {
          limits = {
            cpu    = "1000m"
            memory = "1Gi"
          }
          requests = {
            cpu    = "250m"
            memory = "512Mi"
          }
        }
        metrics = {
          enabled = true
          serviceMonitor = {
            enabled = true
          }
        }
      }

      # Repo server configuration
      repoServer = {
        replicas = 1
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
        metrics = {
          enabled = true
          serviceMonitor = {
            enabled = true
          }
        }
      }

      # Application controller
      applicationSet = {
        enabled = true
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
      }

      # Notifications controller
      notifications = {
        enabled = true
        resources = {
          limits = {
            cpu    = "100m"
            memory = "128Mi"
          }
          requests = {
            cpu    = "50m"
            memory = "64Mi"
          }
        }
      }

      # Redis
      redis = {
        enabled = true
        resources = {
          limits = {
            cpu    = "200m"
            memory = "256Mi"
          }
          requests = {
            cpu    = "100m"
            memory = "128Mi"
          }
        }
      }

      # Configs
      configs = {
        params = {
          "server.insecure" = true # For local development
        }
        cm = {
          # Timeout settings
          "timeout.reconciliation" = "180s"
          "timeout.hard.reconciliation" = "0s"
        }
      }
    })
  ]

  depends_on = [kubernetes_namespace.argocd]
}

# Bootstrap ApplicationSet (optional - requires git repo URL)
resource "kubectl_manifest" "bootstrap_appset" {
  count = var.enable_bootstrap && var.git_repo_url != "" ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "ApplicationSet"
    metadata = {
      name      = "microservices-bootstrap"
      namespace = var.namespace
    }
    spec = {
      generators = [
        {
          git = {
            repoURL  = var.git_repo_url
            revision = "HEAD"
            directories = [
              {
                path = "${var.git_repo_path}/overlays/prod/*"
              }
            ]
          }
        }
      ]
      template = {
        metadata = {
          name = "{{path.basename}}"
        }
        spec = {
          project = "default"
          source = {
            repoURL        = var.git_repo_url
            targetRevision = "HEAD"
            path           = "{{path}}"
          }
          destination = {
            server    = "https://kubernetes.default.svc"
            namespace = "{{path.basename}}"
          }
          syncPolicy = {
            automated = {
              prune    = true
              selfHeal = true
            }
            syncOptions = [
              "CreateNamespace=true"
            ]
          }
        }
      }
    }
  })

  depends_on = [helm_release.argocd]
}
