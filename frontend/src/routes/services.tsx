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
      name: 'API Gateway',
      description:
        'Central entry point for all client requests with comprehensive traffic management',
      icon: ExternalLink,
      port: ':8080',
      features: [
        'Load Balancing',
        'Rate Limiting',
        'Service Discovery',
        'Request Routing',
        'Circuit Breaker',
        'Monitoring',
      ],
      technologies: ['Traefik', 'Consul', 'Go'],
      patterns: ['Gateway Pattern', 'Circuit Breaker', 'Service Discovery'],
      details:
        'Handles all incoming requests and routes them to appropriate microservices. Provides cross-cutting concerns like authentication, logging, and monitoring.',
    },
    {
      name: 'Auth Service',
      description:
        'JWT-based authentication and authorization with role management',
      icon: Lock,
      port: ':8081',
      features: [
        'JWT Authentication',
        'User Registration',
        'Role-based Access Control',
        'Token Refresh',
        'Session Management',
        'Password Security',
      ],
      technologies: ['Go', 'PostgreSQL', 'JWT', 'bcrypt'],
      patterns: ['JWT Pattern', 'Repository Pattern', 'Clean Architecture'],
      details:
        'Manages user authentication, authorization, and access control across all services. Implements secure password handling and token-based authentication.',
    },
    {
      name: 'Product Service',
      description:
        'Product catalog management with both REST and gRPC interfaces',
      icon: Package,
      port: ':8082',
      features: [
        'Product CRUD',
        'gRPC API',
        'Inventory Tracking',
        'Category Management',
        'Search Integration',
        'Price Management',
      ],
      technologies: ['Go', 'PostgreSQL', 'gRPC', 'Protocol Buffers'],
      patterns: ['Domain-Driven Design', 'gRPC Pattern', 'Repository Pattern'],
      details:
        'Handles product catalog operations with dual API support. Integrates with search service for enhanced product discovery.',
    },
    {
      name: 'Order Service',
      description:
        'Complex order processing with saga orchestration and workflow management',
      icon: FileText,
      port: ':8083',
      features: [
        'Order Management',
        'Saga Orchestration',
        'Temporal Workflows',
        'Compensation Logic',
        'State Management',
        'Event Sourcing',
      ],
      technologies: ['Go', 'PostgreSQL', 'Temporal', 'Kafka'],
      patterns: [
        'Saga Pattern',
        'Event Sourcing',
        'Temporal Workflows',
        'CQRS',
      ],
      details:
        'Orchestrates complex order workflows using saga patterns. Manages distributed transactions with compensation logic for failure scenarios.',
    },
    {
      name: 'Payment Service',
      description:
        'Secure payment processing with transaction management and fraud prevention',
      icon: CreditCard,
      port: ':8084',
      features: [
        'Payment Processing',
        'Transaction History',
        'Refund Management',
        'Fraud Detection',
        'Payment Methods',
        'Webhooks',
      ],
      technologies: ['Go', 'PostgreSQL', 'Stripe API', 'Redis'],
      patterns: ['Payment Gateway Pattern', 'Idempotency', 'Retry Pattern'],
      details:
        'Handles secure payment processing with multiple payment methods. Implements fraud detection and comprehensive transaction management.',
    },
    {
      name: 'Fulfillment Service',
      description:
        'Inventory management and order fulfillment with shipping integration',
      icon: Package,
      port: ':8085',
      features: [
        'Inventory Management',
        'Order Fulfillment',
        'Shipping Integration',
        'Stock Tracking',
        'Warehouse Management',
        'Delivery Updates',
      ],
      technologies: ['Go', 'PostgreSQL', 'Redis', 'Kafka'],
      patterns: ['Event-Driven Architecture', 'Saga Pattern', 'Domain Events'],
      details:
        'Manages inventory levels and fulfillment processes. Coordinates with shipping providers and maintains real-time stock information.',
    },
    {
      name: 'Chat Service',
      description:
        'Real-time messaging with WebSocket connections and message persistence',
      icon: MessageCircle,
      port: ':9088',
      features: [
        'WebSocket Support',
        'Real-time Messaging',
        'Multi-user Rooms',
        'Message History',
        'Connection Management',
        'Scalable Architecture',
      ],
      technologies: ['Go', 'WebSocket', 'Redis', 'PostgreSQL'],
      patterns: ['WebSocket Pattern', 'Pub/Sub', 'Connection Pooling'],
      details:
        'Provides real-time communication capabilities with persistent message storage. Supports multiple concurrent connections and room-based messaging.',
    },
    {
      name: 'Notification Service',
      description: 'Event-driven notifications with multiple delivery channels',
      icon: Mail,
      port: ':8086',
      features: [
        'Email Notifications',
        'Push Notifications',
        'SMS Support',
        'Template Management',
        'Delivery Tracking',
        'Event-driven Triggers',
      ],
      technologies: ['Go', 'Kafka', 'SMTP', 'Redis'],
      patterns: [
        'Event-Driven Architecture',
        'Template Pattern',
        'Observer Pattern',
      ],
      details:
        'Handles all system notifications across multiple channels. Processes events from other services to trigger appropriate notifications.',
    },
    {
      name: 'Search Service',
      description:
        'Advanced search capabilities with Elasticsearch integration',
      icon: Search,
      port: ':8087',
      features: [
        'Full-text Search',
        'Elasticsearch Integration',
        'Search Analytics',
        'Auto-suggestions',
        'Faceted Search',
        'Search Indexing',
      ],
      technologies: ['Go', 'Elasticsearch', 'Redis'],
      patterns: ['Search Pattern', 'Indexing Strategy', 'Caching Pattern'],
      details:
        'Provides powerful search capabilities with Elasticsearch. Includes features like auto-completion, faceted search, and search analytics.',
    },
  ]

  const architecturalPatterns = [
    {
      name: 'Domain-Driven Design (DDD)',
      description:
        'Strategic design approach focusing on core domain and domain logic',
      icon: Building,
    },
    {
      name: 'Event-Driven Architecture',
      description:
        'Loosely coupled architecture using events for communication',
      icon: Radio,
    },
    {
      name: 'Saga Pattern',
      description: 'Manages distributed transactions across multiple services',
      icon: RotateCw,
    },
    {
      name: 'CQRS',
      description:
        'Command Query Responsibility Segregation for scalable data access',
      icon: ArrowLeftRight,
    },
    {
      name: 'Circuit Breaker',
      description: 'Prevents cascade failures in distributed systems',
      icon: Zap,
    },
    {
      name: 'API Gateway',
      description:
        'Single entry point for managing microservices communication',
      icon: ExternalLink,
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
                key={pattern.name}
                className="hover:shadow-lg transition-shadow"
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
                key={service.name}
                className="hover:shadow-xl transition-all duration-300 border-l-4 border-blue-500"
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
                          key={feature}
                          variant="secondary"
                          className="text-xs"
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
                        <Badge key={tech} variant="outline" className="text-xs">
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
                          key={pattern}
                          className="text-xs bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200"
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
