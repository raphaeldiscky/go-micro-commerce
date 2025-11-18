# cert-manager Module Outputs

output "namespace" {
  description = "The namespace where cert-manager is deployed"
  value       = var.namespace
}

output "chart_version" {
  description = "The cert-manager Helm chart version"
  value       = var.chart_version
}

output "letsencrypt_staging_issuer" {
  description = "Name of the Let's Encrypt staging ClusterIssuer"
  value       = var.create_cluster_issuers ? var.letsencrypt_staging_issuer_name : null
}

output "letsencrypt_prod_issuer" {
  description = "Name of the Let's Encrypt production ClusterIssuer"
  value       = var.create_cluster_issuers ? var.letsencrypt_prod_issuer_name : null
}

output "status" {
  description = "cert-manager deployment status"
  value       = helm_release.cert_manager.status
}
