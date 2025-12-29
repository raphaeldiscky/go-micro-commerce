import {
  ArgoCDIcon,
  CloudflareIcon,
  DockerIcon,
  ElasticsearchIcon,
  GoogleCloudIcon,
  GrafanaIcon,
  GraphQLIcon,
  GrpcIcon,
  OpenAPIIcon,
  PrometheusIcon,
  TemporalIcon,
  TerraformIcon,
  TraefikIcon,
  ViteIcon,
} from '@/components/icons/tech-icons'
import {
  Go,
  Kafka,
  Kubernetes,
  PostgreSQL,
  React,
  Redis,
  TypeScript,
} from 'developer-icons'
import {
  Bell,
  CreditCard,
  ExternalLink,
  FileText,
  Lock,
  Mail,
  MessageCircle,
  Package,
  Search,
  ShoppingCart,
  Truck,
} from 'lucide-react'
import type { ServiceInfo, TechnologyInfo, TechnologyLink } from './types'
import { TECH_URLS } from './urls'

// Microservices information
export const SERVICES: Array<ServiceInfo> = [
  {
    description: 'Unified entry point with rate limiting and circuit breaker',
    features: ['Rate Limiting', 'Circuit Breaker', 'JWT Validation'],
    icon: ExternalLink,
    name: 'API Gateway',
  },
  {
    description: 'JWT RS256 authentication with refresh token rotation',
    features: ['RS256 JWT', 'Refresh Tokens', 'Email Verification'],
    icon: Lock,
    name: 'Auth Service',
  },
  {
    description: 'Catalog management with gRPC and optimistic locking',
    features: ['gRPC API', 'Optimistic Locking', 'Redis Cache'],
    icon: Package,
    name: 'Product Service',
  },
  {
    description: 'Saga orchestration with Temporal and Asynq scheduling',
    features: ['Saga Pattern', 'Temporal Workflows', 'Asynq Tasks'],
    icon: FileText,
    name: 'Order Service',
  },
  {
    description: 'Stripe integration with idempotent webhook handling',
    features: ['Stripe Gateway', 'Webhooks', 'Idempotency'],
    icon: CreditCard,
    name: 'Payment Service',
  },
  {
    description: 'Shipping coordination and delivery tracking',
    features: ['Shipping Cost', 'Delivery Tracking', 'Fulfillment'],
    icon: Truck,
    name: 'Fulfillment Service',
  },
  {
    description: 'Multi-channel delivery via SSE, email, and SMS',
    features: ['SSE Push', 'Email', 'SMS'],
    icon: Bell,
    name: 'Notification Service',
  },
  {
    description: 'Real-time messaging via WebSocket and Redis Pub/Sub',
    features: ['WebSocket', 'Redis Pub/Sub', 'Typing Indicators'],
    icon: MessageCircle,
    name: 'Chat Service',
  },
  {
    description: 'Full-text search with Elasticsearch indexing',
    features: ['Elasticsearch', 'Fuzzy Search', 'Faceted Filters'],
    icon: Search,
    name: 'Search Service',
  },
  {
    description: 'Shopping cart with checkout sessions and task scheduling',
    features: ['Cart Management', 'Checkout Sessions', 'Asynq Tasks'],
    icon: ShoppingCart,
    name: 'Cart Service',
  },
  {
    description: 'Apollo Federation for unified GraphQL schema',
    features: ['Apollo Router', 'Schema Federation', 'Subscriptions'],
    icon: Mail,
    name: 'GraphQL Gateway',
  },
]

// Technology stack information (sorted by relevance)
export const TECHNOLOGIES: Array<TechnologyInfo> = [
  // Core Languages & Frameworks
  { description: 'Backend Services', icon: Go, name: 'Go' },
  { description: 'Frontend UI', icon: React, name: 'React' },
  { description: 'Build Tool', icon: ViteIcon, name: 'Vite' },
  { description: 'Type Safety', icon: TypeScript, name: 'TypeScript' },
  // API Layer
  { description: 'Service Communication', icon: GrpcIcon, name: 'gRPC' },
  { description: 'API Gateway', icon: GraphQLIcon, name: 'GraphQL' },
  { description: 'API Specification', icon: OpenAPIIcon, name: 'OpenAPI' },
  // Data Layer
  { description: 'Primary Database', icon: PostgreSQL, name: 'PostgreSQL' },
  { description: 'Cache & Pub/Sub', icon: Redis, name: 'Redis Cluster' },
  { description: 'Event Streaming', icon: Kafka, name: 'Kafka KRaft' },
  {
    description: 'Search Engine',
    icon: ElasticsearchIcon,
    name: 'Elasticsearch',
  },
  // Workflow & Orchestration
  { description: 'Workflow Engine', icon: TemporalIcon, name: 'Temporal' },
  // Container & Orchestration
  { description: 'Containerization', icon: DockerIcon, name: 'Docker' },
  { description: 'Orchestration', icon: Kubernetes, name: 'Kubernetes' },
  // Infrastructure & DevOps
  { description: 'Load Balancer', icon: TraefikIcon, name: 'Traefik' },
  {
    description: 'Infrastructure as Code',
    icon: TerraformIcon,
    name: 'Terraform',
  },
  { description: 'GitOps', icon: ArgoCDIcon, name: 'ArgoCD' },
  // Observability
  { description: 'Metrics', icon: PrometheusIcon, name: 'Prometheus' },
  { description: 'Dashboards', icon: GrafanaIcon, name: 'Grafana' },
  // Cloud & Edge
  {
    description: 'Cloud Platform',
    icon: GoogleCloudIcon,
    name: 'Google Cloud',
  },
  { description: 'Edge Network', icon: CloudflareIcon, name: 'Cloudflare' },
]

// Technology links for footer
export const TECHNOLOGY_LINKS: Array<TechnologyLink> = [
  { name: 'Go', url: TECH_URLS.GO },
  { name: 'React', url: TECH_URLS.REACT },
  { name: 'PostgreSQL', url: TECH_URLS.POSTGRESQL },
  { name: 'Redis', url: TECH_URLS.REDIS },
  { name: 'Kafka', url: TECH_URLS.KAFKA },
  { name: 'Docker', url: TECH_URLS.DOCKER },
]
