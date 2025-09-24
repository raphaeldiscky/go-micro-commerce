import { Link, createFileRoute } from '@tanstack/react-router'
import {
  Docker,
  Go,
  Kubernetes,
  PostgreSQL,
  React,
  Redis,
  Kafka,
} from 'developer-icons'
import {
  BookOpen,
  Building,
  CreditCard,
  ExternalLink,
  FileText,
  Link as LinkIcon,
  Lock,
  MessageCircle,
  Package,
  Wrench,
} from 'lucide-react'
import { Button } from '../components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '../components/ui/card'

export const Route = createFileRoute('/')({
  component: HomePage,
})

function HomePage() {
  const services = [
    {
      name: 'API Gateway',
      description: 'Central entry point with load balancing and rate limiting',
      icon: ExternalLink,
      features: ['Load Balancing', 'Rate Limiting', 'Service Discovery'],
    },
    {
      name: 'Auth Service',
      description: 'JWT authentication and authorization management',
      icon: Lock,
      features: ['JWT Tokens', 'User Management', 'Role-based Access'],
    },
    {
      name: 'Product Service',
      description: 'Product catalog management with gRPC support',
      icon: Package,
      features: ['Product CRUD', 'gRPC API', 'Inventory Tracking'],
    },
    {
      name: 'Order Service',
      description: 'Order processing with saga orchestration patterns',
      icon: FileText,
      features: ['Order Management', 'Saga Pattern', 'Workflow Engine'],
    },
    {
      name: 'Payment Service',
      description: 'Secure payment processing and transaction management',
      icon: CreditCard,
      features: ['Payment Processing', 'Transaction History', 'Refunds'],
    },
    {
      name: 'Chat Service',
      description: 'Real-time messaging with WebSocket connections',
      icon: MessageCircle,
      features: ['WebSocket', 'Real-time Messaging', 'Multi-user Support'],
    },
  ]

  const technologies = [
    { name: 'Go', icon: Go, description: 'Backend Services' },
    { name: 'React', icon: React, description: 'Frontend UI' },
    { name: 'PostgreSQL', icon: PostgreSQL, description: 'Database' },
    { name: 'Redis', icon: Redis, description: 'Caching' },
    { name: 'Kafka', icon: Kafka, description: 'Event Streaming' },
    { name: 'Docker', icon: Docker, description: 'Containerization' },
    { name: 'Kubernetes', icon: Kubernetes, description: 'Orchestration' },
    { name: 'gRPC', icon: LinkIcon, description: 'API Communication' },
  ]

  return (
    <div className="min-h-screen">
      {/* Hero Section */}
      <section className="relative overflow-hidden bg-gradient-to-br from-blue-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900">
        <div className="container mx-auto px-4 py-24 sm:py-32">
          <div className="text-center">
            <h1 className="text-4xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-6xl">
              Go Micro Commerce
            </h1>
            <p className="mt-6 text-lg leading-8 text-gray-600 dark:text-gray-300 max-w-2xl mx-auto">
              A modern distributed systems architecture built with Go
              microservices, demonstrating advanced patterns like saga
              orchestration, event-driven communication, and scalable design.
            </p>
            <div className="mt-10 flex items-center justify-center gap-x-6">
              <Button asChild size="lg">
                <Link to="/services" className="flex items-center gap-2">
                  <Wrench className="h-5 w-5" />
                  Explore Services
                </Link>
              </Button>
              <Button variant="outline" size="lg" asChild>
                <Link to="/chat" className="flex items-center gap-2">
                  <MessageCircle className="h-5 w-5" />
                  Try Chat Demo
                </Link>
              </Button>
            </div>
          </div>
        </div>
        {/* Decorative background elements */}
        <div className="absolute inset-0 -z-10 h-full w-full bg-white dark:bg-gray-900 [background:radial-gradient(125%_125%_at_50%_10%,#fff_40%,#63e_100%)] dark:[background:radial-gradient(125%_125%_at_50%_10%,#000_40%,#63e_100%)] opacity-20"></div>
      </section>

      {/* Services Grid */}
      <section className="py-16 bg-white dark:bg-gray-800">
        <div className="container mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-bold text-gray-900 dark:text-white mb-4">
              Microservices Architecture
            </h2>
            <p className="text-lg text-gray-600 dark:text-gray-300 max-w-3xl mx-auto">
              Built with Domain-Driven Design principles, featuring distributed
              systems patterns and event-driven architecture for scalable
              e-commerce operations.
            </p>
          </div>

          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {services.map((service) => (
              <Card
                key={service.name}
                className="hover:shadow-lg transition-shadow"
              >
                <CardHeader>
                  <div className="flex items-center space-x-2">
                    <service.icon className="h-6 w-6 text-blue-600 dark:text-blue-400" />
                    <CardTitle className="text-xl">{service.name}</CardTitle>
                  </div>
                  <CardDescription>{service.description}</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex flex-wrap gap-2">
                    {service.features.map((feature) => (
                      <span
                        key={feature}
                        className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
                      >
                        {feature}
                      </span>
                    ))}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Technology Stack */}
      <section className="py-16 bg-gray-50 dark:bg-gray-900">
        <div className="container mx-auto px-4">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold text-gray-900 dark:text-white mb-4">
              Technology Stack
            </h2>
            <p className="text-lg text-gray-600 dark:text-gray-300">
              Built with modern, production-ready technologies
            </p>
          </div>

          <div className="grid gap-4 grid-cols-2 md:grid-cols-4 lg:grid-cols-4">
            {technologies.map((tech) => (
              <Card
                key={tech.name}
                className="flex flex-col items-center p-4 hover:shadow-md transition-shadow group cursor-default"
              >
                <div className="w-12 h-12 flex items-center justify-center mb-3 rounded-lg bg-muted/50 group-hover:bg-muted transition-colors">
                  <tech.icon className="h-7 w-7 text-foreground" />
                </div>
                <h3 className="text-xs font-semibold text-foreground mb-1">
                  {tech.name}
                </h3>
                <p className="text-xs text-muted-foreground text-center">
                  {tech.description}
                </p>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Quick Links */}
      <section className="py-16 bg-white dark:bg-gray-800">
        <div className="container mx-auto px-4">
          <div className="text-center">
            <h2 className="text-3xl font-bold text-gray-900 dark:text-white mb-8">
              Get Started
            </h2>
            <div className="grid gap-6 md:grid-cols-3 max-w-4xl mx-auto">
              <Card className="text-center hover:shadow-lg transition-shadow">
                <CardHeader>
                  <div className="mb-2">
                    <Building className="h-10 w-10 mx-auto" />
                  </div>
                  <CardTitle>Architecture</CardTitle>
                  <CardDescription>
                    Explore the distributed systems architecture and design
                    patterns
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Button variant="outline" asChild className="w-full">
                    <Link to="/services">View Architecture</Link>
                  </Button>
                </CardContent>
              </Card>

              <Card className="text-center hover:shadow-lg transition-shadow">
                <CardHeader>
                  <div className="mb-2">
                    <MessageCircle className="h-10 w-10 mx-auto" />
                  </div>
                  <CardTitle>Live Demo</CardTitle>
                  <CardDescription>
                    Try the real-time chat with WebSocket connections
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Button asChild className="w-full">
                    <Link to="/chat">Try Chat Demo</Link>
                  </Button>
                </CardContent>
              </Card>

              <Card className="text-center hover:shadow-lg transition-shadow">
                <CardHeader>
                  <div className="mb-2">
                    <BookOpen className="h-10 w-10 mx-auto" />
                  </div>
                  <CardTitle>Learn More</CardTitle>
                  <CardDescription>
                    Discover the technical concepts and implementation details
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Button variant="outline" asChild className="w-full">
                    <Link to="/about">About Project</Link>
                  </Button>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}
