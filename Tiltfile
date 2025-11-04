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

helm_repo('strimzi', 'https://strimzi.io/charts/')
helm_repo('ot-helm', 'https://ot-container-kit.github.io/helm-charts/')
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
# INFRASTRUCTURE - REDIS CLUSTER (OT-Container-Kit Operator)
#==================================================================================

# Install OT-Container-Kit Redis Operator
helm_resource(
    'redis-operator',
    'ot-helm/redis-operator',
    flags=[
        '--values=deployments/k8s/infrastructure/redis/redis-operator-values.yaml',
    ],
    labels=['operators'],
    resource_deps=['kube-prometheus-stack'],  # Wait for Prometheus CRDs
)

# Wait for Redis CRDs to be installed by the operator
local_resource(
    'wait-redis-crds',
    'kubectl wait --for condition=established --timeout=60s crd/redisclusters.redis.redis.opstreelabs.in',
    resource_deps=['redis-operator'],
    labels=['operators'],
)

# Deploy RedisCluster CRD (6 nodes: 3 masters + 3 replicas)
k8s_yaml('deployments/k8s/infrastructure/redis/redis-cluster.yaml')
k8s_resource(
    objects=['redis-cluster:rediscluster'],
    new_name='redis-cluster',
    labels=['infra-cache'],
    resource_deps=['wait-redis-crds'],
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
# INFRASTRUCTURE - KAFKA CLUSTER (Strimzi Operator)
#==================================================================================

# Install Strimzi Kafka Operator
helm_resource(
    'strimzi-operator',
    'strimzi/strimzi-kafka-operator',
    namespace='default',
    flags=[
        '--values=deployments/k8s/infrastructure/kafka/strimzi-operator-values.yaml',
        '--version=0.48.0',
    ],
    labels=['operators'],
    resource_deps=['kube-prometheus-stack'], 
)

# Wait for Kafka CRDs to be installed by the operator
local_resource(
    'wait-kafka-crds',
    'kubectl wait --for condition=established --timeout=60s crd/kafkas.kafka.strimzi.io crd/kafkanodepools.kafka.strimzi.io',
    resource_deps=['strimzi-operator'],
    labels=['operators'],
)

# Deploy Kafka cluster CRD (3-node KRaft cluster)
k8s_yaml('deployments/k8s/infrastructure/kafka/kafka-cluster.yaml')
k8s_resource(
    objects=['kafka-cluster:kafka'],
    new_name='kafka-cluster',
    labels=['infra-messaging'],
    resource_deps=['wait-kafka-crds'],
)

# Kafka UI
k8s_yaml('deployments/k8s/infrastructure/kafka/kafka-ui.yaml')
k8s_resource(
    'kafka-ui',
    port_forwards=['8090:8080'],
    labels=['infra-messaging'],
    resource_deps=['kafka-cluster'],
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
    resource_deps=['kube-prometheus-stack'], 
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
    resource_deps=['kube-prometheus-stack'], 
)

# OpenTelemetry Collector
k8s_yaml('deployments/k8s/infrastructure/monitoring/otel-collector.yaml')
k8s_resource(
    'otel-collector',
    labels=['monitoring'],
    resource_deps=['tempo', 'loki'],
)

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

#==================================================================================
# MIGRATION IMAGES
#==================================================================================

# List of services with database migrations
migration_services = [
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

# Build migration images (lightweight images with only golang-migrate + SQL files)
for service in migration_services:
    docker_build(
        'localhost:5000/%s-migrations' % service,
        context='./%s' % service,
        dockerfile='./%s/Dockerfile.migrations' % service,
        only=['db/migrations'],
    )

# Register migration job resources (deployed by Kustomize, managed by Tilt)
for service in migration_services:
    k8s_resource(
        'local-%s-migration' % service,
        new_name='%s-migration' % service,
        labels=['migrations'],
        resource_deps=[service_to_db.get(service)],  # Wait for database to be ready
    )

# Services that require proto directory access (gRPC-enabled services)
services_needing_proto = ['product-service', 'order-service', 'payment-service', 'cart-service', 'fulfillment-service']

# Build and deploy each service
for service in services:
    # Handle graphql-gateway separately (non-Go service, different requirements)
    if service == 'graphql-gateway':
        docker_build(
            'localhost:5000/%s' % service,
            context='./%s' % service,
            dockerfile='./%s/Dockerfile' % service,
            only=[service],
            ignore=[
                './%s/.git' % service,
                './%s/.env' % service,
            ],
        )
    else:
        # Go services - need access to pkg/ and optionally proto/
        watch_paths = [service, 'pkg']
        if service in services_needing_proto:
            watch_paths.append('proto')

        # Docker build with hot reload
        docker_build(
            'localhost:5000/%s' % service,
            context='.',  # Repository root - can access pkg/, proto/, and service dirs
            dockerfile='./%s/Dockerfile' % service,
            only=watch_paths,  # Only rebuild when these paths change
            # Live update for Go hot-reload
            live_update=[
                sync('./%s' % service, '/workspace/%s' % service),
                run(
                    'cd /workspace/%s && go build -o main ./cmd/api' % service,
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
    # Services wait for actual infrastructure clusters to be ready
    deps = ['redis-cluster', 'kafka-cluster', 'otel-collector']

    # Add database dependency if service uses one
    if service in service_to_db:
        deps.append(service_to_db[service])

    # Add migration job dependency if service has migrations
    # Services must wait for migrations to complete before starting
    if service in migration_services:
        deps.append('%s-migration' % service)

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
# - operators: Strimzi, OT-Container-Kit (Redis)
# - infra-db: PostgreSQL databases
# - infra-cache: Redis cluster, Redis Insight
# - infra-messaging: Kafka cluster, Kafka UI
# - monitoring: Prometheus, Grafana, Loki, Tempo, OTEL
# - dev-tools: MailHog
# - migrations: Database migration jobs (9 jobs)
# - apps: Microservices

print("Tiltfile loaded successfully!")
print("")
print("=== Operator-Based Infrastructure ===")
print("  - Strimzi Kafka Operator (CNCF) - Kafka 4.0.0 KRaft mode")
print("  - OT-Container-Kit Redis Operator - Redis 7.x Cluster mode")
print("")
print("=== Resource Groups ===")
print("  - operators: Strimzi, OT-Container-Kit (Redis)")
print("  - infra-db: 9 PostgreSQL databases")
print("  - infra-cache: Redis Cluster (6 nodes) + Redis Insight")
print("  - infra-messaging: Kafka Cluster (3 nodes) + Kafka UI")
print("  - monitoring: Prometheus, Grafana, Loki, Tempo, OTEL")
print("  - dev-tools: MailHog")
print("  - migrations: 9 database migration jobs")
print("  - apps: 11 microservices")
print("")
print("=== Access UIs (automatically port-forwarded) ===")
print("  - Grafana: http://localhost:3000 (admin/admin)")
print("  - Prometheus: http://localhost:9090")
print("  - Kafka UI: http://localhost:8090")
print("  - Redis Insight: http://localhost:5540")
print("  - MailHog: http://localhost:8025")
print("")
print("Run 'tilt up' to start!")

