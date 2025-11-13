# Traefik Module Outputs

output "namespace" {
  description = "Namespace where Traefik is installed"
  value       = var.namespace
}

output "release_name" {
  description = "Helm release name"
  value       = helm_release.traefik.name
}

output "chart_version" {
  description = "Installed chart version"
  value       = helm_release.traefik.version
}

output "status" {
  description = "Release status"
  value       = helm_release.traefik.status
}

output "dashboard_url" {
  description = "URL to access Traefik dashboard (if enabled)"
  value       = var.enable_dashboard ? "http://localhost:9000/dashboard/ (kubectl port-forward -n ${var.namespace} $(kubectl get pods -n ${var.namespace} -l app.kubernetes.io/name=traefik -o name | head -n1) 9000:9000)" : "Dashboard disabled"
}

output "service_type" {
  description = "Service type for Traefik"
  value       = var.service_type
}

output "load_balancer_ip" {
  description = "External IP address of the Traefik LoadBalancer service"
  value       = var.service_type == "LoadBalancer" ? try(data.kubernetes_service.traefik.status[0].load_balancer[0].ingress[0].ip, "") : ""
}

output "load_balancer_hostname" {
  description = "External hostname of the Traefik LoadBalancer service (for cloud providers that use hostnames)"
  value       = var.service_type == "LoadBalancer" ? try(data.kubernetes_service.traefik.status[0].load_balancer[0].ingress[0].hostname, "") : ""
}
