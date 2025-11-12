# Monitoring Module Outputs

output "namespace" {
  description = "Namespace where monitoring stack is installed"
  value       = var.namespace
}

output "prometheus_release_name" {
  description = "Helm release name for Prometheus"
  value       = helm_release.kube_prometheus_stack.name
}

output "loki_release_name" {
  description = "Helm release name for Loki"
  value       = helm_release.loki.name
}

output "tempo_release_name" {
  description = "Helm release name for Tempo"
  value       = helm_release.tempo.name
}

output "grafana_url" {
  description = "URL to access Grafana (port-forward required)"
  value       = "http://localhost:3000 (kubectl port-forward -n ${var.namespace} svc/kube-prometheus-stack-grafana 3000:80)"
}

output "grafana_admin_user" {
  description = "Grafana admin username"
  value       = "admin"
}

output "prometheus_url" {
  description = "URL to access Prometheus (port-forward required)"
  value       = "http://localhost:9090 (kubectl port-forward -n ${var.namespace} svc/kube-prometheus-stack-prometheus 9090:9090)"
}
