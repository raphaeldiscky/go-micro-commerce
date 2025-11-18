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
  version    = var.kube_prometheus_stack_chart_version

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
              cpu    = "1000m"
              memory = "2Gi"
            }
            requests = {
              cpu    = "200m"
              memory = "1Gi"
            }
          }
          # Node affinity - run on monitoring pool
          nodeSelector = {
            workload-type = "monitoring"
          }
          # Tolerations for monitoring pool taint
          tolerations = [
            {
              key      = "workload-type"
              operator = "Equal"
              value    = "monitoring"
              effect   = "NoSchedule"
            }
          ]
        }
      }

      # Grafana configuration
      grafana = {
        enabled = true
        adminPassword = var.grafana_admin_password
        persistence = {
          enabled = true
          size    = "5Gi"
        }
        # Ingress configuration for external access
        ingress = {
          enabled          = var.grafana_enable_ingress
          ingressClassName = "traefik"
          hosts            = [var.grafana_domain_name]
          path             = "/"
          tls = [
            {
              secretName = "grafana-tls"
              hosts      = [var.grafana_domain_name]
            }
          ]
          annotations = {
            "cert-manager.io/cluster-issuer" = var.grafana_tls_issuer
            "traefik.ingress.kubernetes.io/router.entrypoints" = "websecure"
          }
        }
        # Additional datasources (Prometheus and Alertmanager are auto-configured by the chart)
        additionalDataSources = [
          {
            name   = "Loki"
            type   = "loki"
            url    = "http://loki-gateway.${var.namespace}"
            access = "proxy"
          },
          {
            name   = "Tempo"
            type   = "tempo"
            url    = "http://tempo.${var.namespace}:3100"
            access = "proxy"
          }
        ]
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
        # Node affinity - run on monitoring pool
        nodeSelector = {
          workload-type = "monitoring"
        }
        # Tolerations for monitoring pool taint
        tolerations = [
          {
            key      = "workload-type"
            operator = "Equal"
            value    = "monitoring"
            effect   = "NoSchedule"
          }
        ]
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
                    storage = "2Gi"
                  }
                }
              }
            }
          }
          # Node affinity - run on monitoring pool
          nodeSelector = {
            workload-type = "monitoring"
          }
          # Tolerations for monitoring pool taint
          tolerations = [
            {
              key      = "workload-type"
              operator = "Equal"
              value    = "monitoring"
              effect   = "NoSchedule"
            }
          ]
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
  version    = var.loki_chart_version

  create_namespace = false

  values = [
    yamlencode({
      # Deployment mode
      deploymentMode = "SingleBinary"

      loki = {
        storage = {
          type = "filesystem"
          bucketNames = {
            chunks = "chunks"
            ruler  = "ruler"
            admin  = "admin"
          }
        }

        structuredConfig = {
          auth_enabled = false

          server = {
            http_listen_port = 3100
            grpc_listen_port = 9095
          }

          common = {
            path_prefix = "/var/loki"
            storage = {
              filesystem = {
                chunks_directory = "/var/loki/chunks"
                rules_directory  = "/var/loki/rules"
              }
            }
            replication_factor = 1
            ring = {
              kvstore = {
                store = "inmemory"
              }
            }
          }

          schema_config = {
            configs = [
              {
                from         = "2024-01-01"
                store        = "tsdb"
                object_store = "filesystem"
                schema       = "v13"
                index = {
                  prefix = "index_"
                  period = "24h"
                }
              }
            ]

          }

          limits_config = {
            reject_old_samples          = true
            reject_old_samples_max_age  = "168h"
            max_entries_limit_per_query = 5000
            ingestion_rate_mb           = 16
            ingestion_burst_size_mb     = 24
            retention_period            = "168h"
          }

          compactor = {
            working_directory          = "/var/loki/compactor"
            retention_enabled          = true
            retention_delete_delay     = "2h"
            retention_delete_worker_count = 150
            delete_request_store       = "filesystem"
          }

          query_range = {
            align_queries_with_step = true
            cache_results           = true
          }
        }
      }

      # SingleBinary configuration
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

        # Node affinity - run on monitoring pool
        nodeSelector = {
          workload-type = "monitoring"
        }

        # Tolerations for monitoring pool taint
        tolerations = [
          {
            key      = "workload-type"
            operator = "Equal"
            value    = "monitoring"
            effect   = "NoSchedule"
          }
        ]
      }

      # Monitoring configuration
      monitoring = {
        selfMonitoring = {
          enabled = false
          grafanaAgent = {
            installOperator = false
          }
        }
        serviceMonitor = {
          enabled = true
        }
      }

      # Gateway configuration
      gateway = {
        enabled  = true
        replicas = 1

        service = {
          type = "ClusterIP"
          port = 80
        }

        # Fix nginx DNS resolver to use kube-dns ClusterIP directly
        # Prevents chicken-and-egg problem of nginx needing DNS to resolve DNS service hostname
        nginxConfig = {
          resolver = "10.8.0.10"
        }

        resources = {
          limits = {
            cpu    = "100m"
            memory = "128Mi"
          }
          requests = {
            cpu    = "50m"
            memory = "64Mi"
          }
        }

        # Node affinity - run on monitoring pool
        nodeSelector = {
          workload-type = "monitoring"
        }

        # Tolerations for monitoring pool taint
        tolerations = [
          {
            key      = "workload-type"
            operator = "Equal"
            value    = "monitoring"
            effect   = "NoSchedule"
          }
        ]
      }

      # Disable components not needed for SingleBinary
      backend = {
        replicas = 0
      }
      read = {
        replicas = 0
      }
      write = {
        replicas = 0
      }
      chunksCache = {
        enabled = false
      }
      resultsCache = {
        enabled = false
      }
      lokiCanary = {
        enabled = true
      }
    })
  ]

  depends_on = [helm_release.kube_prometheus_stack]
}

# Install Grafana Alloy for log collection
resource "helm_release" "alloy" {
  name       = "alloy"
  namespace  = var.namespace
  repository = "https://grafana.github.io/helm-charts"
  chart      = "alloy"
  version    = var.alloy_chart_version

  create_namespace = false

  values = [
    yamlencode({
      # Deploy as DaemonSet to collect logs from all nodes
      alloy = {
        mode = "daemonset"

        # Alloy configuration using River language
        configMap = {
          create  = true
          content = <<-EOT
            // Discover Kubernetes pods
            discovery.kubernetes "pods" {
              role = "pod"
            }

            // Relabel discovered pods to extract metadata
            discovery.relabel "kubernetes_pods" {
              targets = discovery.kubernetes.pods.targets

              // Set job label to namespace/pod_name
              rule {
                source_labels = ["__meta_kubernetes_namespace", "__meta_kubernetes_pod_name"]
                separator     = "/"
                target_label  = "job"
              }

              // Set pod label
              rule {
                source_labels = ["__meta_kubernetes_pod_name"]
                target_label  = "pod"
              }

              // Set container label
              rule {
                source_labels = ["__meta_kubernetes_pod_container_name"]
                target_label  = "container"
              }

              // Set namespace label
              rule {
                source_labels = ["__meta_kubernetes_namespace"]
                target_label  = "namespace"
              }

              // Set app label from pod labels
              rule {
                source_labels = ["__meta_kubernetes_pod_label_app"]
                target_label  = "app"
              }

              // Set component label from pod labels
              rule {
                source_labels = ["__meta_kubernetes_pod_label_component"]
                target_label  = "component"
              }

              // Drop empty targets
              rule {
                source_labels = ["__meta_kubernetes_pod_container_name"]
                action        = "drop"
                regex         = ""
              }

              // Set path to container logs
              rule {
                source_labels = ["__meta_kubernetes_pod_uid", "__meta_kubernetes_pod_container_name"]
                separator     = "/"
                target_label  = "__path__"
                replacement   = "/var/log/pods/*$1/*.log"
              }
            }

            // Scrape logs from discovered pods
            loki.source.kubernetes "pods" {
              targets    = discovery.relabel.kubernetes_pods.output
              forward_to = [loki.write.local.receiver]
            }

            // Write logs to Loki
            loki.write "local" {
              endpoint {
                url = "http://loki-gateway.${var.namespace}/loki/api/v1/push"
              }
            }
          EOT
        }

        # Resource limits for production
        resources = {
          requests = {
            cpu    = "100m"
            memory = "128Mi"
          }
          limits = {
            cpu    = var.alloy_cpu_limit
            memory = var.alloy_memory_limit
          }
        }

        # Mount host paths to access container logs
        mounts = {
          varlog           = true
          dockercontainers = false
        }

        # Enable ServiceMonitor for Prometheus metrics
        serviceMonitor = {
          enabled = true
        }

        # Security context
        securityContext = {
          privileged = false
          runAsUser  = 0
          runAsGroup = 0
        }
      }

      # Controller configuration
      controller = {
        type = "daemonset"

        # Host network access for log collection
        hostNetwork = false
        hostPID     = false
      }

      # Service account with proper RBAC
      rbac = {
        create = true
      }

      serviceAccount = {
        create = true
        name   = "alloy"
      }
    })
  ]

  depends_on = [helm_release.loki]
}

# Install Tempo for distributed tracing
resource "helm_release" "tempo" {
  name       = "tempo"
  namespace  = var.namespace
  repository = "https://grafana.github.io/helm-charts"
  chart      = "tempo"
  version    = var.tempo_chart_version

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
          cpu    = "500m"
          memory = "1Gi"
        }
        requests = {
          cpu    = "100m"
          memory = "256Mi"
        }
      }

      # Node affinity - run on monitoring pool
      nodeSelector = {
        workload-type = "monitoring"
      }

      # Tolerations for monitoring pool taint
      tolerations = [
        {
          key      = "workload-type"
          operator = "Equal"
          value    = "monitoring"
          effect   = "NoSchedule"
        }
      ]

      serviceMonitor = {
        enabled = true
      }
    })
  ]

  depends_on = [helm_release.kube_prometheus_stack]
}
