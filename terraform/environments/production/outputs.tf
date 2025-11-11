# Production environment outputs

output "cluster_info" {
  description = "Cluster information"
  value = {
    name              = module.talos_cluster.cluster_name
    endpoint          = module.talos_cluster.cluster_endpoint
    kubernetes_host   = module.talos_cluster.kubernetes_host
    talos_endpoints   = module.talos_cluster.talos_endpoints
  }
}

output "kubeconfig_file" {
  description = "Path to kubeconfig file"
  value       = module.talos_cluster.kubeconfig_file
}

output "talosconfig_file" {
  description = "Path to talosconfig file"
  value       = module.talos_cluster.talosconfig_file
}

output "nodes" {
  description = "Cluster nodes"
  value = {
    control_plane = module.talos_cluster.control_plane_nodes
    workers       = module.talos_cluster.worker_nodes
  }
}

output "health_check_command" {
  description = "Command to check cluster health"
  value       = module.talos_cluster.health_check_command
}

output "next_steps" {
  description = "Next steps after cluster creation"
  value       = <<-EOT
    Cluster created successfully!

    1. Export kubeconfig:
       export KUBECONFIG=${module.talos_cluster.kubeconfig_file}

    2. Check cluster health:
       ${module.talos_cluster.health_check_command}

    3. Install CNI (Cilium recommended):
       helm repo add cilium https://helm.cilium.io/
       helm install cilium cilium/cilium --namespace kube-system \
         --set ipam.mode=kubernetes \
         --set kubeProxyReplacement=true

    4. Install Longhorn:
       cd ../../../terraform/environments/production
       terraform apply -target=module.longhorn

    5. Install External Secrets Operator:
       terraform apply -target=module.external_secrets

    6. Install ArgoCD:
       terraform apply -target=module.argocd

    7. Deploy applications via ArgoCD:
       Follow ArgoCD setup in argocd/ directory
  EOT
}
