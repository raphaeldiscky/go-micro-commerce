# Strimzi Kafka Operator Module
# Installs Strimzi operator for managing Kafka clusters with KRaft

# Create namespace for the operator
resource "kubernetes_namespace" "kafka" {
  count = var.create_namespace ? 1 : 0

  metadata {
    name = var.namespace

    labels = {
      "app.kubernetes.io/name"       = "strimzi"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

# Install Strimzi Kafka operator via Helm
resource "helm_release" "strimzi" {
  name       = "strimzi-kafka-operator"
  namespace  = var.namespace
  repository = "https://strimzi.io/charts/"
  chart      = "strimzi-kafka-operator"
  version    = var.chart_version

  create_namespace = false # Already created above

  values = [
    yamlencode({
      # Watch specific namespaces or all
      watchNamespaces = length(var.watch_namespaces) > 0 ? var.watch_namespaces : [var.namespace]
      watchAnyNamespace = length(var.watch_namespaces) == 0

      # Operator configuration
      logLevel = var.log_level

      # Feature gates for KRaft mode
      featureGates = "+UseKRaft,+KafkaNodePools"

      # Resource limits
      resources = {
        limits = {
          cpu    = "1000m"
          memory = "1Gi"
        }
        requests = {
          cpu    = "200m"
          memory = "256Mi"
        }
      }

      # Replica count
      replicas = 1

      # Node affinity for control plane pool
      nodeSelector = {
        workload-type = "control-plane"
      }

      tolerations = [
        {
          key      = "workload-type"
          operator = "Equal"
          value    = "control-plane"
          effect   = "NoSchedule"
        }
      ]

      # Security context
      podSecurityContext = {
        runAsNonRoot = true
      }
    })
  ]

  depends_on = [kubernetes_namespace.kafka]
}
