# ArgoCD Module Outputs

output "namespace" {
  description = "Namespace where ArgoCD is installed"
  value       = var.namespace
}

output "release_name" {
  description = "Helm release name"
  value       = helm_release.argocd.name
}

output "chart_version" {
  description = "Installed chart version"
  value       = helm_release.argocd.version
}

output "status" {
  description = "Release status"
  value       = helm_release.argocd.status
}

output "argocd_url" {
  description = "URL to access ArgoCD"
  value       = var.enable_ingress ? "https://${var.domain_name}" : "http://localhost:8080 (kubectl port-forward -n ${var.namespace} svc/argocd-server 8080:80)"
}

output "argocd_domain" {
  description = "Domain name for ArgoCD"
  value       = var.domain_name
}

output "ingress_enabled" {
  description = "Whether ingress is enabled"
  value       = var.enable_ingress
}

output "argocd_admin_user" {
  description = "ArgoCD admin username"
  value       = "admin"
}

output "argocd_password_command" {
  description = "Command to get ArgoCD admin password"
  value       = "kubectl -n ${var.namespace} get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d"
}
