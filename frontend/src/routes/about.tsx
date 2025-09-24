import { createFileRoute } from '@tanstack/react-router'
import {
  BookOpen,
  Building2,
  CheckCircle,
  FileText,
  MessageCircle,
  Radio,
  Rocket,
  RotateCcw,
  Sparkles,
  Target,
} from 'lucide-react'
import { Button } from '../components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '../components/ui/card'

export const Route = createFileRoute('/about')({
  component: AboutPage,
})

function AboutPage() {
  const keyFeatures = [
    {
      title: 'Microservices Architecture',
      description:
        'Distributed system with independent services for scalability and maintainability',
      icon: Building2,
    },
    {
      title: 'Event-Driven Communication',
      description:
        'Services communicate through events using Kafka for loose coupling',
      icon: Radio,
    },
    {
      title: 'Saga Orchestration',
      description:
        'Complex workflows managed with saga patterns and compensation logic',
      icon: RotateCcw,
    },
    {
      title: 'Advanced Patterns',
      description:
        'Implementation of DDD, CQRS, Event Sourcing, and Circuit Breaker patterns',
      icon: Target,
    },
    {
      title: 'Modern Infrastructure',
      description:
        'Containerized services with Docker, service discovery, and observability',
      icon: Rocket,
    },
    {
      title: 'Real-time Features',
      description:
        'WebSocket-based chat system with multi-user support and message persistence',
      icon: MessageCircle,
    },
  ]

  const technicalGoals = [
    {
      category: 'Architecture Exploration',
      goals: [
        'Implement Domain-Driven Design (DDD) principles',
        'Demonstrate microservices communication patterns',
        'Showcase event-driven architecture with Kafka',
        'Apply CQRS and Event Sourcing patterns',
      ],
    },
    {
      category: 'Distributed Systems',
      goals: [
        'Service discovery with Consul',
        'Load balancing and API Gateway patterns',
        'Circuit breaker implementation',
        'Distributed transaction management with Saga',
      ],
    },
    {
      category: 'Advanced Go Concepts',
      goals: [
        'Clean architecture implementation',
        'gRPC services with Protocol Buffers',
        'Concurrent programming patterns',
        'Performance optimization techniques',
      ],
    },
    {
      category: 'DevOps & Observability',
      goals: [
        'Containerization with Docker',
        'Infrastructure as Code',
        'Comprehensive monitoring with Prometheus/Grafana',
        'Distributed tracing with OpenTelemetry',
      ],
    },
  ]

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-white dark:from-gray-900 dark:to-gray-800">
      <div className="container mx-auto px-4 py-12">
        {/* Header */}
        <div className="text-center mb-16">
          <h1 className="text-4xl font-bold text-gray-900 dark:text-white mb-4">
            <BookOpen className="inline-block mr-2" size={28} /> About This Project
          </h1>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-3xl mx-auto">
            Go Micro Commerce is a comprehensive exploration of distributed
            systems architecture, built to demonstrate advanced patterns and
            technologies in modern software development.
          </p>
        </div>

        {/* Project Overview */}
        <section className="mb-16">
          <Card className="max-w-4xl mx-auto">
            <CardHeader>
              <CardTitle className="text-2xl text-center">
                <Target className="inline-block mr-2" size={24} /> Project Mission
              </CardTitle>
            </CardHeader>
            <CardContent className="prose dark:prose-invert max-w-none">
              <p className="text-gray-700 dark:text-gray-300 text-center text-lg leading-relaxed">
                This application is primarily intended for{' '}
                <strong>exploring technical concepts</strong>. The goal is to
                experiment with different technologies, software architecture
                designs, and all the essential components involved in building
                distributed systems in Golang.
              </p>
              <div className="mt-8 bg-blue-50 dark:bg-blue-900/20 p-6 rounded-lg border-l-4 border-blue-500">
                <p className="text-gray-700 dark:text-gray-300 mb-0">
                  <strong>Note:</strong> While this project simulates an
                  e-commerce platform, its primary purpose is educational and
                  technical exploration rather than production deployment.
                </p>
              </div>
            </CardContent>
          </Card>
        </section>

        {/* Key Features */}
        <section className="mb-16">
          <h2 className="text-3xl font-bold text-gray-900 dark:text-white mb-8 text-center">
            <Sparkles className="inline-block mr-2" size={28} /> Key Features & Concepts
          </h2>
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {keyFeatures.map((feature) => (
              <Card
                key={feature.title}
                className="hover:shadow-lg transition-shadow"
              >
                <CardHeader>
                  <div className="flex items-center space-x-2 mb-2">
                    <feature.icon className="text-blue-500" size={24} />
                    <CardTitle className="text-lg">{feature.title}</CardTitle>
                  </div>
                  <CardDescription>{feature.description}</CardDescription>
                </CardHeader>
              </Card>
            ))}
          </div>
        </section>

        {/* Technical Goals */}
        <section className="mb-16">
          <h2 className="text-3xl font-bold text-gray-900 dark:text-white mb-8 text-center">
            <Target className="inline-block mr-2" size={28} /> Technical Learning Goals
          </h2>
          <div className="grid gap-6 md:grid-cols-2">
            {technicalGoals.map((section) => (
              <Card key={section.category} className="h-fit">
                <CardHeader>
                  <CardTitle className="text-xl">{section.category}</CardTitle>
                </CardHeader>
                <CardContent>
                  <ul className="space-y-2">
                    {section.goals.map((goal, index) => (
                      <li key={index} className="flex items-start space-x-2">
                        <CheckCircle className="text-green-500 mt-1" size={16} />
                        <span className="text-gray-700 dark:text-gray-300 text-sm">
                          {goal}
                        </span>
                      </li>
                    ))}
                  </ul>
                </CardContent>
              </Card>
            ))}
          </div>
        </section>

        {/* Architecture Highlights */}
        <section className="mb-16">
          <Card className="max-w-5xl mx-auto bg-gradient-to-r from-blue-50 to-purple-50 dark:from-blue-900/20 dark:to-purple-900/20">
            <CardHeader>
              <CardTitle className="text-2xl text-center">
                <Building2 className="inline-block mr-2" size={24} /> Architecture Highlights
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid gap-6 md:grid-cols-2">
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                    Service Architecture
                  </h3>
                  <ul className="space-y-2 text-sm text-gray-700 dark:text-gray-300">
                    <li>
                      • <strong>9 Microservices</strong> with distinct
                      responsibilities
                    </li>
                    <li>
                      • <strong>Domain-Driven Design</strong> with bounded
                      contexts
                    </li>
                    <li>
                      • <strong>Clean Architecture</strong> with layered
                      approach
                    </li>
                    <li>
                      • <strong>gRPC & REST APIs</strong> for service
                      communication
                    </li>
                  </ul>
                </div>
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                    Advanced Patterns
                  </h3>
                  <ul className="space-y-2 text-sm text-gray-700 dark:text-gray-300">
                    <li>
                      • <strong>Saga Orchestration</strong> for distributed
                      transactions
                    </li>
                    <li>
                      • <strong>Event Sourcing</strong> with Kafka integration
                    </li>
                    <li>
                      • <strong>CQRS</strong> for scalable data operations
                    </li>
                    <li>
                      • <strong>Temporal Workflows</strong> for complex business
                      logic
                    </li>
                  </ul>
                </div>
              </div>

              <div className="pt-6 border-t border-gray-200 dark:border-gray-700">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                  Infrastructure & Observability
                </h3>
                <div className="grid gap-4 md:grid-cols-3">
                  <div className="space-y-2">
                    <h4 className="font-medium text-gray-900 dark:text-white">
                      Data Layer
                    </h4>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      PostgreSQL, Redis, Elasticsearch for comprehensive data
                      management
                    </p>
                  </div>
                  <div className="space-y-2">
                    <h4 className="font-medium text-gray-900 dark:text-white">
                      Messaging
                    </h4>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      Apache Kafka with outbox pattern and dead letter queues
                    </p>
                  </div>
                  <div className="space-y-2">
                    <h4 className="font-medium text-gray-900 dark:text-white">
                      Monitoring
                    </h4>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      OpenTelemetry, Prometheus, Grafana for full observability
                    </p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </section>

        {/* Call to Action */}
        <section className="text-center">
          <Card className="max-w-2xl mx-auto bg-gradient-to-r from-green-50 to-blue-50 dark:from-green-900/20 dark:to-blue-900/20">
            <CardContent className="p-8">
              <h3 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
                <Rocket className="inline-block mr-2" size={24} /> Explore the Architecture
              </h3>
              <p className="text-gray-700 dark:text-gray-300 mb-6">
                Dive into the technical implementation and see these concepts in
                action
              </p>
              <div className="flex flex-col sm:flex-row gap-4 justify-center">
                <Button asChild size="lg">
                  <a
                    href="https://github.com/yourusername/go-micro-commerce"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center space-x-2"
                  >
                    <svg
                      className="h-5 w-5"
                      fill="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
                    </svg>
                    <span>View Source Code</span>
                  </a>
                </Button>
                <Button variant="outline" size="lg" asChild>
                  <a href="/services" className="flex items-center space-x-2">
                    <FileText size={16} />
                    <span>Browse Services</span>
                  </a>
                </Button>
              </div>
            </CardContent>
          </Card>
        </section>
      </div>
    </div>
  )
}
