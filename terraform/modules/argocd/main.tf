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
        domain = var.domain_name
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

        # Service configuration
        service = {
          type = "ClusterIP"
        }

        # Ingress configuration for external access
        ingress = {
          enabled     = var.enable_ingress
          ingressClassName = "traefik"
          hosts = [
            var.domain_name
          ]
          tls = [
            {
              secretName = "argocd-server-tls"
              hosts = [
                var.domain_name
              ]
            }
          ]
          annotations = {
            "cert-manager.io/cluster-issuer" = var.tls_issuer
            "traefik.ingress.kubernetes.io/router.entrypoints" = "websecure"
          }
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
            memory = "2Gi"
          }
          requests = {
            cpu    = "250m"
            memory = "1Gi"
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

# Wait for External Secrets CRDs to be available
# Required before deploying ExternalSecret for ArgoCD Git credentials
resource "null_resource" "wait_for_eso_crd" {
  provisioner "local-exec" {
    command = <<EOF
      for i in {1..60}; do
        if kubectl get crd externalsecrets.external-secrets.io 2>/dev/null; then
          echo "External Secrets CRDs are ready"
          exit 0
        fi
        echo "Waiting for External Secrets CRDs... ($i/60)"
        sleep 5
      done
      echo "Timeout waiting for External Secrets CRDs"
      exit 1
    EOF
  }

  depends_on = [helm_release.argocd]
}

# Bootstrap ArgoCD Git Credentials
# This ExternalSecret is deployed via Terraform to bootstrap ArgoCD's ability to sync from Git
# Once ArgoCD is functional, it will take over management via the external-secrets Application
resource "kubectl_manifest" "argocd_git_credentials" {
  yaml_body = yamlencode({
    apiVersion = "external-secrets.io/v1"
    kind       = "ExternalSecret"
    metadata = {
      name      = "argocd-git-credentials"
      namespace = var.namespace
      labels = {
        "app.kubernetes.io/name"       = "argocd"
        "app.kubernetes.io/component"  = "repo-credentials"
        "app.kubernetes.io/managed-by" = "terraform"
        "app.kubernetes.io/part-of"    = "go-micro-commerce"
        "environment"                  = "production"
      }
      annotations = {
        "argocd.argoproj.io/sync-options" = "Prune=false"
      }
    }
    spec = {
      refreshInterval = "1h"
      secretStoreRef = {
        name = var.cluster_secret_store_name
        kind = "ClusterSecretStore"
      }
      target = {
        name           = "git-repo-credentials"
        creationPolicy = "Owner"
        deletionPolicy = "Retain"
        template = {
          type          = "Opaque"
          engineVersion = "v2"
          metadata = {
            labels = {
              "argocd.argoproj.io/secret-type" = "repository"
            }
          }
          data = {
            type          = "git"
            url           = var.git_repo_url
            sshPrivateKey = "{{ .sshPrivateKey }}"
          }
        }
      }
      data = [
        {
          secretKey = "sshPrivateKey"
          remoteRef = {
            key = "argocd-git-ssh-private-key"
          }
        }
      ]
    }
  })

  depends_on = [
    null_resource.wait_for_eso_crd,
    kubernetes_namespace.argocd
  ]
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
# This ApplicationSet deploys all ApplicationSet manifests from git
# which in turn deploy infrastructure and workloads
resource "kubectl_manifest" "bootstrap_appset" {
  count = var.enable_bootstrap && var.git_repo_url != "" ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "ApplicationSet"
    metadata = {
      name      = "bootstrap-applicationsets"
      namespace = var.namespace
      labels = {
        "app.kubernetes.io/managed-by" = "terraform"
        "app.kubernetes.io/component"  = "bootstrap"
      }
    }
    spec = {
      goTemplate = true
      goTemplateOptions = ["missingkey=error"]
      generators = [
        {
          list = {
            elements = [
              {
                name      = "argocd-projects"
                path      = "${var.git_repo_path}/apps/projects"
                syncWave  = "-1"
              },
              {
                name      = "applicationsets"
                path      = "${var.git_repo_path}/apps/applicationsets"
                syncWave  = "0"
              }
            ]
          }
        }
      ]
      template = {
        metadata = {
          name = "{{.name}}"
          labels = {
            "app.kubernetes.io/managed-by" = "argocd"
            "app.kubernetes.io/component"  = "applicationset"
          }
          annotations = {
            "argocd.argoproj.io/sync-wave" = "{{.syncWave}}"
          }
        }
        spec = {
          project = "default"
          source = {
            repoURL        = var.git_repo_url
            targetRevision = "HEAD"
            path           = "{{.path}}"
            directory = {
              recurse = false
            }
          }
          destination = {
            server    = "https://kubernetes.default.svc"
            namespace = var.namespace
          }
          syncPolicy = {
            automated = {
              prune    = true
              selfHeal = true
              allowEmpty = false
            }
            syncOptions = [
              "CreateNamespace=true",
              "PrunePropagationPolicy=foreground"
            ]
            retry = {
              limit = 5
              backoff = {
                duration    = "5s"
                factor      = 2
                maxDuration = "3m"
              }
            }
          }
        }
      }
    }
  })

  depends_on = [
    null_resource.wait_for_argocd_crd,
    kubectl_manifest.argocd_git_credentials
  ]
}
