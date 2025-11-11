# Generate Talos machine secrets for the cluster
# These secrets are used for PKI, cluster authentication, and etcd encryption
resource "talos_machine_secrets" "this" {
  talos_version = var.talos_version
}

# Generate Talos client configuration for cluster management
# This is used by talosctl to communicate with the cluster
data "talos_client_configuration" "this" {
  cluster_name         = var.cluster_name
  client_configuration = talos_machine_secrets.this.client_configuration
  endpoints            = [for node in var.control_plane_nodes : node.ip]
  nodes                = [for node in var.control_plane_nodes : node.ip]
}
