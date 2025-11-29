# Gateway API Architecture

## Overview

The infrastructure uses **Kubernetes Gateway API** for modern, secure, cross-namespace HTTP routing with role-based access control. This is the successor to the Ingress API, providing better multi-tenancy, security, and protocol support.

## Architecture Diagram

```
Client (HTTPS)
    ↓
api.discky.com
    ↓
┌─────────────────────────────────────────────────┐
│ Traefik Gateway Controller (traefik namespace)      │
│                                                     │
│ Gateway: shared-gateway                             │
│ - HTTPS Listener: *.discky.com:443                  │
│ - HTTP Listener: *.discky.com:80                    │
│ - Allows routes from all namespaces                 │
└─────────────────────────────────────────────────┘
    ↓
    ├─── HTTPRoute (gateway namespace) ───► API Gateway
    │    Path: /
    │    Backend: api-gateway:8080
    │
    └─── HTTPRoute (graphql namespace) ───► Apollo Router
         Path: /graph
         Backend: apollo-router:4000
         + ReferenceGrant (security)
```

## Traffic Flow

### GraphQL Query Example

```
1. Client → https://api.discky.com/graph
2. Traefik → Looks up Gateway "shared-gateway"
3. Traefik → Finds HTTPRoute "graphql-route" with path /graph
4. Traefik → Checks ReferenceGrant for permission ✓
5. Traefik → Routes to apollo-router.graphql.svc.cluster.local:4000
6. Apollo Router → Federates query to subgraphs (auth, chat, order, etc.)
7. Response ← Back through the chain
```

### REST API Call Example

```
1. Client → https://api.discky.com/api/products
2. Traefik → Looks up Gateway "shared-gateway"
3. Traefik → Finds HTTPRoute "api-gateway-route" with path /
4. Traefik → Routes to api-gateway.gateway.svc.cluster.local:8080
5. API Gateway → Discovers product-service via Consul/K8s DNS
6. API Gateway → Routes to product-service.application.svc.cluster.local:8080
7. Response ← Back through the chain
```

## Key Components

### Gateway (Shared Entry Point)

- **Namespace:** `traefik`
- **Name:** `shared-gateway`
- **Purpose:** Centralized traffic entry point for all HTTPRoutes
- **Listeners:** HTTPS (443) and HTTP (80)
- **Security:** Allows routes from all namespaces with explicit ReferenceGrant

### HTTPRoute (API Gateway)

- **Namespace:** `gateway`
- **Path:** `/` (all traffic except /graph)
- **Backend:** api-gateway:8080
- **Purpose:** Routes REST, gRPC, WebSocket, SSE traffic to custom gateway

### HTTPRoute (GraphQL)

- **Namespace:** `graphql`
- **Path:** `/graph`
- **Backend:** apollo-router:4000
- **Purpose:** Routes GraphQL federation traffic to Apollo Router

### ReferenceGrant (Security)

- **Namespace:** `graphql`
- **Purpose:** Explicitly permits cross-namespace routing from Gateway to apollo-router service
- **Security Model:** Default-deny, explicit-allow for cross-namespace access

## Security Model

Gateway API uses **ReferenceGrant** for explicit, auditable cross-namespace permissions:

```
Without ReferenceGrant:
  HTTPRoute (graphql ns) → service (graphql ns)    ✓ ALLOWED
  HTTPRoute (graphql ns) → service (gateway ns)    ✗ DENIED

With ReferenceGrant:
  HTTPRoute references Gateway (traefik ns)        ✓ ALLOWED (via ReferenceGrant)
  Gateway routes to apollo-router (graphql ns)     ✓ ALLOWED (via ReferenceGrant)
```

## Benefits Over Ingress API

- **Cross-namespace routing** with ReferenceGrant security
- **Protocol-specific routing** (HTTP, gRPC, TCP, UDP)
- **Header-based routing** and traffic splitting
- **Better multi-tenancy** with clear RBAC boundaries
- **Future-proof** (Gateway API is GA, Ingress is stable but frozen)

## References

- [Gateway API Documentation](https://gateway-api.sigs.k8s.io/)
- [Traefik Gateway API Guide](https://doc.traefik.io/traefik/routing/providers/kubernetes-gateway/)
- [ReferenceGrant Security Model](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1beta1.ReferenceGrant)
