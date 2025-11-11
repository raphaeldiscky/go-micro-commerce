# Cluster information outputs

output "cluster_name" {
  description = "Name of the Kubernetes cluster"
  value       = var.cluster_name
}

output "cluster_endpoint" {
  description = "Kubernetes API server endpoint"
  value       = var.cluster_endpoint
}

# Talos client configuration
output "talosconfig" {
  description = "Talos client configuration for talosctl"
  value       = data.talos_client_configuration.this.talos_config
  sensitive   = true
}

output "talos_endpoints" {
  description = "Talos API endpoints (control plane nodes)"
  value       = [for node in var.control_plane_nodes : node.ip]
}

# Kubernetes configuration
output "kubeconfig" {
  description = "Kubernetes kubeconfig for kubectl"
  value       = talos_cluster_kubeconfig.this.kubeconfig_raw
  sensitive   = true
}

output "kubeconfig_file" {
  description = "Path to the kubeconfig file"
  value       = local_sensitive_file.kubeconfig.filename
}

output "talosconfig_file" {
  description = "Path to the talosconfig file"
  value       = local_sensitive_file.talosconfig.filename
}

output "kubernetes_ca_certificate" {
  description = "Kubernetes CA certificate"
  value       = talos_cluster_kubeconfig.this.kubernetes_client_configuration.ca_certificate
  sensitive   = true
}

output "kubernetes_client_certificate" {
  description = "Kubernetes client certificate"
  value       = talos_cluster_kubeconfig.this.kubernetes_client_configuration.client_certificate
  sensitive   = true
}

output "kubernetes_client_key" {
  description = "Kubernetes client key"
  value       = talos_cluster_kubeconfig.this.kubernetes_client_configuration.client_key
  sensitive   = true
}

output "kubernetes_host" {
  description = "Kubernetes API server host"
  value       = talos_cluster_kubeconfig.this.kubernetes_client_configuration.host
}

# Node information
output "control_plane_nodes" {
  description = "Control plane node information"
  value = {
    for node in var.control_plane_nodes : node.name => {
      ip   = node.ip
      role = "control-plane"
    }
  }
}

output "worker_nodes" {
  description = "Worker node information"
  value = {
    for node in var.worker_nodes : node.name => {
      ip   = node.ip
      role = "worker"
    }
  }
}

# Cluster health check command
output "health_check_command" {
  description = "Command to check cluster health"
  value       = "kubectl --kubeconfig=${local_sensitive_file.kubeconfig.filename} get nodes"
}

# Talos commands
output "talos_commands" {
  description = "Useful talosctl commands"
  value = {
    get_nodes          = "talosctl --talosconfig=<(echo '${data.talos_client_configuration.this.talos_config}') get members"
    get_kubernetes     = "talosctl --talosconfig=<(echo '${data.talos_client_configuration.this.talos_config}') kubeconfig"
    dashboard          = "talosctl --talosconfig=<(echo '${data.talos_client_configuration.this.talos_config}') dashboard"
    health             = "talosctl --talosconfig=<(echo '${data.talos_client_configuration.this.talos_config}') health"
  }
  sensitive = true
}

output "external_secrets_namespace" {
  value       = local.external_secrets_namespace
  description = "Namespace where External Secrets Operator is deployed"
}

output "external_secret_store_name" {
  value       = "gcp-secret-manager"
  description = "Name of the created ClusterSecretStore"
}
