# cert-manager Module
# Installs cert-manager for automated TLS certificate management with Let's Encrypt

# Create namespace for cert-manager
resource "kubernetes_namespace" "cert_manager" {
  count = var.create_namespace ? 1 : 0

  metadata {
    name = var.namespace

    labels = {
      "app.kubernetes.io/name"       = "cert-manager"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

# Install cert-manager via Helm
resource "helm_release" "cert_manager" {
  name       = "cert-manager"
  namespace  = var.namespace
  repository = "https://charts.jetstack.io"
  chart      = "cert-manager"
  version    = var.chart_version

  create_namespace = false

  values = [
    yamlencode({
      # Install CRDs
      installCRDs = true

      # Enable Prometheus metrics
      prometheus = {
        enabled = var.enable_monitoring
        servicemonitor = {
          enabled = var.enable_monitoring
        }
      }

      # Controller configuration
      replicaCount = var.replicas

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

      # Webhook configuration
      webhook = {
        replicaCount = 1
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

      # CA injector configuration
      cainjector = {
        replicaCount = 1
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
    })
  ]

  depends_on = [kubernetes_namespace.cert_manager]
}

# Wait for cert-manager CRDs to be registered
resource "null_resource" "wait_for_cert_manager_crd" {
  provisioner "local-exec" {
    command = <<EOF
      for i in {1..60}; do
        if kubectl get crd clusterissuers.cert-manager.io 2>/dev/null; then
          echo "cert-manager CRDs are ready"
          exit 0
        fi
        echo "Waiting for cert-manager CRDs... ($i/60)"
        sleep 5
      done
      echo "Timeout waiting for cert-manager CRDs"
      exit 1
    EOF
  }

  depends_on = [helm_release.cert_manager]
}

# Let's Encrypt Staging ClusterIssuer (for testing)
resource "kubectl_manifest" "letsencrypt_staging" {
  count = var.create_cluster_issuers ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "cert-manager.io/v1"
    kind       = "ClusterIssuer"
    metadata = {
      name = var.letsencrypt_staging_issuer_name
      labels = {
        "app.kubernetes.io/managed-by" = "terraform"
      }
    }
    spec = {
      acme = {
        # Let's Encrypt staging server
        server = "https://acme-staging-v02.api.letsencrypt.org/directory"
        email  = var.letsencrypt_email

        # Store account key in secret
        privateKeySecretRef = {
          name = "letsencrypt-staging"
        }

        # HTTP-01 challenge solver
        solvers = [
          {
            http01 = {
              ingress = {
                class = "traefik"
              }
            }
          }
        ]
      }
    }
  })

  depends_on = [null_resource.wait_for_cert_manager_crd]
}

# Let's Encrypt Production ClusterIssuer
resource "kubectl_manifest" "letsencrypt_prod" {
  count = var.create_cluster_issuers ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "cert-manager.io/v1"
    kind       = "ClusterIssuer"
    metadata = {
      name = var.letsencrypt_prod_issuer_name
      labels = {
        "app.kubernetes.io/managed-by" = "terraform"
      }
    }
    spec = {
      acme = {
        # Let's Encrypt production server
        server = "https://acme-v02.api.letsencrypt.org/directory"
        email  = var.letsencrypt_email

        # Store account key in secret
        privateKeySecretRef = {
          name = "letsencrypt-prod"
        }

        # HTTP-01 challenge solver
        solvers = [
          {
            http01 = {
              ingress = {
                class = "traefik"
              }
            }
          }
        ]
      }
    }
  })

  depends_on = [null_resource.wait_for_cert_manager_crd]
}
