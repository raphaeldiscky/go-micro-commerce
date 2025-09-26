import { createFileRoute } from '@tanstack/react-router'
import {
  ArrowLeftRight,
  Building,
  CreditCard,
  ExternalLink,
  FileText,
  Lock,
  Mail,
  MessageCircle,
  Package,
  Radio,
  RotateCw,
  Search,
  Wrench,
  Zap,
} from 'lucide-react'
import { Badge } from '../components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '../components/ui/card'

export const Route = createFileRoute('/services')({
  component: ServicesPage,
})

function ServicesPage() {
  const services = [
    {
      description:
        'Central entry point for all client requests with comprehensive traffic management',
      details:
        'Handles all incoming requests and routes them to appropriate microservices. Provides cross-cutting concerns like authentication, logging, and monitoring.',
      features: [
        'Load Balancing',
        'Rate Limiting',
        'Service Discovery',
        'Request Routing',
        'Circuit Breaker',
        'Monitoring',
      ],
      icon: ExternalLink,
      name: 'API Gateway',
      patterns: ['Gateway Pattern', 'Circuit Breaker', 'Service Discovery'],
      port: ':8080',
      technologies: ['Traefik', 'Consul', 'Go'],
    },
    {
      description:
        'JWT-based authentication and authorization with role management',
      details:
        'Manages user authentication, authorization, and access control across all services. Implements secure password handling and token-based authentication.',
      features: [
        'JWT Authentication',
        'User Registration',
        'Role-based Access Control',
        'Token Refresh',
        'Session Management',
        'Password Security',
      ],
      icon: Lock,
      name: 'Auth Service',
      patterns: ['JWT Pattern', 'Repository Pattern', 'Clean Architecture'],
      port: ':8081',
      technologies: ['Go', 'PostgreSQL', 'JWT', 'bcrypt'],
    },
    {
      description:
        'Product catalog management with both REST and gRPC interfaces',
      details:
        'Handles product catalog operations with dual API support. Integrates with search service for enhanced product discovery.',
      features: [
        'Product CRUD',
        'gRPC API',
        'Inventory Tracking',
        'Category Management',
        'Search Integration',
        'Price Management',
      ],
      icon: Package,
      name: 'Product Service',
      patterns: ['Domain-Driven Design', 'gRPC Pattern', 'Repository Pattern'],
      port: ':8082',
      technologies: ['Go', 'PostgreSQL', 'gRPC', 'Protocol Buffers'],
    },
    {
      description:
        'Complex order processing with saga orchestration and workflow management',
      details:
        'Orchestrates complex order workflows using saga patterns. Manages distributed transactions with compensation logic for failure scenarios.',
      features: [
        'Order Management',
        'Saga Orchestration',
        'Temporal Workflows',
        'Compensation Logic',
        'State Management',
        'Event Sourcing',
      ],
      icon: FileText,
      name: 'Order Service',
      patterns: [
        'Saga Pattern',
        'Event Sourcing',
        'Temporal Workflows',
        'CQRS',
      ],
      port: ':8083',
      technologies: ['Go', 'PostgreSQL', 'Temporal', 'Kafka'],
    },
    {
      description:
        'Secure payment processing with transaction management and fraud prevention',
      details:
        'Handles secure payment processing with multiple payment methods. Implements fraud detection and comprehensive transaction management.',
      features: [
        'Payment Processing',
        'Transaction History',
        'Refund Management',
        'Fraud Detection',
        'Payment Methods',
        'Webhooks',
      ],
      icon: CreditCard,
      name: 'Payment Service',
      patterns: ['Payment Gateway Pattern', 'Idempotency', 'Retry Pattern'],
      port: ':8084',
      technologies: ['Go', 'PostgreSQL', 'Stripe API', 'Redis'],
    },
    {
      description:
        'Inventory management and order fulfillment with shipping integration',
      details:
        'Manages inventory levels and fulfillment processes. Coordinates with shipping providers and maintains real-time stock information.',
      features: [
        'Inventory Management',
        'Order Fulfillment',
        'Shipping Integration',
        'Stock Tracking',
        'Warehouse Management',
        'Delivery Updates',
      ],
      icon: Package,
      name: 'Fulfillment Service',
      patterns: ['Event-Driven Architecture', 'Saga Pattern', 'Domain Events'],
      port: ':8085',
      technologies: ['Go', 'PostgreSQL', 'Redis', 'Kafka'],
    },
    {
      description:
        'Real-time messaging with WebSocket connections and message persistence',
      details:
        'Provides real-time communication capabilities with persistent message storage. Supports multiple concurrent connections and room-based messaging.',
      features: [
        'WebSocket Support',
        'Real-time Messaging',
        'Multi-user Rooms',
        'Message History',
        'Connection Management',
        'Scalable Architecture',
      ],
      icon: MessageCircle,
      name: 'Chat Service',
      patterns: ['WebSocket Pattern', 'Pub/Sub', 'Connection Pooling'],
      port: ':9088',
      technologies: ['Go', 'WebSocket', 'Redis', 'PostgreSQL'],
    },
    {
      description: 'Event-driven notifications with multiple delivery channels',
      details:
        'Handles all system notifications across multiple channels. Processes events from other services to trigger appropriate notifications.',
      features: [
        'Email Notifications',
        'Push Notifications',
        'SMS Support',
        'Template Management',
        'Delivery Tracking',
        'Event-driven Triggers',
      ],
      icon: Mail,
      name: 'Notification Service',
      patterns: [
        'Event-Driven Architecture',
        'Template Pattern',
        'Observer Pattern',
      ],
      port: ':8086',
      technologies: ['Go', 'Kafka', 'SMTP', 'Redis'],
    },
    {
      description:
        'Advanced search capabilities with Elasticsearch integration',
      details:
        'Provides powerful search capabilities with Elasticsearch. Includes features like auto-completion, faceted search, and search analytics.',
      features: [
        'Full-text Search',
        'Elasticsearch Integration',
        'Search Analytics',
        'Auto-suggestions',
        'Faceted Search',
        'Search Indexing',
      ],
      icon: Search,
      name: 'Search Service',
      patterns: ['Search Pattern', 'Indexing Strategy', 'Caching Pattern'],
      port: ':8087',
      technologies: ['Go', 'Elasticsearch', 'Redis'],
    },
  ]

  const architecturalPatterns = [
    {
      description:
        'Strategic design approach focusing on core domain and domain logic',
      icon: Building,
      name: 'Domain-Driven Design (DDD)',
    },
    {
      description:
        'Loosely coupled architecture using events for communication',
      icon: Radio,
      name: 'Event-Driven Architecture',
    },
    {
      description: 'Manages distributed transactions across multiple services',
      icon: RotateCw,
      name: 'Saga Pattern',
    },
    {
      description:
        'Command Query Responsibility Segregation for scalable data access',
      icon: ArrowLeftRight,
      name: 'CQRS',
    },
    {
      description: 'Prevents cascade failures in distributed systems',
      icon: Zap,
      name: 'Circuit Breaker',
    },
    {
      description:
        'Single entry point for managing microservices communication',
      icon: ExternalLink,
      name: 'API Gateway',
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
            A comprehensive e-commerce platform built with Go microservices,
            showcasing modern distributed systems patterns, event-driven
            architecture, and advanced orchestration techniques.
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
                  <div className="flex items-start justify-between">
                    <div className="flex items-center space-x-3">
                      <service.icon className="h-8 w-8" />
                      <div>
                        <CardTitle className="text-xl">
                          {service.name}
                        </CardTitle>
                        <CardDescription className="text-sm font-mono text-blue-600 dark:text-blue-400">
                          {service.port}
                        </CardDescription>
                      </div>
                    </div>
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
                All services are containerized with Docker and orchestrated
                using Docker Compose for local development. The architecture
                supports Kubernetes deployment with proper service discovery,
                load balancing, and observability.
              </p>
              <div className="flex flex-wrap justify-center gap-2">
                <Badge variant="outline">Docker</Badge>
                <Badge variant="outline">Docker Compose</Badge>
                <Badge variant="outline">Kubernetes</Badge>
                <Badge variant="outline">Consul</Badge>
                <Badge variant="outline">Traefik</Badge>
                <Badge variant="outline">OpenTelemetry</Badge>
                <Badge variant="outline">Prometheus</Badge>
                <Badge variant="outline">Grafana</Badge>
              </div>
            </CardContent>
          </Card>
        </section>
      </div>
    </div>
  )
}
