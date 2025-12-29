// Wrapper components for custom tech SVGs
export function ElasticsearchIcon({ className }: { className?: string }) {
  return (
    <img
      alt="Elasticsearch"
      className={className}
      src="/svg/elastic-search.svg"
    />
  )
}

export function GrpcIcon({ className }: { className?: string }) {
  return <img alt="gRPC" className={className} src="/svg/grpc.svg" />
}

export function GrafanaIcon({ className }: { className?: string }) {
  return <img alt="Grafana" className={className} src="/svg/grafana.svg" />
}

export function GraphQLIcon({ className }: { className?: string }) {
  return <img alt="GraphQL" className={className} src="/svg/grapqhl.svg" />
}

export function TerraformIcon({ className }: { className?: string }) {
  return <img alt="Terraform" className={className} src="/svg/terraform.svg" />
}

export function ArgoCDIcon({ className }: { className?: string }) {
  return <img alt="ArgoCD" className={className} src="/svg/argocd.svg" />
}

export function OpenAPIIcon({ className }: { className?: string }) {
  return <img alt="OpenAPI" className={className} src="/svg/open-api.svg" />
}

export function GoogleCloudIcon({ className }: { className?: string }) {
  return (
    <img alt="Google Cloud" className={className} src="/svg/google-cloud.svg" />
  )
}

export function CloudflareIcon({ className }: { className?: string }) {
  return (
    <img alt="Cloudflare" className={className} src="/svg/cloudflare.svg" />
  )
}

export function DockerIcon({ className }: { className?: string }) {
  return <img alt="Docker" className={className} src="/svg/docker.svg" />
}

export function ViteIcon({ className }: { className?: string }) {
  return <img alt="Vite" className={className} src="/svg/vite.svg" />
}

export function TemporalIcon({ className }: { className?: string }) {
  return (
    <img
      alt="Temporal"
      className={`${className} scale-300`}
      src="/svg/temporal.svg"
    />
  )
}

export function TraefikIcon({ className }: { className?: string }) {
  return <img alt="Traefik" className={className} src="/svg/traefik.svg" />
}

export function PrometheusIcon({ className }: { className?: string }) {
  return (
    <img alt="Prometheus" className={className} src="/svg/prometheus.svg" />
  )
}
