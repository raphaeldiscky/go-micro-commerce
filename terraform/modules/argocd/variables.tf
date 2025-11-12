# ArgoCD Module Variables

variable "namespace" {
  description = "Kubernetes namespace for ArgoCD"
  type        = string
  default     = "argocd"
}

variable "create_namespace" {
  description = "Create the namespace if it doesn't exist"
  type        = bool
  default     = true
}

variable "chart_version" {
  description = "Helm chart version for ArgoCD"
  type        = string
  default     = "7.7.12"
}

variable "admin_password" {
  description = "Admin password for ArgoCD (leave empty for auto-generated)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "git_repo_url" {
  description = "Git repository URL for application manifests"
  type        = string
  default     = ""
}

variable "git_repo_path" {
  description = "Path in Git repository for application manifests"
  type        = string
  default     = "deployments/k8s"
}

variable "enable_bootstrap" {
  description = "Enable bootstrap ApplicationSet for microservices"
  type        = bool
  default     = true
}
