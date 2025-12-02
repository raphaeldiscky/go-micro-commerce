# Cloudflare DNS Module Variables

variable "cloudflare_zone_id" {
  description = "Cloudflare Zone ID for the domain"
  type        = string
}

variable "domain_name" {
  description = "The domain name managed in Cloudflare (e.g., discky.com)"
  type        = string
}

variable "api_subdomain" {
  description = "Subdomain for backend API (e.g., api)"
  type        = string
}

variable "enable_api_wildcard" {
  description = "Enable wildcard DNS record for API subdomains (e.g., *.api.discky.com)"
  type        = bool
  default     = false
}

variable "enable_argocd_dns" {
  description = "Enable DNS record for ArgoCD web UI"
  type        = bool
  default     = true
}

variable "argocd_subdomain" {
  description = "Full subdomain for ArgoCD (e.g., argocd.api or just argocd)"
  type        = string
  default     = "argocd.api"
}

variable "enable_grafana_dns" {
  description = "Enable DNS record for Grafana monitoring UI"
  type        = bool
  default     = true
}

variable "grafana_subdomain" {
  description = "Full subdomain for Grafana (e.g., grafana.api or just grafana)"
  type        = string
  default     = "grafana.api"
}

variable "enable_traefik_dns" {
  description = "Enable DNS record for Traefik dashboard UI"
  type        = bool
  default     = true
}

variable "traefik_subdomain" {
  description = "Full subdomain for Traefik dashboard (e.g., traefik)"
  type        = string
  default     = "traefik"
}

variable "traefik_ip" {
  description = "External IP address of Traefik LoadBalancer in GKE"
  type        = string
}
