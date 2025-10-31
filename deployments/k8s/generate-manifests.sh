#!/bin/bash

# Script to generate Kubernetes base manifests for all microservices
# This creates a consistent structure for all services

set -e

# Service configuration: name:port:tier
# tier can be: gateway, backend, infrastructure
SERVICES=(
  "api-gateway:8080:gateway"
  "graphql-gateway:4000:gateway"
  "auth-service:8081:backend"
  "product-service:8082:backend"
  "order-service:8083:backend"
  "payment-service:8084:backend"
  "cart-service:8089:backend"
  "fulfillment-service:8085:backend"
  "notification-service:8086:backend"
  "search-service:8087:backend"
  "chat-service:8088:backend"
)

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/base"

# Function to create deployment.yaml
create_deployment() {
  local service_name=$1
  local port=$2
  local tier=$3

  cat > "${BASE_DIR}/${service_name}/deployment.yaml" <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${service_name}
  labels:
    app: ${service_name}
    tier: ${tier}
    version: v1
spec:
  replicas: 2
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: ${service_name}
  template:
    metadata:
      labels:
        app: ${service_name}
        tier: ${tier}
        version: v1
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "${port}"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: ${service_name}
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
      - name: ${service_name}
        image: ${service_name}:latest
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          containerPort: ${port}
          protocol: TCP
        env:
        - name: APP_NAME
          value: "${service_name}"
        - name: APP_ENVIRONMENT
          valueFrom:
            configMapKeyRef:
              name: ${service_name}-config
              key: APP_ENVIRONMENT
        - name: HTTP_SERVER_HOST
          value: "0.0.0.0"
        - name: HTTP_SERVER_PORT
          value: "${port}"
        - name: CONSUL_ENABLED
          value: "false"
        - name: TRACING_ENABLED
          value: "true"
        - name: TRACING_SERVICE_NAME
          value: "${service_name}"
        - name: METRICS_ENABLED
          value: "true"
        envFrom:
        - configMapRef:
            name: ${service_name}-config
        - secretRef:
            name: ${service_name}-secrets
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: false
          capabilities:
            drop:
            - ALL
EOF
}

# Function to create service.yaml
create_service() {
  local service_name=$1
  local port=$2
  local tier=$3

  cat > "${BASE_DIR}/${service_name}/service.yaml" <<EOF
apiVersion: v1
kind: Service
metadata:
  name: ${service_name}
  labels:
    app: ${service_name}
    tier: ${tier}
spec:
  type: ClusterIP
  selector:
    app: ${service_name}
  ports:
  - name: http
    port: ${port}
    targetPort: http
    protocol: TCP
  sessionAffinity: None
EOF
}

# Function to create configmap.yaml
create_configmap() {
  local service_name=$1

  cat > "${BASE_DIR}/${service_name}/configmap.yaml" <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: ${service_name}-config
  labels:
    app: ${service_name}
data:
  APP_ENVIRONMENT: "production"
  APP_LOGGER_LEVEL: "4"
  APP_TIMEOUT_SHUTDOWN: "10s"

  # Observability
  TRACING_URL: "otel-collector:4318"
  TRACING_SAMPLING_RATE: "0.1"
  TRACING_ENVIRONMENT: "production"
  METRICS_PATH: "/metrics"

  # Add service-specific config here in overlays
EOF
}

# Function to create secret.yaml
create_secret() {
  local service_name=$1

  cat > "${BASE_DIR}/${service_name}/secret.yaml" <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: ${service_name}-secrets
  labels:
    app: ${service_name}
type: Opaque
stringData:
  # Add sensitive configuration here
  # This is a template - actual values should be set per environment
  # Consider using sealed-secrets or external-secrets operator for production
  PLACEHOLDER: "replace-with-actual-secrets"
EOF
}

# Function to create serviceaccount.yaml
create_serviceaccount() {
  local service_name=$1

  cat > "${BASE_DIR}/${service_name}/serviceaccount.yaml" <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ${service_name}
  labels:
    app: ${service_name}
automountServiceAccountToken: true
EOF
}

# Function to create hpa.yaml
create_hpa() {
  local service_name=$1

  cat > "${BASE_DIR}/${service_name}/hpa.yaml" <<EOF
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: ${service_name}
  labels:
    app: ${service_name}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: ${service_name}
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 30
      - type: Pods
        value: 2
        periodSeconds: 30
      selectPolicy: Max
EOF
}

# Function to create pdb.yaml
create_pdb() {
  local service_name=$1

  cat > "${BASE_DIR}/${service_name}/pdb.yaml" <<EOF
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: ${service_name}
  labels:
    app: ${service_name}
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: ${service_name}
EOF
}

# Function to create kustomization.yaml
create_kustomization() {
  local service_name=$1
  local tier=$2

  cat > "${BASE_DIR}/${service_name}/kustomization.yaml" <<EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

metadata:
  name: ${service_name}

namespace: default

labels:
- includeSelectors: true
  pairs:
    app.kubernetes.io/name: ${service_name}
    app.kubernetes.io/component: ${tier}
    app.kubernetes.io/part-of: go-micro-commerce

resources:
  - deployment.yaml
  - service.yaml
  - configmap.yaml
  - secret.yaml
  - serviceaccount.yaml
  - hpa.yaml
  - pdb.yaml

images:
  - name: ${service_name}
    newName: ${service_name}
    newTag: latest
EOF
}

# Main execution
echo "Generating Kubernetes manifests for all services..."

for service_config in "${SERVICES[@]}"; do
  IFS=':' read -r service_name port tier <<< "$service_config"

  echo "Creating manifests for ${service_name}..."

  # Create directory if it doesn't exist
  mkdir -p "${BASE_DIR}/${service_name}"

  # Create all manifest files
  create_deployment "$service_name" "$port" "$tier"
  create_service "$service_name" "$port" "$tier"
  create_configmap "$service_name"
  create_secret "$service_name"
  create_serviceaccount "$service_name"
  create_hpa "$service_name"
  create_pdb "$service_name"
  create_kustomization "$service_name" "$tier"

  echo "  ✓ Created manifests for ${service_name}"
done

echo ""
echo "✅ All service manifests generated successfully!"
echo "📁 Location: ${BASE_DIR}/"
echo ""
echo "Next steps:"
echo "1. Review and customize ConfigMaps for each service"
echo "2. Add secrets to Secret manifests (use sealed-secrets for prod)"
echo "3. Create environment-specific overlays in deployments/k8s/overlays/"
echo "4. Apply manifests: kubectl apply -k deployments/k8s/base/<service-name>/"
