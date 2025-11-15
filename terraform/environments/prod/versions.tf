# Terraform and provider version constraints
# Copy this file to each environment directory or reference it via symlink

terraform {
  required_version = ">= 1.13.5"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 7.11"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.38"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 3.1.0"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = "~> 1.19"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.7"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.12"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.12.0"
    }
  }
}
