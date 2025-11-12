# Production Environment - Provider Configuration

# Kubernetes Provider - uses kubeconfig from gcloud
provider "kubernetes" {
  config_path = "~/.kube/config"
}

# Helm Provider - uses kubeconfig from gcloud
provider "helm" {
  kubernetes = {
    config_path = "~/.kube/config"
  }
}

# Kubectl Provider - uses kubeconfig from gcloud
provider "kubectl" {
  config_path      = "~/.kube/config"
  load_config_file = true
}
