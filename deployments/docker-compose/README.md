# Monitoring Stack

This directory contains the complete monitoring stack for the Go microservices template, including metrics, logging, and distributed tracing.

## Components

### Prometheus (Port 9090)

- **Purpose**: Metrics collection and storage
- **URL**: <http://localhost:9090>
- **Configuration**: `prometheus.yml`

### Grafana (Port 3000)

- **Purpose**: Metrics visualization and dashboards
- **URL**: <http://localhost:3000>
- **Credentials**: admin/admin
- **Configuration**: `grafana/provisioning/`
- **Dashboards**: `grafana/dashboards/`

### Loki (Port 3100)

- **Purpose**: Log aggregation and storage
- **URL**: <http://localhost:3100>
- **Configuration**: `loki-config.yaml`

### Tempo (Port 3200)

- **Purpose**: Distributed tracing storage
- **URL**: <http://localhost:3200>
- **Configuration**: `tempo-config.yaml`
- **OTLP Endpoints**:
  - gRPC: localhost:4317
  - HTTP: localhost:4318

### OpenTelemetry Collector (Ports 4317/4318)

- **Purpose**: Telemetry data collection, processing, and export
- **Configuration**: `otel-collector-config.yaml`
- **Endpoints**:
  - OTLP gRPC: localhost:4317
  - OTLP HTTP: localhost:4318
  - Prometheus metrics: localhost:8888
  - Health check: localhost:13133

## Usage

### Starting the Monitoring Stack

```bash
# Start infrastructure services first
docker-compose -f infra.yaml up -d

# Start monitoring services
docker-compose -f monitoring.yaml up -d

# Start application services
docker-compose -f apps.yaml up -d
```

### Accessing the Services

1. **Grafana Dashboard**: <http://localhost:3000>
   - Username: admin
   - Password: admin
   - Pre-configured dashboards and data sources

2. **Prometheus**: <http://localhost:9090>
   - Query metrics directly
   - Check service discovery and targets

3. **Tempo**: <http://localhost:3200>
   - View traces through Grafana (integrated)
   - Direct API access for trace queries

4. **Loki**: <http://localhost:3100>
   - View logs through Grafana (integrated)
   - Direct API access for log queries

### Data Sources in Grafana

All data sources are automatically provisioned:

- **Prometheus**: Metrics from all services
- **Loki**: Logs from all services
- **Tempo**: Distributed traces with correlation to logs and metrics

### Metrics Collected

- HTTP request rates, latencies, and error rates
- Service-specific business metrics
- Infrastructure metrics (CPU, memory, etc.)
- Custom application metrics

### Traces

- End-to-end request tracing across all microservices
- Service dependency mapping
- Performance bottleneck identification
- Error correlation with logs

### Logs

- Centralized logging from all services
- Structured logging with trace correlation
- Log-based alerting capabilities

## Configuration Files

- `monitoring/config/prometheus.yml`: Prometheus scraping configuration
- `monitoring/config/loki-config.yaml`: Loki storage and ingestion configuration
- `monitoring/config/tempo-config.yaml`: Tempo tracing backend configuration
- `monitoring/config/otel-collector-config.yaml`: OpenTelemetry Collector pipeline configuration
- `monitoring/grafana/provisioning/`: Grafana auto-provisioning configuration
- `monitoring/grafana/dashboards/`: Pre-built Grafana dashboards## Service Configuration

Update your microservices to send telemetry to the OpenTelemetry Collector:

```bash
# Tracing endpoint
TRACING_URL=http://otel-collector:4318/v1/traces

# Metrics endpoint (if using OTLP metrics)
METRICS_URL=http://otel-collector:4318/v1/metrics

# Logs endpoint (if using OTLP logs)
LOGS_URL=http://otel-collector:4318/v1/logs
```

## Volumes

- `prometheus_data`: Prometheus metrics storage
- `grafana_data`: Grafana configuration and dashboards
- `loki_data`: Loki log storage
- `tempo_data`: Tempo trace storage

## Migration from Jaeger

This setup replaces Jaeger with Tempo for distributed tracing:

- **Before**: Direct export to Jaeger via OTLP
- **After**: Export to OpenTelemetry Collector, which forwards to Tempo
- **Benefits**: Better integration with Prometheus and Grafana, more efficient storage

### API Gateway Monitoring Endpoints

The API Gateway now includes comprehensive monitoring endpoints:

- **`/health`** - Health check with detailed status
- **`/ready`** - Readiness probe
- **`/info`** - Service information and version
- **`/metrics`** - Application metrics (JSON format)
- **`/test/trace`** - Create test trace for validation
- **`/test/error`** - Create test error for monitoring

## Troubleshooting

1. **Services not showing metrics**: Check Prometheus targets at <http://localhost:9090/targets>
2. **No traces appearing**: Verify OTLP collector is receiving data at <http://localhost:8888>
3. **Grafana data source issues**: Check provisioning logs in Grafana container
4. **Performance issues**: Adjust sampling rates in service configurations

## Register services manually with Consul

```
curl -X PUT http://localhost:8500/v1/agent/service/register \
-H "Content-Type: application/json" \
-d '{
  "ID": "auth-service-Raphael-8081",
  "Name": "auth-service",
  "Tags": ["http", "api", "microservice"],
  "Address": "192.168.0.107",
  "Port": 8081,
  "Check": {
    "HTTP": "http://192.168.0.107:8081/health",
    "Interval": "30s",
    "Timeout": "10s"
  }
}'
```

## Deregister

```
curl http://localhost:8500/v1/catalog/services
curl http://localhost:8500/v1/health/service/product-service
curl -X PUT http://localhost:8500/v1/agent/service/deregister/product-service-Raphael-8081
```
