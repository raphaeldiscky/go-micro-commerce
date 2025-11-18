# cert-manager Module Variables

variable "namespace" {
  description = "Kubernetes namespace for cert-manager"
  type        = string
  default     = "cert-manager"
}

variable "create_namespace" {
  description = "Whether to create the namespace"
  type        = bool
  default     = true
}

variable "chart_version" {
  description = "cert-manager Helm chart version"
  type        = string
  default     = "v1.19.1"
}

variable "replicas" {
  description = "Number of cert-manager controller replicas"
  type        = number
  default     = 1
}

variable "enable_monitoring" {
  description = "Enable Prometheus monitoring"
  type        = bool
  default     = true
}

variable "create_cluster_issuers" {
  description = "Whether to create Let's Encrypt ClusterIssuers"
  type        = bool
  default     = true
}

variable "letsencrypt_email" {
  description = "Email address for Let's Encrypt certificate notifications"
  type        = string
}

variable "letsencrypt_staging_issuer_name" {
  description = "Name for Let's Encrypt staging ClusterIssuer"
  type        = string
  default     = "letsencrypt-staging"
}

variable "letsencrypt_prod_issuer_name" {
  description = "Name for Let's Encrypt production ClusterIssuer"
  type        = string
  default     = "letsencrypt-prod"
}
