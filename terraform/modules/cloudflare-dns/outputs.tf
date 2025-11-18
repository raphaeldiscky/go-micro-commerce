# Cloudflare DNS Module Outputs

output "zone_id" {
  description = "The Cloudflare zone ID"
  value       = var.cloudflare_zone_id
}

output "zone_name" {
  description = "The Cloudflare zone name"
  value       = var.domain_name
}

output "api_record_id" {
  description = "The ID of the API DNS record"
  value       = cloudflare_dns_record.api.id
}

output "api_record_hostname" {
  description = "The full hostname of the API"
  value       = cloudflare_dns_record.api.name
}

output "api_record_proxied" {
  description = "Whether the API record is proxied (should be false)"
  value       = cloudflare_dns_record.api.proxied
}

output "api_wildcard_record_id" {
  description = "The ID of the API wildcard DNS record (if enabled)"
  value       = var.enable_api_wildcard ? cloudflare_dns_record.api_wildcard[0].id : null
}

output "api_wildcard_record_hostname" {
  description = "The full hostname of the API wildcard (if enabled)"
  value       = var.enable_api_wildcard ? cloudflare_dns_record.api_wildcard[0].name : null
}

output "argocd_record_id" {
  description = "The ID of the ArgoCD DNS record (if enabled)"
  value       = var.enable_argocd_dns ? cloudflare_dns_record.argocd[0].id : null
}

output "argocd_record_hostname" {
  description = "The full hostname of ArgoCD (if enabled)"
  value       = var.enable_argocd_dns ? cloudflare_dns_record.argocd[0].name : null
}

output "argocd_url" {
  description = "The full HTTPS URL for ArgoCD (if enabled)"
  value       = var.enable_argocd_dns ? "https://${cloudflare_dns_record.argocd[0].name}" : null
}

output "grafana_record_id" {
  description = "The ID of the Grafana DNS record (if enabled)"
  value       = var.enable_grafana_dns ? cloudflare_dns_record.grafana[0].id : null
}

output "grafana_record_hostname" {
  description = "The full hostname of Grafana (if enabled)"
  value       = var.enable_grafana_dns ? cloudflare_dns_record.grafana[0].name : null
}

output "grafana_url" {
  description = "The full HTTPS URL for Grafana (if enabled)"
  value       = var.enable_grafana_dns ? "https://${cloudflare_dns_record.grafana[0].name}" : null
}

output "dns_records" {
  description = "Summary of all created DNS records"
  value = {
    api = {
      hostname = cloudflare_dns_record.api.name
      type     = cloudflare_dns_record.api.type
      content  = cloudflare_dns_record.api.content
      proxied  = cloudflare_dns_record.api.proxied
    }
    api_wildcard = var.enable_api_wildcard ? {
      hostname = cloudflare_dns_record.api_wildcard[0].name
      type     = cloudflare_dns_record.api_wildcard[0].type
      content  = cloudflare_dns_record.api_wildcard[0].content
      proxied  = cloudflare_dns_record.api_wildcard[0].proxied
    } : null
    argocd = var.enable_argocd_dns ? {
      hostname = cloudflare_dns_record.argocd[0].name
      type     = cloudflare_dns_record.argocd[0].type
      content  = cloudflare_dns_record.argocd[0].content
      proxied  = cloudflare_dns_record.argocd[0].proxied
      url      = "https://${cloudflare_dns_record.argocd[0].name}"
    } : null
    grafana = var.enable_grafana_dns ? {
      hostname = cloudflare_dns_record.grafana[0].name
      type     = cloudflare_dns_record.grafana[0].type
      content  = cloudflare_dns_record.grafana[0].content
      proxied  = cloudflare_dns_record.grafana[0].proxied
      url      = "https://${cloudflare_dns_record.grafana[0].name}"
    } : null
  }
}
