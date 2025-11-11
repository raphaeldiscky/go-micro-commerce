# Generate machine configuration for control plane nodes
data "talos_machine_configuration" "control_plane" {
  cluster_name     = var.cluster_name
  machine_type     = "controlplane"
  cluster_endpoint = var.cluster_endpoint
  machine_secrets  = talos_machine_secrets.this.machine_secrets
  talos_version    = var.talos_version
  kubernetes_version = var.kubernetes_version

  config_patches = concat(
    [
      # Install disk configuration
      yamlencode({
        machine = {
          # Talos API certificate SANs (includes both external and internal IPs)
          certSANs = concat(
            [for node in var.control_plane_nodes : node.ip],  # Internal IPs
            var.control_plane_external_ips,                    # External IPs
            ["127.0.0.1"]                                      # Localhost
          )

          install = {
            disk  = var.install_disk
            image = var.install_image != "" ? var.install_image : null
          }
          # System extensions for Longhorn (iSCSI support)
          # Note: If using custom image from Image Factory, extensions are baked in
          # Otherwise, uncomment below and specify extensions
          # install = {
          #   extensions = [
          #     {
          #       image = "ghcr.io/siderolabs/iscsi-tools:v0.1.4"
          #     },
          #     {
          #       image = "ghcr.io/siderolabs/util-linux-tools:2.40.2"
          #     }
          #   ]
          # }

          # Network configuration
          network = {
            nameservers = ["1.1.1.1", "8.8.8.8"]
          }

          # Kernel modules required for Longhorn
          kernel = {
            modules = [
              {
                name = "iscsi_tcp"
              }
            ]
          }

          # System disk encryption (optional - enable if needed)
          # systemDiskEncryption = {
          #   state = {
          #     provider = "luks2"
          #     keys = [{
          #       static = {
          #         passphrase = "changeme"
          #       }
          #     }]
          #   }
          # }

          # Kubelet configuration
          kubelet = {
            # Resource reservations for system
            nodeIP = {
              validSubnets = ["0.0.0.0/0"]
            }
          }
        }

        # Cluster configuration
        cluster = {
          network = {
            cni = {
              name = "none" # We'll install CNI separately (Cilium or Calico)
            }
            podSubnets     = [var.pod_cidr]
            serviceSubnets = [var.service_cidr]
          }

          # API server configuration
          apiServer = {
            certSANs = concat(
              [for node in var.control_plane_nodes : node.ip],
              var.control_plane_external_ips,  # External IPs
              [var.cluster_endpoint]
            )
          }

          # Allow workloads on control plane for small clusters (optional)
          # Uncomment if you want to use control plane for workloads
          # allowSchedulingOnControlPlanes = true
        }
      })
    ],
    var.control_plane_patches
  )
}

# Apply machine configuration to each control plane node
resource "talos_machine_configuration_apply" "control_plane" {
  for_each = { for idx, node in var.control_plane_nodes : node.name => node }

  client_configuration        = talos_machine_secrets.this.client_configuration
  machine_configuration_input = data.talos_machine_configuration.control_plane.machine_configuration
  node                        = each.value.ip
  endpoint                    = each.value.ip

  config_patches = [
    # Node-specific configuration (hostname, network)
    yamlencode({
      machine = {
        network = {
          hostname = each.value.name
          interfaces = [
            {
              interface = "eth0"
              dhcp      = true  # GCP uses DHCP for network configuration and metadata access
            }
          ]
        }

        # Node labels
        nodeLabels = merge(
          {
            "node-role.kubernetes.io/control-plane" = ""
            "environment"                            = var.environment
            "cloud-provider"                         = "gcp"
          },
          var.tags
        )
      }
    })
  ]
}
