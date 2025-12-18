# Traffic Flow

Example 1: GraphQL Query

1. Client → https://api.discky.com/graph
2. Traefik → Looks up Gateway "shared-gateway"
3. Traefik → Finds HTTPRoute "graphql-route" with path /graph
4. Traefik → Checks ReferenceGrant for permission
5. Traefik → Routes to apollo-router.graphql.svc.cluster.local:4000
6. Apollo Router → Federates query to subgraphs (auth, chat, order, etc.)
7. Response ← Back through the chainrem

Example 2: REST API Call

1. Client → https://api.discky.com/api/products
2. Traefik → Looks up Gateway "shared-gateway"
3. Traefik → Finds HTTPRoute "api-gateway-route" with path /
4. Traefik → Routes to api-gateway.gateway.svc.cluster.local:8080
5. API Gateway → Discovers product-service via Consul/K8s DNS
6. API Gateway → Routes to product-service.application.svc.cluster.local:8080
7. Response ← Back through the chain
