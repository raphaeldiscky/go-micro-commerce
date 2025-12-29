import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { PATH_ROOT } from '@/constants'
import type { FileRoutesByPath } from '@tanstack/react-router'
import { createFileRoute } from '@tanstack/react-router'
import {
  Bell,
  Building,
  CreditCard,
  ExternalLink,
  FileText,
  Lock,
  Mail,
  MessageCircle,
  Package,
  Radio,
  Rocket,
  RotateCw,
  Search,
  ShoppingCart,
  Truck,
  Wrench,
} from 'lucide-react'

export const Route = createFileRoute(
  PATH_ROOT.services as keyof FileRoutesByPath,
)({
  component: ServicesPage,
})

function ServicesPage() {
  const services = [
    {
      description:
        'Unified entry point with rate limiting, circuit breaker, and JWT validation',
      details:
        'Routes requests to microservices with cross-cutting concerns like authentication, rate limiting, and circuit breaker patterns for fault tolerance.',
      features: [
        'Rate Limiting',
        'Circuit Breaker',
        'JWT Validation',
        'Service Discovery',
        'Request Routing',
        'CORS Management',
      ],
      icon: ExternalLink,
      name: 'API Gateway',
      patterns: [
        'Circuit Breaker',
        'Rate Limiting',
        'Proxy',
        'Service Discovery',
      ],
      technologies: ['Go', 'Traefik'],
    },
    {
      description:
        'JWT RS256 authentication with refresh token rotation and email verification',
      details:
        'Secure authentication using asymmetric RS256 algorithm with short-lived access tokens and automatic refresh token rotation.',
      features: [
        'RS256 JWT',
        'Refresh Token Rotation',
        'Email Verification',
        'Session Management',
        'Rate Limiting',
        'Password Security',
      ],
      icon: Lock,
      name: 'Auth Service',
      patterns: ['Clean Architecture', 'JWT Bearer'],
      technologies: ['Go', 'PostgreSQL', 'JWT', 'Bcrypt'],
    },
    {
      description:
        'Catalog management with gRPC, optimistic locking, and Redis caching',
      details:
        'Product CRUD with version-based optimistic locking to prevent lost updates. Cache-aside pattern with Redis for high-read performance.',
      features: [
        'gRPC API',
        'Optimistic Locking',
        'Redis Cache',
        'Stock Reservation',
        'Search Integration',
        'Price Management',
      ],
      icon: Package,
      name: 'Product Service',
      patterns: ['Clean Architecture', 'Cache-Aside', 'Optimistic Lock'],
      technologies: ['Go', 'PostgreSQL', 'gRPC', 'Redis'],
    },
    {
      description:
        'Saga orchestration with dual implementation: PostgreSQL and Temporal',
      details:
        'Complex order workflows with both custom PostgreSQL-based saga and Temporal-managed workflows. Asynq for payment reminders and expiration.',
      features: [
        'Saga Orchestration',
        'Temporal Workflows',
        'Asynq Scheduling',
        'Compensation Logic',
        'Order Expiration',
        'Payment Reminders',
      ],
      icon: FileText,
      name: 'Order Service',
      patterns: [
        'Clean Architecture',
        'Saga Orchestration',
        'Outbox',
        'Compensation',
      ],
      technologies: ['Go', 'PostgreSQL', 'Temporal', 'Kafka', 'Asynq'],
    },
    {
      description:
        'Stripe integration with idempotent webhooks and secure processing',
      details:
        'Payment gateway factory pattern supporting Stripe. Idempotent processing with keys and secure webhook verification.',
      features: [
        'Stripe Gateway',
        'Idempotent Webhooks',
        'Refund Processing',
        'Payment Analytics',
        'Secure Verification',
        'Multi-currency',
      ],
      icon: CreditCard,
      name: 'Payment Service',
      patterns: ['Factory', 'Idempotent Receiver', 'Outbox', 'Inbox'],
      technologies: ['Go', 'PostgreSQL', 'Stripe', 'Kafka'],
    },
    {
      description:
        'Shipping coordination with cost calculation and delivery tracking',
      details:
        'Manages fulfillment lifecycle from order confirmation to delivery. Calculates shipping costs and coordinates with providers.',
      features: [
        'Shipping Cost',
        'Delivery Tracking',
        'Status Updates',
        'Provider Integration',
        'ETA Calculation',
        'Fulfillment Status',
      ],
      icon: Truck,
      name: 'Fulfillment Service',
      patterns: ['Clean Architecture', 'Outbox', 'Inbox'],
      technologies: ['Go', 'PostgreSQL', 'Kafka', 'gRPC'],
    },
    {
      description:
        'Multi-channel delivery via SSE, email, and SMS with templates',
      details:
        'Real-time push notifications with Server-Sent Events. Async email processing and SMS support with template management.',
      features: [
        'SSE Push',
        'Email Delivery',
        'SMS Support',
        'Template Engine',
        'Delivery Tracking',
        'Redis Pub/Sub',
      ],
      icon: Bell,
      name: 'Notification Service',
      patterns: ['Clean Architecture', 'Pub/Sub', 'Template Method', 'Inbox'],
      technologies: ['Go', 'PostgreSQL', 'Kafka', 'Redis', 'SMTP'],
    },
    {
      description:
        'Real-time messaging via WebSocket with Redis Pub/Sub broadcasting',
      details:
        'Bi-directional WebSocket communication with Redis Pub/Sub for cross-instance message broadcasting. Supports typing indicators and presence.',
      features: [
        'WebSocket',
        'Redis Pub/Sub',
        'Typing Indicators',
        'Online Status',
        'Message History',
        'Group Chats',
      ],
      icon: MessageCircle,
      name: 'Chat Service',
      patterns: ['Clean Architecture', 'Pub/Sub'],
      technologies: ['Go', 'PostgreSQL', 'WebSocket', 'Redis'],
    },
    {
      description:
        'Full-text search with Elasticsearch indexing and faceted filters',
      details:
        'Real-time document indexing via Kafka events. Advanced full-text search with fuzzy matching, faceted filters, and relevance scoring.',
      features: [
        'Elasticsearch',
        'Fuzzy Search',
        'Faceted Filters',
        'Auto-suggestions',
        'Relevance Scoring',
        'Real-time Indexing',
      ],
      icon: Search,
      name: 'Search Service',
      patterns: ['Clean Architecture', 'Inbox'],
      technologies: ['Go', 'Elasticsearch', 'Kafka', 'Redis'],
    },
    {
      description:
        'Shopping cart with checkout sessions and Asynq task scheduling',
      details:
        'Cart lifecycle management with checkout session generation. Promotional code validation and Asynq for cart abandonment recovery.',
      features: [
        'Cart Management',
        'Checkout Sessions',
        'Promo Codes',
        'Cart Abandonment',
        'gRPC API',
        'Asynq Tasks',
      ],
      icon: ShoppingCart,
      name: 'Cart Service',
      patterns: ['Clean Architecture', 'Outbox'],
      technologies: ['Go', 'PostgreSQL', 'Redis', 'Asynq', 'gRPC'],
    },
    {
      description:
        'Apollo Federation for unified GraphQL schema with subscriptions',
      details:
        'GraphQL gateway using Apollo Router for schema federation. Supports queries, mutations, and real-time subscriptions.',
      features: [
        'Apollo Router',
        'Schema Federation',
        'Subscriptions',
        'Type-safe API',
        'JWT Auth',
        'Query Planning',
      ],
      icon: Mail,
      name: 'GraphQL Gateway',
      patterns: ['Federation', 'API Gateway'],
      technologies: ['Apollo Router', 'GraphQL', 'gqlgen'],
    },
  ]

  const architecturalPatterns = [
    {
      description:
        'Bounded contexts with clean architecture and layered design',
      icon: Building,
      name: 'Domain-Driven Design',
    },
    {
      description: 'Kafka KRaft cluster with outbox pattern and DLQ handling',
      icon: Radio,
      name: 'Event-Driven Architecture',
    },
    {
      description:
        'Dual implementation with PostgreSQL-based and Temporal workflows',
      icon: RotateCw,
      name: 'Saga Orchestration',
    },
    {
      description: 'ArgoCD with Kustomize for declarative deployments',
      icon: Rocket,
      name: 'GitOps',
    },
  ]

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-white dark:from-gray-900 dark:to-gray-800">
      <div className="container mx-auto px-4 py-12">
        {/* Header */}
        <div className="text-center mb-16">
          <h1 className="text-4xl font-bold text-gray-900 dark:text-white mb-4 flex items-center justify-center gap-2">
            <Wrench className="h-10 w-10" />
            Microservices Architecture
          </h1>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-3xl mx-auto">
            11 Go microservices with saga orchestration, Kafka KRaft, Temporal
            workflows, and production-grade Kubernetes deployment.
          </p>
        </div>

        {/* Architectural Patterns */}
        <section className="mb-16">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-8 text-center">
            Key Architectural Patterns
          </h2>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {architecturalPatterns.map((pattern) => (
              <Card
                className="hover:shadow-lg transition-shadow"
                key={pattern.name}
              >
                <CardHeader className="pb-3">
                  <div className="flex items-center space-x-2">
                    <pattern.icon className="h-6 w-6" />
                    <CardTitle className="text-lg">{pattern.name}</CardTitle>
                  </div>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-gray-600 dark:text-gray-300">
                    {pattern.description}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        </section>

        {/* Services Grid */}
        <section>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-8 text-center">
            Service Catalog
          </h2>
          <div className="grid gap-8 lg:grid-cols-1 xl:grid-cols-2">
            {services.map((service) => (
              <Card
                className="hover:shadow-xl transition-all duration-300 border-l-4 border-blue-500"
                key={service.name}
              >
                <CardHeader>
                  <div className="flex items-center space-x-3">
                    <service.icon className="h-8 w-8" />
                    <CardTitle className="text-xl">{service.name}</CardTitle>
                  </div>
                  <p className="text-gray-600 dark:text-gray-300 mt-2">
                    {service.description}
                  </p>
                </CardHeader>

                <CardContent className="space-y-6">
                  <p className="text-sm text-gray-700 dark:text-gray-300">
                    {service.details}
                  </p>

                  {/* Features */}
                  <div>
                    <h4 className="font-semibold text-gray-900 dark:text-white mb-2">
                      Key Features
                    </h4>
                    <div className="flex flex-wrap gap-2">
                      {service.features.map((feature) => (
                        <Badge
                          className="text-xs"
                          key={feature}
                          variant="secondary"
                        >
                          {feature}
                        </Badge>
                      ))}
                    </div>
                  </div>

                  {/* Technologies */}
                  <div>
                    <h4 className="font-semibold text-gray-900 dark:text-white mb-2">
                      Technologies
                    </h4>
                    <div className="flex flex-wrap gap-2">
                      {service.technologies.map((tech) => (
                        <Badge className="text-xs" key={tech} variant="outline">
                          {tech}
                        </Badge>
                      ))}
                    </div>
                  </div>

                  {/* Patterns */}
                  <div>
                    <h4 className="font-semibold text-gray-900 dark:text-white mb-2">
                      Design Patterns
                    </h4>
                    <div className="flex flex-wrap gap-2">
                      {service.patterns.map((pattern) => (
                        <Badge
                          className="text-xs bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200"
                          key={pattern}
                        >
                          {pattern}
                        </Badge>
                      ))}
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </section>

        {/* Infrastructure Note */}
        <section className="mt-16 text-center">
          <Card className="max-w-4xl mx-auto bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800">
            <CardContent className="p-8">
              <h3 className="text-xl font-bold text-gray-900 dark:text-white mb-4 flex items-center justify-center gap-2">
                <Building className="h-6 w-6" />
                Infrastructure & Deployment
              </h3>
              <p className="text-gray-700 dark:text-gray-300 mb-4">
                Production deployment on GKE with Terraform IaC, GitOps via
                ArgoCD, and LGTM observability stack. Frontend deployed on
                Cloudflare Pages for global edge distribution.
              </p>
              <div className="flex flex-wrap justify-center gap-2">
                <Badge variant="outline">Kubernetes</Badge>
                <Badge variant="outline">Terraform</Badge>
                <Badge variant="outline">ArgoCD</Badge>
                <Badge variant="outline">Kafka KRaft</Badge>
                <Badge variant="outline">Redis Cluster</Badge>
                <Badge variant="outline">Temporal</Badge>
                <Badge variant="outline">LGTM Stack</Badge>
                <Badge variant="outline">Cloudflare Pages</Badge>
              </div>
            </CardContent>
          </Card>
        </section>
      </div>
    </div>
  )
}
