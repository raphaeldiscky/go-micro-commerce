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
            cpu    = "50m"    # Reduced from 100m
            memory = "96Mi"    # Reduced from 128Mi
          }
        }

        # Node affinity for control plane pool
        nodeSelector = {
          workload-type = "control-plane"
        }
        tolerations = [
          {
            key      = "workload-type"
            operator = "Equal"
            value    = "control-plane"
            effect   = "NoSchedule"
          }
        ]

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
        # Node affinity for control plane pool
        nodeSelector = {
          workload-type = "control-plane"
        }
        tolerations = [
          {
            key      = "workload-type"
            operator = "Equal"
            value    = "control-plane"
            effect   = "NoSchedule"
          }
        ]
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
            cpu    = "50m"    # Reduced from 100m
            memory = "96Mi"    # Reduced from 128Mi
          }
        }
        # Node affinity for control plane pool
        nodeSelector = {
          workload-type = "control-plane"
        }
        tolerations = [
          {
            key      = "workload-type"
            operator = "Equal"
            value    = "control-plane"
            effect   = "NoSchedule"
          }
        ]
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
            cpu    = "50m"    # Reduced from 100m
            memory = "96Mi"    # Reduced from 128Mi
          }
        }
        # Node affinity for control plane pool
        nodeSelector = {
          workload-type = "control-plane"
        }
        tolerations = [
          {
            key      = "workload-type"
            operator = "Equal"
            value    = "control-plane"
            effect   = "NoSchedule"
          }
        ]
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
        # Node affinity for control plane pool
        nodeSelector = {
          workload-type = "control-plane"
        }
        tolerations = [
          {
            key      = "workload-type"
            operator = "Equal"
            value    = "control-plane"
            effect   = "NoSchedule"
          }
        ]
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
        # Node affinity for control plane pool
        nodeSelector = {
          workload-type = "control-plane"
        }
        tolerations = [
          {
            key      = "workload-type"
            operator = "Equal"
            value    = "control-plane"
            effect   = "NoSchedule"
          }
        ]
      }

      # Dex server configuration (OAuth/SSO)
      dex = {
        enabled = true
        resources = {
          limits = {
            cpu    = "200m"
            memory = "256Mi"
          }
          requests = {
            cpu    = "50m"
            memory = "64Mi"
          }
        }
        # Node affinity for control plane pool
        nodeSelector = {
          workload-type = "control-plane"
        }
        tolerations = [
          {
            key      = "workload-type"
            operator = "Equal"
            value    = "control-plane"
            effect   = "NoSchedule"
          }
        ]
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
        secret = {
          # Dex server secret (required to prevent crash)
          "dex.server.secretkey" = "temporary-secret-key-change-in-production"
        }
      }
    })
  ]

  depends_on = [kubernetes_namespace.argocd]
}

# Wait for ArgoCD CRDs to be registered
resource "null_resource" "wait_for_argocd_crd" {
  count = var.enable_bootstrap && var.git_repo_url != "" ? 1 : 0

  provisioner "local-exec" {
    command = <<EOF
      for i in {1..60}; do
        if kubectl get crd applicationsets.argoproj.io 2>/dev/null; then
          echo "ArgoCD ApplicationSet CRD is ready"
          exit 0
        fi
        echo "Waiting for ArgoCD CRDs... ($i/60)"
        sleep 5
      done
      echo "Timeout waiting for ArgoCD CRDs"
      exit 1
    EOF
  }

  depends_on = [helm_release.argocd]
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

  depends_on = [null_resource.wait_for_argocd_crd]
}
