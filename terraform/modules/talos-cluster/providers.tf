terraform {
  required_version = ">= 1.12.2"

  required_providers {
    talos = {
      source  = "siderolabs/talos"
      version = "~> 0.9.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 3.1.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.38.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.5.3"
    }
    null = {
      source  = "hashicorp/null"
      version = "~> 3.2.4"
    }
  }
}
