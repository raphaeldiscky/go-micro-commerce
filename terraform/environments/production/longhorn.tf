# Longhorn distributed block storage for Kubernetes
# Provides persistent volume support across worker nodes

# Create namespace with privileged pod security
resource "kubernetes_namespace" "longhorn_system" {
  metadata {
    name = "longhorn-system"
    labels = {
      "pod-security.kubernetes.io/enforce" = "privileged"
      "pod-security.kubernetes.io/audit"   = "privileged"
      "pod-security.kubernetes.io/warn"    = "privileged"
    }
  }

  depends_on = [module.talos_cluster]
}

# Deploy Longhorn via Helm
resource "helm_release" "longhorn" {
  name       = "longhorn"
  repository = "https://charts.longhorn.io"
  chart      = "longhorn"
  version    = "1.10.0" 
  namespace  = kubernetes_namespace.longhorn_system.metadata[0].name

  # Wait for CRDs to be ready
  wait          = true
  wait_for_jobs = true
  timeout       = 600 # 10 minutes

  values = [
    yamlencode({
      # Default settings
      defaultSettings = {
        # Storage path on worker nodes (matches Talos kubelet extraMount)
        defaultDataPath = var.longhorn_disk_path

        # Replica settings for HA
        defaultReplicaCount = 2  # 2 replicas across 2 workers

        # Talos-specific settings
        taintToleration                     = "node-role.kubernetes.io/control-plane:NoSchedule"
        systemManagedComponentsNodeSelector = "node-role.kubernetes.io/worker:true"
        createDefaultDiskLabeledNodes       = true

        # Storage overprovisioning
        storageOverProvisioningPercentage = 200
        storageMinimalAvailablePercentage = 10

        # Backup settings (optional - configure if needed)
        # backupTarget = "s3://your-bucket@us-east-1/"
        # backupTargetCredentialSecret = "longhorn-backup-secret"

        # Node drain policies
        nodeDownPodDeletionPolicy          = "delete-both-statefulset-and-deployment-pod"
        allowNodeDrainWithLastHealthyReplica = false

        # Performance settings
        concurrentAutomaticEngineUpgradePerNodeLimit = 1
      }

      # Longhorn Manager (UI and API)
      longhornManager = {
        tolerations = [
          {
            key      = "node-role.kubernetes.io/control-plane"
            operator = "Exists"
            effect   = "NoSchedule"
          }
        ]
      }

      # Longhorn Driver (CSI)
      longhornDriver = {
        tolerations = [
          {
            key      = "node-role.kubernetes.io/control-plane"
            operator = "Exists"
            effect   = "NoSchedule"
          }
        ]
      }

      # Longhorn UI
      longhornUI = {
        tolerations = [
          {
            key      = "node-role.kubernetes.io/control-plane"
            operator = "Exists"
            effect   = "NoSchedule"
          }
        ]
      }

      # Service configuration
      service = {
        ui = {
          type = "ClusterIP"  # Change to LoadBalancer or use Ingress for external access
        }
      }

      # Resource limits
      resources = {
        limits = {
          cpu    = "200m"
          memory = "512Mi"
        }
        requests = {
          cpu    = "100m"
          memory = "256Mi"
        }
      }

      # Persistence
      persistence = {
        defaultClass       = true
        defaultClassReplicaCount = 2
        reclaimPolicy      = "Retain"  # Change to Delete if you want PVs deleted with PVCs
        migratable         = false
        recurringJobSelector = {
          enable = false
        }
      }
    })
  ]

  depends_on = [
    kubernetes_namespace.longhorn_system
  ]
}

# Create a StorageClass for Longhorn (if not created by Helm)
resource "kubernetes_storage_class" "longhorn" {
  metadata {
    name = "longhorn"
    annotations = {
      "storageclass.kubernetes.io/is-default-class" = "true"
    }
  }

  storage_provisioner = "driver.longhorn.io"
  reclaim_policy      = "Retain"
  allow_volume_expansion = true
  volume_binding_mode = "Immediate"

  parameters = {
    numberOfReplicas    = "2"
    staleReplicaTimeout = "2880"  # 48 hours
    fromBackup          = ""
    fsType              = "ext4"
  }

  depends_on = [helm_release.longhorn]
}

# Output Longhorn UI access information
output "longhorn_ui_access" {
  description = "Commands to access Longhorn UI"
  value       = <<-EOT
    Access Longhorn UI:

    1. Port-forward:
       kubectl port-forward -n longhorn-system svc/longhorn-frontend 8000:80

    2. Access at:
       http://localhost:8000

    3. Or create an Ingress (after installing ingress controller):
       kubectl apply -f deployments/k8s/infrastructure/ingress/longhorn-ingress.yaml
  EOT
}
