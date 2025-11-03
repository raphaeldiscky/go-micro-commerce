# Traefik Ingress Controller Setup

This directory contains configuration for deploying Traefik as the Ingress Controller for the go-micro-commerce platform.

## Architecture

Traefik serves as the **external ingress controller**, handling:

- TLS termination
- External routing to API Gateway and GraphQL Gateway
- Rate limiting
- Security headers
- CORS policies
- Metrics collection

The **API Gateway** then handles application-level concerns like:

- Service discovery (Kubernetes DNS)
- Internal routing to microservices
- Circuit breaking
- Authentication/Authorization

## Installation

### Prerequisites

1. **Helm 3.x** installed
2. **cert-manager** for TLS certificates (optional but recommended)

### Install cert-manager (for TLS)

```bash
# https://cert-manager.io/docs/installation/
# Install cert-manager CRDs
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/vx.x.x/cert-manager.yaml

# Add Jetstack Helm repository
helm repo add jetstack https://charts.jetstack.io
helm repo update

# Install cert-manager
helm install \
  cert-manager oci://quay.io/jetstack/charts/cert-manager \
  --version vx.x.x \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true
```

### Install Traefik

```bash
# https://doc.traefik.io/traefik/getting-started/quick-start-with-kubernetes/
# Add Traefik Helm repository
helm repo add traefik https://traefik.github.io/charts
helm repo update

# Install Traefik with custom values
helm install traefik traefik/traefik \
  -f values.yaml \
  -n traefik \
  --create-namespace
```

### Verify Installation

```bash
# Check Traefik pods
kubectl get pods -n traefik

# Check Traefik service
kubectl get svc -n traefik

# Access Traefik dashboard (if enabled)
kubectl port-forward -n traefik svc/traefik 9000:9000
# Visit http://localhost:9000/dashboard/
```

## Apply Ingress Resources

```bash
# Apply ingress resources for production
kubectl apply -f ingress.yaml

# Verify ingress
kubectl get ingress -n production
```

## Configure DNS

After Traefik is deployed, configure your DNS to point to the LoadBalancer:

```bash
# Get LoadBalancer external IP
kubectl get svc -n traefik traefik

# Configure DNS A records:
# api.yourdomain.com -> <EXTERNAL-IP>
# graphql.yourdomain.com -> <EXTERNAL-IP>
```

## TLS Certificate Setup

Create a ClusterIssuer for Let's Encrypt:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: traefik
EOF
```

## Monitoring

Traefik exposes Prometheus metrics on port 9100:

```yaml
# Add to Prometheus scrape config
- job_name: "traefik"
  kubernetes_sd_configs:
    - role: pod
      namespaces:
        names:
          - traefik
  relabel_configs:
    - source_labels: [__meta_kubernetes_pod_label_app_kubernetes_io_name]
      action: keep
      regex: traefik
```

## Troubleshooting

### Check Traefik logs

```bash
kubectl logs -n traefik -l app.kubernetes.io/name=traefik -f
```

### Verify Ingress configuration

```bash
kubectl describe ingress -n production api-gateway-ingress
```

### Test connectivity

```bash
curl -I https://api.yourdomain.com/health
```

## Customization

### Rate Limiting

Adjust rate limits in `ingress.yaml`:

```yaml
spec:
  rateLimit:
    average: 1000 # requests per period
    burst: 100 # burst capacity
    period: 1m # time window
```

### Security Headers

Modify security headers in the Middleware resource:

```yaml
spec:
  headers:
    frameDeny: true
    stsSeconds: 31536000
    # Add more headers as needed
```

## Production Checklist

- [ ] Configure DNS records
- [ ] Set up TLS certificates with Let's Encrypt
- [ ] Adjust rate limiting for expected traffic
- [ ] Configure CORS for your frontend domains
- [ ] Set up monitoring and alerting
- [ ] Review security headers
- [ ] Test failover scenarios
- [ ] Configure backup ingress controller (optional)

## References

- [Traefik Documentation](https://doc.traefik.io/traefik/)
- [Traefik Kubernetes Ingress](https://doc.traefik.io/traefik/providers/kubernetes-ingress/)
- [cert-manager Documentation](https://cert-manager.io/docs/)
