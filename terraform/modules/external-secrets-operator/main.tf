# External Secrets Operator Module
# Installs External Secrets Operator for syncing secrets from Google Secret Manager

# Create namespace for External Secrets Operator
resource "kubernetes_namespace" "eso" {
  count = var.create_namespace ? 1 : 0

  metadata {
    name = var.namespace

    labels = {
      "app.kubernetes.io/name"       = "external-secrets"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

# Install External Secrets Operator via Helm
resource "helm_release" "external_secrets" {
  name       = "external-secrets"
  namespace  = var.namespace
  repository = "https://charts.external-secrets.io"
  chart      = "external-secrets"
  version    = var.chart_version

  create_namespace = false # Already created above

  values = [
    yamlencode({
      installCRDs  = true
      replicaCount = var.replicas

      webhook = {
        create = true
      }

      certController = {
        create = true
      }

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
        enabled = var.enable_monitoring
        service = {
          enabled = var.enable_monitoring
        }
      }
    })
  ]

  depends_on = [kubernetes_namespace.eso]
}

# GCP Service Account for External Secrets Operator
resource "google_service_account" "eso_sa" {
  account_id   = var.gcp_service_account_name
  display_name = "External Secrets Operator Service Account"
  description  = "Service account for External Secrets Operator to access Google Secret Manager"
  project      = var.project_id
}

# Grant Secret Manager Secret Accessor role to the service account
resource "google_project_iam_member" "eso_secret_accessor" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.eso_sa.email}"
}

# Workload Identity binding: Allow K8s SA to impersonate GCP SA
resource "google_service_account_iam_member" "eso_workload_identity" {
  service_account_id = google_service_account.eso_sa.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.project_id}.svc.id.goog[${var.namespace}/${var.k8s_service_account_name}]"
}

# Kubernetes Service Account for ESO
resource "kubernetes_service_account" "eso_sa" {
  metadata {
    name      = var.k8s_service_account_name
    namespace = var.namespace

    annotations = {
      "iam.gke.io/gcp-service-account" = google_service_account.eso_sa.email
    }

    labels = {
      "app.kubernetes.io/name"       = "external-secrets"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }

  depends_on = [
    kubernetes_namespace.eso,
    google_service_account_iam_member.eso_workload_identity
  ]
}

# Wait for External Secrets CRDs to be registered
resource "null_resource" "wait_for_eso_crd" {
  count = var.create_cluster_secret_store ? 1 : 0

  provisioner "local-exec" {
    command = <<EOF
      for i in {1..60}; do
        if kubectl get crd clustersecretstores.external-secrets.io 2>/dev/null; then
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

  depends_on = [helm_release.external_secrets]
}

# ClusterSecretStore for Google Secret Manager
resource "kubectl_manifest" "cluster_secret_store" {
  count = var.create_cluster_secret_store ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "external-secrets.io/v1"
    kind       = "ClusterSecretStore"
    metadata = {
      name = var.cluster_secret_store_name
      labels = {
        "app.kubernetes.io/managed-by" = "terraform"
      }
    }
    spec = {
      provider = {
        gcpsm = {
          projectID = var.project_id
          auth = {
            workloadIdentity = {
              clusterLocation = var.cluster_location
              clusterName     = var.cluster_name
              serviceAccountRef = {
                name      = kubernetes_service_account.eso_sa.metadata[0].name
                namespace = var.namespace
              }
            }
          }
        }
      }
    }
  })

  depends_on = [
    null_resource.wait_for_eso_crd,
    kubernetes_service_account.eso_sa,
    google_project_iam_member.eso_secret_accessor
  ]
}
