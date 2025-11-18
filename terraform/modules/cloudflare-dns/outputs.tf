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
  }
}
