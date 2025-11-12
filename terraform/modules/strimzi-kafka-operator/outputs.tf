# Strimzi Kafka Operator Module Outputs

output "namespace" {
  description = "Namespace where the operator is installed"
  value       = var.namespace
}

output "release_name" {
  description = "Helm release name"
  value       = helm_release.strimzi.name
}

output "chart_version" {
  description = "Installed chart version"
  value       = helm_release.strimzi.version
}

output "status" {
  description = "Release status"
  value       = helm_release.strimzi.status
}

output "watch_namespaces" {
  description = "Namespaces being watched by the operator"
  value       = var.watch_namespaces
}
