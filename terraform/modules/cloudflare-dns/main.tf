# Cloudflare DNS Module
# Manages DNS records for backend API only
# Note: Frontend (go.micro.commerce.discky.com) is managed by Cloudflare Pages

# Lookup the Cloudflare zone by domain name
data "cloudflare_zone" "main" {
  name = var.domain_name
}

# Backend API A record (DNS only, no proxy)
# Points directly to Traefik LoadBalancer in GKE
resource "cloudflare_dns_record" "api" {
  zone_id = data.cloudflare_zone.main.id
  name    = var.api_subdomain
  type    = "A"
  content = var.traefik_ip
  ttl     = 300  # 5 minutes TTL
  proxied = false  # Gray cloud - direct to origin (no CDN)
  comment = "Backend API endpoints in GKE cluster"

  tags = ["terraform", "backend", "api", "gke"]
}

# Optional wildcard API A record for service subdomains
# e.g., product.api.discky.com, order.api.discky.com
resource "cloudflare_dns_record" "api_wildcard" {
  count = var.enable_api_wildcard ? 1 : 0

  zone_id = data.cloudflare_zone.main.id
  name    = "*.${var.api_subdomain}"
  type    = "A"
  content = var.traefik_ip
  ttl     = 300
  proxied = false
  comment = "Backend API service subdomains"

  tags = ["terraform", "backend", "api", "wildcard", "gke"]
}
