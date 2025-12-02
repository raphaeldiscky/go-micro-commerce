# Cloudflare DNS Module
# Manages DNS records for backend API only
# Note: Frontend (go.micro.commerce.discky.com) is managed by Cloudflare Pages

# Backend API A record (DNS only, no proxy)
# Points directly to Traefik LoadBalancer in GKE
resource "cloudflare_dns_record" "api" {
  zone_id = var.cloudflare_zone_id
  name    = var.api_subdomain
  type    = "A"
  content = var.traefik_ip
  ttl     = 300  # 5 minutes TTL
  proxied = false  # Gray cloud - direct to origin (no CDN)
  comment = "Backend API endpoints in GKE cluster"
}

# Optional wildcard API A record for service subdomains
# e.g., product.api.discky.com, order.api.discky.com
resource "cloudflare_dns_record" "api_wildcard" {
  count = var.enable_api_wildcard ? 1 : 0

  zone_id = var.cloudflare_zone_id
  name    = "*.${var.api_subdomain}"
  type    = "A"
  content = var.traefik_ip
  ttl     = 300
  proxied = false
  comment = "Backend API service subdomains"
}

# ArgoCD A record (DNS only, no proxy)
# Points directly to Traefik LoadBalancer in GKE
resource "cloudflare_dns_record" "argocd" {
  count = var.enable_argocd_dns ? 1 : 0

  zone_id = var.cloudflare_zone_id
  name    = var.argocd_subdomain
  type    = "A"
  content = var.traefik_ip
  ttl     = 300  # 5 minutes TTL
  proxied = false  # Gray cloud - direct to origin (no CDN)
  comment = "ArgoCD web UI in GKE cluster"
}

# Grafana A record (DNS only, no proxy)
# Points directly to Traefik LoadBalancer in GKE
resource "cloudflare_dns_record" "grafana" {
  count = var.enable_grafana_dns ? 1 : 0

  zone_id = var.cloudflare_zone_id
  name    = var.grafana_subdomain
  type    = "A"
  content = var.traefik_ip
  ttl     = 300  # 5 minutes TTL
  proxied = false  # Gray cloud - direct to origin (no CDN)
  comment = "Grafana monitoring UI in GKE cluster"
}

# Traefik Dashboard A record (DNS only, no proxy)
# Points directly to Traefik LoadBalancer in GKE
resource "cloudflare_dns_record" "traefik" {
  count = var.enable_traefik_dns ? 1 : 0

  zone_id = var.cloudflare_zone_id
  name    = var.traefik_subdomain
  type    = "A"
  content = var.traefik_ip
  ttl     = 300  # 5 minutes TTL
  proxied = false  # Gray cloud - direct to origin (no CDN)
  comment = "Traefik dashboard UI in GKE cluster"
}
