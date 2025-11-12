locals {
  external_secrets_namespace = "external-secrets-system"
}

# Install External Secrets Operator via Helm
resource "helm_release" "external_secrets" {
  count = var.enable_external_secrets ? 1 : 0

  name             = "external-secrets"
  repository       = "https://charts.external-secrets.io"
  chart            = "external-secrets"
  version          = "1.0.0"
  namespace        = local.external_secrets_namespace
  create_namespace = true

  wait          = true
  wait_for_jobs = true
  timeout       = 600

  values = [
    yamlencode({
      installCRDs  = true
      replicaCount = 1

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

      serviceAccount = {
        create = true
        name   = "external-secrets-sa"
      }
    })
  ]

  depends_on = [
    talos_cluster_kubeconfig.this
  ]
}

# Wait until the ClusterSecretStore CRD exists
resource "null_resource" "wait_for_external_secret_crd" {
  count = var.enable_external_secrets ? 1 : 0

  provisioner "local-exec" {
    command = <<-EOT
      echo "Waiting for ClusterSecretStore CRD registration..."
      for i in {1..60}; do
        if kubectl get crd clustersecretstores.external-secrets.io &> /dev/null; then
          echo "ClusterSecretStore CRD is available!"
          exit 0
        fi
        echo "Attempt $i/60: CRD not ready yet, waiting 5s..."
        sleep 5
      done
      echo "CRD not available after 5 minutes."
      exit 1
    EOT

    environment = {
      KUBECONFIG = local_sensitive_file.kubeconfig.filename
    }
  }

  depends_on = [
    helm_release.external_secrets
  ]
}

# Create Secret with GCP Service Account key
resource "kubernetes_secret" "gcp_sa_key" {
  count = var.enable_external_secrets && var.gcp_sa_key_path != "" ? 1 : 0

  metadata {
    name      = "gcp-sa-secret"
    namespace = local.external_secrets_namespace
  }

  data = {
    "credentials.json" = filebase64(var.gcp_sa_key_path)
  }

  type = "Opaque"

  depends_on = [
    null_resource.wait_for_external_secret_crd
  ]
}

# Create ClusterSecretStore for GCP Secret Manager
resource "kubernetes_manifest" "cluster_secret_store_gcp" {
  count = var.enable_external_secrets && var.gcp_project_id != "" ? 1 : 0

  manifest = {
    apiVersion = "external-secrets.io/v1beta1"
    kind       = "ClusterSecretStore"
    metadata = {
      name = "gcp-secret-manager"
    }
    spec = {
      provider = {
        gcpsm = {
          projectID = var.gcp_project_id
          auth = {
            secretRef = {
              secretAccessKeySecretRef = {
                name      = "gcp-sa-secret"
                key       = "credentials.json"
                namespace = local.external_secrets_namespace
              }
            }
          }
        }
      }
    }
  }

  # Skip validation during plan phase (CRDs are installed by Helm during apply)
  computed_fields = ["metadata", "spec"]

  depends_on = [
    null_resource.wait_for_external_secret_crd,
    kubernetes_secret.gcp_sa_key
  ]
}

# Wait until ClusterSecretStore resource is ready
resource "null_resource" "wait_for_cluster_secret_store" {
  count = var.enable_external_secrets ? 1 : 0

  provisioner "local-exec" {
    command = <<-EOT
      echo "Waiting for ClusterSecretStore to be ready..."
      for i in {1..60}; do
        if kubectl get clustersecretstore gcp-secret-manager &> /dev/null; then
          echo "ClusterSecretStore is ready!"
          exit 0
        fi
        echo "Attempt $i/60: Not ready yet, waiting 5s..."
        sleep 5
      done
      echo "ClusterSecretStore did not become ready in time."
      exit 1
    EOT

    environment = {
      KUBECONFIG = local_sensitive_file.kubeconfig.filename
    }
  }

  depends_on = [
    kubernetes_manifest.cluster_secret_store_gcp
  ]
}
