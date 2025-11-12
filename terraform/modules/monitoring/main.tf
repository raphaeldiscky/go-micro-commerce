# Monitoring Module
# Installs Prometheus, Grafana, Loki, and Tempo for observability

# Create namespace for monitoring
resource "kubernetes_namespace" "monitoring" {
  count = var.create_namespace ? 1 : 0

  metadata {
    name = var.namespace

    labels = {
      "app.kubernetes.io/name"       = "monitoring"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

# Install kube-prometheus-stack (Prometheus + Grafana + Alertmanager)
resource "helm_release" "kube_prometheus_stack" {
  name       = "kube-prometheus-stack"
  namespace  = var.namespace
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  version    = var.kube_prometheus_stack_version

  create_namespace = false

  values = [
    yamlencode({
      # Prometheus configuration
      prometheus = {
        prometheusSpec = {
          retention        = var.prometheus_retention
          retentionSize    = "45GB"
          storageSpec = {
            volumeClaimTemplate = {
              spec = {
                accessModes = ["ReadWriteOnce"]
                resources = {
                  requests = {
                    storage = var.prometheus_storage_size
                  }
                }
              }
            }
          }
          resources = {
            limits = {
              cpu    = "2000m"
              memory = "4Gi"
            }
            requests = {
              cpu    = "500m"
              memory = "2Gi"
            }
          }
        }
      }

      # Grafana configuration
      grafana = {
        enabled = true
        adminPassword = var.grafana_admin_password
        persistence = {
          enabled = true
          size    = "10Gi"
        }
        datasources = {
          "datasources.yaml" = {
            apiVersion = 1
            datasources = [
              {
                name      = "Prometheus"
                type      = "prometheus"
                url       = "http://kube-prometheus-stack-prometheus.${var.namespace}:9090"
                access    = "proxy"
                isDefault = true
              },
              {
                name   = "Loki"
                type   = "loki"
                url    = "http://loki.${var.namespace}:3100"
                access = "proxy"
              },
              {
                name   = "Tempo"
                type   = "tempo"
                url    = "http://tempo.${var.namespace}:3100"
                access = "proxy"
              }
            ]
          }
        }
        resources = {
          limits = {
            cpu    = "500m"
            memory = "512Mi"
          }
          requests = {
            cpu    = "100m"
            memory = "128Mi"
          }
        }
      }

      # Alertmanager configuration
      alertmanager = {
        alertmanagerSpec = {
          storage = {
            volumeClaimTemplate = {
              spec = {
                accessModes = ["ReadWriteOnce"]
                resources = {
                  requests = {
                    storage = "10Gi"
                  }
                }
              }
            }
          }
        }
      }
    })
  ]

  depends_on = [kubernetes_namespace.monitoring]
}

# Install Loki for log aggregation
resource "helm_release" "loki" {
  name       = "loki"
  namespace  = var.namespace
  repository = "https://grafana.github.io/helm-charts"
  chart      = "loki"
  version    = var.loki_version

  create_namespace = false

  values = [
    yamlencode({
      loki = {
        auth_enabled = false
        storage = {
          type = "filesystem"
        }
        commonConfig = {
          replication_factor = 1
        }
      }

      singleBinary = {
        replicas = 1
        persistence = {
          enabled = true
          size    = var.loki_storage_size
        }
        resources = {
          limits = {
            cpu    = "1000m"
            memory = "2Gi"
          }
          requests = {
            cpu    = "200m"
            memory = "512Mi"
          }
        }
      }

      monitoring = {
        serviceMonitor = {
          enabled = true
        }
      }
    })
  ]

  depends_on = [helm_release.kube_prometheus_stack]
}

# Install Tempo for distributed tracing
resource "helm_release" "tempo" {
  name       = "tempo"
  namespace  = var.namespace
  repository = "https://grafana.github.io/helm-charts"
  chart      = "tempo"
  version    = var.tempo_version

  create_namespace = false

  values = [
    yamlencode({
      tempo = {
        storage = {
          trace = {
            backend = "local"
            local = {
              path = "/var/tempo/traces"
            }
          }
        }
      }

      persistence = {
        enabled = true
        size    = var.tempo_storage_size
      }

      resources = {
        limits = {
          cpu    = "1000m"
          memory = "2Gi"
        }
        requests = {
          cpu    = "200m"
          memory = "512Mi"
        }
      }

      serviceMonitor = {
        enabled = true
      }
    })
  ]

  depends_on = [helm_release.kube_prometheus_stack]
}
