# Tiltfile for go-micro-commerce
# Local Kubernetes development with Tilt
# Docs: https://docs.tilt.dev/

# Tilt settings
update_settings(max_parallel_updates=3, k8s_upsert_timeout_secs=300)

# Allow Kubernetes contexts (adjust for your local cluster)
allow_k8s_contexts(['kind-go-micro-commerce', 'minikube', 'docker-desktop', 'rancher-desktop'])

# Load Tilt extensions
load('ext://helm_resource', 'helm_resource', 'helm_repo')
load('ext://restart_process', 'docker_build_with_restart')

#==================================================================================
# HELM REPOSITORIES
#==================================================================================

helm_repo('bitnami', 'https://charts.bitnami.com/bitnami')
helm_repo('prometheus-community', 'https://prometheus-community.github.io/helm-charts')
helm_repo('grafana', 'https://grafana.github.io/helm-charts')

#==================================================================================
# INFRASTRUCTURE - DATABASES (PostgreSQL)
#==================================================================================

# Deploy 9 PostgreSQL instances (one per microservice)
postgres_services = [
    'postgresql-auth',
    'postgresql-product',
    'postgresql-order',
    'postgresql-payment',
    'postgresql-fulfillment',
    'postgresql-notification',
    'postgresql-search',
    'postgresql-chat',
    'postgresql-cart',
]

for pg_service in postgres_services:
    k8s_yaml('deployments/k8s/infrastructure/postgresql/postgres-' + pg_service.replace('postgresql-', '') + '.yaml')
    k8s_resource(
        pg_service,
        labels=['infra-db'],
        resource_deps=[],
    )

#==================================================================================
# INFRASTRUCTURE - REDIS CLUSTER
#==================================================================================

helm_resource(
    'redis-cluster',
    'bitnami/redis-cluster',
    flags=[
        '--values=deployments/k8s/infrastructure/redis/values.yaml',
        '--version=8.2.1',
    ],
    labels=['infra-cache'],
    resource_deps=['kube-prometheus-stack'],  # Wait for Prometheus CRDs to be installed
)

# Redis Insight UI
k8s_yaml('deployments/k8s/infrastructure/redis/redis-insight.yaml')
k8s_resource(
    'redis-insight',
    port_forwards=['5540:5540'],
    labels=['infra-cache'],
    resource_deps=['redis-cluster'],
)

#==================================================================================
# INFRASTRUCTURE - KAFKA CLUSTER
#==================================================================================

helm_resource(
    'kafka',
    'bitnami/kafka',
    flags=[
        '--values=deployments/k8s/infrastructure/kafka/values.yaml',
        '--version=32.4.3',
    ],
    labels=['infra-messaging'],
    resource_deps=['kube-prometheus-stack'],  # Wait for Prometheus CRDs to be installed
)

# Kafka UI
k8s_yaml('deployments/k8s/infrastructure/kafka/kafka-ui.yaml')
k8s_resource(
    'kafka-ui',
    port_forwards=['8090:8080'],
    labels=['infra-messaging'],
    resource_deps=['kafka'],
)

#==================================================================================
# INFRASTRUCTURE - MONITORING STACK
#==================================================================================

# Prometheus + Grafana + Alertmanager
helm_resource(
    'kube-prometheus-stack',
    'prometheus-community/kube-prometheus-stack',
    flags=[
        '--values=deployments/k8s/infrastructure/monitoring/prometheus-values.yaml',
        '--version=79.1.1',
    ],
    labels=['monitoring'],
    resource_deps=[],
    port_forwards=[
        '3000:80',   # Grafana
        '9090:9090', # Prometheus
    ],
)

# Loki + Promtail (log aggregation)
helm_resource(
    'loki',
    'grafana/loki-stack',
    flags=[
        '--values=deployments/k8s/infrastructure/monitoring/loki-values.yaml',
        '--version=2.10.3',
    ],
    labels=['monitoring'],
    resource_deps=['kube-prometheus-stack'],  # Wait for Prometheus CRDs
)

# Tempo (distributed tracing) - matches Docker Compose tempo:2.8.1
helm_resource(
    'tempo',
    'grafana/tempo',
    flags=[
        '--values=deployments/k8s/infrastructure/monitoring/tempo-values.yaml',
        '--version=1.23.2',
    ],
    labels=['monitoring'],
    resource_deps=['kube-prometheus-stack'],  # Wait for Prometheus CRDs
)

# OpenTelemetry Collector
k8s_yaml('deployments/k8s/infrastructure/monitoring/otel-collector.yaml')
k8s_resource(
    'otel-collector',
    labels=['monitoring'],
    resource_deps=['tempo', 'loki'],
)

# Grafana and Prometheus are now automatically port-forwarded via helm_resource() above
# Access at: http://localhost:3000 (Grafana) and http://localhost:9090 (Prometheus)

#==================================================================================
# INFRASTRUCTURE - DEV TOOLS
#==================================================================================

# MailHog (SMTP testing)
k8s_yaml('deployments/k8s/infrastructure/dev-tools/mailhog.yaml')
k8s_resource(
    'mailhog',
    port_forwards=['8025:8025', '1025:1025'],
    labels=['dev-tools'],
)

#==================================================================================
# INFRASTRUCTURE - TRAEFIK (API Gateway/Ingress)
#==================================================================================

# Note: Traefik is deployed via existing K8s manifests
# If you have Traefik Helm chart, uncomment and configure:
# helm_resource(
#     'traefik',
#     'traefik/traefik',
#     flags=[
#         '--values=deployments/k8s/infrastructure/traefik/values.yaml',
#     ],
#     labels=['infra-gateway'],
# )

#==================================================================================
# APPLICATION SERVICES
#==================================================================================

# List of microservices
services = [
    'api-gateway',
    'graphql-gateway',
    'auth-service',
    'product-service',
    'order-service',
    'payment-service',
    'cart-service',
    'fulfillment-service',
    'notification-service',
    'search-service',
    'chat-service',
]

# Map services to their database dependencies
service_to_db = {
    'auth-service': 'postgresql-auth',
    'product-service': 'postgresql-product',
    'order-service': 'postgresql-order',
    'payment-service': 'postgresql-payment',
    'cart-service': 'postgresql-cart',
    'fulfillment-service': 'postgresql-fulfillment',
    'notification-service': 'postgresql-notification',
    'search-service': 'postgresql-search',
    'chat-service': 'postgresql-chat',
}

# Deploy all services using Kustomize (once, not per service)
k8s_yaml(kustomize('deployments/k8s/overlays/local'))

# Build and deploy each service
for service in services:
    # Docker build with hot reload
    docker_build(
        'localhost:5000/%s' % service,
        context='./%s' % service,
        dockerfile='./%s/Dockerfile' % service,
        # Live update for Go hot-reload (optional - requires air or similar)
        live_update=[
            sync('./%s' % service, '/app'),
            run(
                'cd /app && go build -o main ./cmd/api',
                trigger=['./%s/**/*.go' % service]
            ),
        ],
        # Ignore files to speed up builds
        ignore=[
            './%s/**/*_test.go' % service,
            './%s/.git' % service,
            './%s/.env' % service,
        ],
    )

    # Resource dependencies
    deps = ['redis-cluster', 'kafka', 'otel-collector']

    # Add database dependency if service uses one
    if service in service_to_db:
        deps.append(service_to_db[service])

    # Configure K8s resource (with local- prefix from kustomization.yaml)
    k8s_resource(
        'local-%s' % service,
        new_name=service,
        labels=['apps'],
        resource_deps=deps,
    )

#==================================================================================
# PORT FORWARDS
#==================================================================================

# Application services (if needed for direct access)
# Uncomment to expose individual services:
# k8s_resource('api-gateway', port_forwards=['8080:8080'])
# k8s_resource('product-service', port_forwards=['8081:8080'])

#==================================================================================
# RESOURCE GROUPS (for Tilt UI organization)
#==================================================================================

# This organizes resources in the Tilt UI by labels:
# - infra-db: PostgreSQL databases
# - infra-cache: Redis cluster
# - infra-messaging: Kafka cluster
# - monitoring: Prometheus, Grafana, Loki, Tempo, OTEL
# - dev-tools: MailHog, Redis Insight, Kafka UI
# - apps: Microservices

print("Tiltfile loaded successfully!")
print("Resource groups:")
print("  - infra-db: 9 PostgreSQL databases")
print("  - infra-cache: Redis Cluster + Redis Insight")
print("  - infra-messaging: Kafka + Kafka UI")
print("  - monitoring: Prometheus, Grafana, Loki, Tempo, OTEL")
print("  - dev-tools: MailHog")
print("  - apps: 11 microservices")
print("")
print("Access UIs (automatically port-forwarded):")
print("  - Grafana: http://localhost:3000 (admin/admin)")
print("  - Prometheus: http://localhost:9090")
print("  - Kafka UI: http://localhost:8090")
print("  - Redis Insight: http://localhost:5540")
print("  - MailHog: http://localhost:8025")
print("")
print("Run 'tilt up' to start!")
