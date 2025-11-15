# External Secrets Operator Module Outputs

output "namespace" {
  description = "Namespace where External Secrets Operator is installed"
  value       = var.namespace
}

output "release_name" {
  description = "Helm release name"
  value       = helm_release.external_secrets.name
}

output "chart_version" {
  description = "Installed chart version"
  value       = helm_release.external_secrets.version
}

output "status" {
  description = "Helm release status"
  value       = helm_release.external_secrets.status
}

output "gcp_service_account_email" {
  description = "GCP service account email for Secret Manager access"
  value       = google_service_account.eso_sa.email
}

output "gcp_service_account_name" {
  description = "GCP service account name"
  value       = google_service_account.eso_sa.name
}

output "k8s_service_account_name" {
  description = "Kubernetes service account name"
  value       = kubernetes_service_account.eso_sa.metadata[0].name
}

output "cluster_secret_store_name" {
  description = "ClusterSecretStore name for referencing in ExternalSecrets"
  value       = var.create_cluster_secret_store ? var.cluster_secret_store_name : null
}
