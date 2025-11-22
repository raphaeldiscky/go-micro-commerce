# Backend API Ingress

This directory contains Kubernetes manifests for the backend API HTTPS ingress configuration, deployed via ArgoCD GitOps workflow.

## Overview

Provides secure HTTPS access to backend services at `api.discky.com` with automated TLS certificate management via cert-manager and Let's Encrypt.

## Architecture

### Path-Based Routing

Single domain (`api.discky.com`) with path-based routing to multiple backend services:

- **GraphQL Federation**: `/graph` -> apollo-router:4000
  - GraphQL queries and mutations: `https://api.discky.com/graph`
  - WebSocket subscriptions: `wss://api.discky.com/graph/subscriptions/ws`
  - SSE subscriptions: `https://api.discky.com/graph/subscriptions/sse`
- **REST + gRPC (Connect-RPC)**: `/` -> api-gateway:8080

### TLS Certificate

- **Domain**: api.discky.com
- **Issuer**: cert-manager with Let's Encrypt production
- **Secret**: api-backend-tls
- **Auto-renewal**: Certificates renewed automatically 30 days before expiration

### Middlewares

1. **CORS Middleware** (`api-cors`):
   - Allows frontend origin: `https://go.micro.commerce.discky.com`
   - Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
   - Credentials: Enabled
   - Max Age: 3600 seconds
   - WebSocket Support: Includes headers for GraphQL subscription handshake (Connection, Upgrade, Sec-WebSocket-\*)

2. **Security Headers** (`security-headers`):
   - X-Frame-Options: DENY
   - X-Content-Type-Options: nosniff
   - X-XSS-Protection: 1; mode=block
   - HSTS: 31536000 seconds (1 year)
   - Referrer-Policy: strict-origin-when-cross-origin
   - Permissions-Policy: Restricted geolocation, microphone, camera

## Files

- **kustomization.yaml**: Kustomize configuration with common labels
- **ingress.yaml**: HTTPS ingress with path-based routing and TLS configuration
- **middlewares.yaml**: Traefik CORS and security headers middlewares

## Deployment

### Prerequisites

- ArgoCD installed and configured in the cluster
- ApplicationSet with directory generator watching `deployments/k8s/apps/`
- cert-manager installed with `letsencrypt-prod` ClusterIssuer
- Traefik ingress controller running
- DNS record for `api.discky.com` pointing to Traefik LoadBalancer IP

### Deploy via ArgoCD

1. **Commit manifests to git**:

   ```bash
   git add deployments/k8s/apps/api-backend/
   git commit -m "feat(k8s): add backend API HTTPS ingress with cert-manager"
   git push
   ```

2. **ArgoCD auto-discovery**:
   - ApplicationSet automatically discovers the new `api-backend` directory
   - Creates an ArgoCD Application resource
   - Syncs manifests to the cluster

3. **Monitor deployment**:

   ```bash
   # Watch ArgoCD applications
   kubectl get applications -n argocd -w

   # Check ingress resource
   kubectl get ingress api-backend-ingress -n default

   # Check middlewares
   kubectl get middleware -n default
   ```

### Certificate Provisioning

After ingress is created, cert-manager automatically:

1. Detects the `cert-manager.io/cluster-issuer` annotation
2. Creates a Certificate resource
3. Performs HTTP-01 challenge via Let's Encrypt
4. Stores the TLS certificate in the `api-backend-tls` secret

**Timeline**: 2-3 minutes for initial certificate issuance

**Check certificate status**:

```bash
# List certificates
kubectl get certificates -n default

# Detailed certificate status
kubectl describe certificate api-backend-tls -n default

# Check TLS secret
kubectl get secret api-backend-tls -n default
```

## Verification

### 1. Check Ingress Configuration

```bash
kubectl get ingress api-backend-ingress -n default -o yaml
```

Expected output should show:

- Host: api.discky.com
- TLS secret: api-backend-tls
- Paths: /graph and /
- Middleware annotations

### 2. Test HTTPS Access

```bash
# Test API Gateway (REST endpoint)
curl -v https://api.discky.com/health

# Test Apollo Router (GraphQL endpoint)
curl -v https://api.discky.com/graph \
  -H "Content-Type: application/json" \
  -d '{"query":"{ __typename }"}'
```

### 3. Test from Frontend

Access your frontend at `https://go.micro.commerce.discky.com` and verify:

- No certificate errors (`ERR_CERT_AUTHORITY_INVALID`)
- CORS headers present in API responses
- GraphQL and REST requests successful

### 4. Verify CORS Headers

```bash
curl -v -H "Origin: https://go.micro.commerce.discky.com" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -X OPTIONS https://api.discky.com/graph
```

Expected response headers:

- `Access-Control-Allow-Origin: https://go.micro.commerce.discky.com`
- `Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS`
- `Access-Control-Allow-Credentials: true`

## Troubleshooting

### Certificate Not Issued

**Symptoms**: Certificate stuck in "Pending" state

**Debug**:

```bash
# Check certificate status
kubectl describe certificate api-backend-tls -n default

# Check certificate request
kubectl get certificaterequest -n default

# Check challenge
kubectl get challenge -n default

# Check cert-manager logs
kubectl logs -n cert-manager deployment/cert-manager -f
```

**Common causes**:

- DNS not resolving to correct IP
- Traefik not routing HTTP-01 challenge path
- Firewall blocking port 80/443

### CORS Errors

**Symptoms**: Browser console shows CORS errors

**Debug**:

```bash
# Check middleware is applied
kubectl get ingress api-backend-ingress -n default -o jsonpath='{.metadata.annotations}'

# Verify middleware exists
kubectl get middleware api-cors -n default -o yaml
```

**Common causes**:

- Middleware not referenced in ingress annotations
- Origin mismatch (check frontend domain matches CORS config)
- Middleware in wrong namespace

### Path Routing Issues

**Symptoms**: Requests to /graph going to wrong service

**Debug**:

```bash
# Check ingress paths order
kubectl get ingress api-backend-ingress -n default -o yaml | grep -A 10 paths

# Check service endpoints
kubectl get endpoints apollo-router -n default
kubectl get endpoints api-gateway -n default
```

**Important**: GraphQL path (`/graph`) must be listed **before** the catch-all path (`/`) in ingress.yaml

### Service Not Found

**Symptoms**: 503 Service Unavailable errors

**Debug**:

```bash
# Check if services exist
kubectl get svc apollo-router -n default
kubectl get svc api-gateway -n default

# Check service endpoints
kubectl get endpoints apollo-router -n default
kubectl get endpoints api-gateway -n default

# Check pod status
kubectl get pods -n default -l app=apollo-router
kubectl get pods -n default -l app=api-gateway
```

**Fix**: Ensure apollo-router and api-gateway services are deployed and running

## Maintenance

### Update CORS Origins

To add or modify allowed origins:

1. Edit `middlewares.yaml`
2. Update `accessControlAllowOriginList` in the `api-cors` middleware
3. Commit and push changes
4. ArgoCD will automatically sync the update

### Certificate Renewal

Certificates are automatically renewed by cert-manager 30 days before expiration. No manual intervention required.

**Monitor renewal**:

```bash
# Check certificate expiration
kubectl get certificate api-backend-tls -n default -o jsonpath='{.status.notAfter}'

# Watch cert-manager renewal logs
kubectl logs -n cert-manager deployment/cert-manager -f | grep renewal
```

### Modify Path Routing

To change backend service routing:

1. Edit `ingress.yaml`
2. Update the `paths` section under `spec.rules[0].http`
3. Ensure more specific paths are listed before general paths
4. Commit and push changes

## Related Components

- **cert-manager**: TLS certificate lifecycle management
- **Traefik**: Ingress controller with middleware support
- **ArgoCD**: GitOps continuous deployment
- **Cloudflare DNS**: DNS record management for api.discky.com
- **apollo-router**: GraphQL Federation gateway
- **api-gateway**: REST and gRPC-Connect API gateway

## References

- [cert-manager Documentation](https://cert-manager.io/docs/)
- [Traefik Middleware](https://doc.traefik.io/traefik/middlewares/overview/)
- [Kubernetes Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)
- [ArgoCD ApplicationSet](https://argo-cd.readthedocs.io/en/stable/user-guide/application-set/)
- [Let's Encrypt HTTP-01 Challenge](https://letsencrypt.org/docs/challenge-types/)
