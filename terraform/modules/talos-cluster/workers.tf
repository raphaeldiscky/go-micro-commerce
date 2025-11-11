# Generate machine configuration for worker nodes
data "talos_machine_configuration" "worker" {
  cluster_name     = var.cluster_name
  machine_type     = "worker"
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
            [for node in var.worker_nodes : node.ip],           # Internal IPs
            var.worker_external_ips,                             # External IPs
            ["127.0.0.1"]                                        # Localhost
          )

          install = {
            disk  = var.install_disk
            image = var.install_image != "" ? var.install_image : null
          }

          # Network configuration
          network = {
            nameservers = ["1.1.1.1", "8.8.8.8"]
          }

          # Kernel modules required for Longhorn
          kernel = {
            modules = [
              {
                name = "nbd"
              },
              {
                name = "iscsi_tcp"
              },
              {
                name = "configfs"
              }
            ]
          }

          # Kubelet configuration with Longhorn support
          kubelet = var.enable_longhorn ? {
            # Longhorn requires access to host directories
            extraMounts = [
              {
                destination = var.longhorn_disk_path
                type        = "bind"
                source      = var.longhorn_disk_path
                options = [
                  "bind",
                  "rshared",
                  "rw"
                ]
              }
            ]

            nodeIP = {
              validSubnets = ["0.0.0.0/0"]
            }
          } : {
            extraMounts = []
            nodeIP = {
              validSubnets = ["0.0.0.0/0"]
            }
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
        }

        # Cluster configuration
        cluster = {
          network = {
            cni = {
              name = "none" # We'll install CNI separately
            }
          }
        }
      })
    ],
    var.worker_patches
  )
}

# Apply machine configuration to each worker node
resource "talos_machine_configuration_apply" "worker" {
  for_each = { for idx, node in var.worker_nodes : node.name => node }

  client_configuration        = talos_machine_secrets.this.client_configuration
  machine_configuration_input = data.talos_machine_configuration.worker.machine_configuration
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

        # Node labels for scheduling
        nodeLabels = merge(
          {
            "node-role.kubernetes.io/worker" = ""
            "environment"                     = var.environment
            "storage"                         = var.enable_longhorn ? "longhorn" : "none"
            "cloud-provider"                  = "gcp"
          },
          var.tags
        )

        # Node taints (optional - for dedicated workload nodes)
        # nodeTaints = {
        #   "workload=database:NoSchedule" = ""
        # }
      }
    })
  ]
  # Ensure control plane is configured first
  depends_on = [
    talos_machine_configuration_apply.control_plane
  ]
}
