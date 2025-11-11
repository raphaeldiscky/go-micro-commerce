# Bootstrap the Talos cluster
# This initializes etcd on the first control plane node and forms the cluster
resource "talos_machine_bootstrap" "this" {
  client_configuration = talos_machine_secrets.this.client_configuration
  node                 = var.control_plane_nodes[0].ip
  endpoint             = var.control_plane_nodes[0].ip

  depends_on = [
    talos_machine_configuration_apply.control_plane,
    talos_machine_configuration_apply.worker
  ]
}

# Retrieve the cluster kubeconfig
resource "talos_cluster_kubeconfig" "this" {
  client_configuration = talos_machine_secrets.this.client_configuration
  node                 = var.control_plane_nodes[0].ip
  endpoint             = var.control_plane_nodes[0].ip

  depends_on = [
    talos_machine_bootstrap.this
  ]
}

# Save kubeconfig to local file for later use
resource "local_sensitive_file" "kubeconfig" {
  content         = talos_cluster_kubeconfig.this.kubeconfig_raw
  filename        = "${path.root}/kubeconfig-${var.environment}"
  file_permission = "0600"
}

# Wait for Kubernetes API to be available
resource "null_resource" "wait_for_api" {
  depends_on = [
    talos_cluster_kubeconfig.this
  ]

  provisioner "local-exec" {
    command = <<-EOT
      echo "Waiting for Kubernetes API to be ready..."
      for i in {1..60}; do
        if kubectl --kubeconfig=${local_sensitive_file.kubeconfig.filename} get nodes &> /dev/null; then
          echo "Kubernetes API is ready!"
          exit 0
        fi
        echo "Attempt $i/60: API not ready yet, waiting 10s..."
        sleep 10
      done
      echo "Kubernetes API did not become ready in time"
      exit 1
    EOT
  }
}
