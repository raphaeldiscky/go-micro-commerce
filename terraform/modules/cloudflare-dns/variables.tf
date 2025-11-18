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

variable "traefik_ip" {
  description = "External IP address of Traefik LoadBalancer in GKE"
  type        = string
}
