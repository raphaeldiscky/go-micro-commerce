# CloudNative PostgreSQL Operator Module Outputs

output "namespace" {
  description = "Namespace where the operator is installed"
  value       = var.namespace
}

output "release_name" {
  description = "Helm release name"
  value       = helm_release.cloudnative_pg.name
}

output "chart_version" {
  description = "Installed chart version"
  value       = helm_release.cloudnative_pg.version
}

output "status" {
  description = "Release status"
  value       = helm_release.cloudnative_pg.status
}
