import {
  Building2,
  MessageCircle,
  Radio,
  Rocket,
  RotateCcw,
  Target,
} from 'lucide-react'
import type { KeyFeature, TechnicalGoalSection } from './types'

// About page content
export const KEY_FEATURES: Array<KeyFeature> = [
  {
    description:
      'Distributed system with independent services for scalability and maintainability',
    icon: Building2,
    title: 'Microservices Architecture',
  },
  {
    description:
      'Services communicate through events using Kafka for loose coupling',
    icon: Radio,
    title: 'Event-Driven Communication',
  },
  {
    description:
      'Complex workflows managed with saga patterns and compensation logic',
    icon: RotateCcw,
    title: 'Saga Orchestration',
  },
  {
    description:
      'Implementation of DDD, CQRS, Event Sourcing, and Circuit Breaker patterns',
    icon: Target,
    title: 'Advanced Patterns',
  },
  {
    description:
      'Containerized services with Docker, service discovery, and observability',
    icon: Rocket,
    title: 'Modern Infrastructure',
  },
  {
    description:
      'WebSocket-based chat system with multi-user support and message persistence',
    icon: MessageCircle,
    title: 'Real-time Features',
  },
]

export const TECHNICAL_GOALS: Array<TechnicalGoalSection> = [
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

// Page titles and descriptions
export const PAGE_CONTENT = {
  ABOUT: {
    ARCHITECTURE_HIGHLIGHTS: {
      ADVANCED_ITEMS: [
        'Saga Orchestration for distributed transactions',
        'Event Sourcing with Kafka integration',
        'CQRS for scalable data operations',
        'Temporal Workflows for complex business logic',
      ],
      ADVANCED_PATTERNS: 'Advanced Patterns',
      DATA_LAYER: {
        DESCRIPTION:
          'PostgreSQL, Redis, Elasticsearch for comprehensive data management',
        TITLE: 'Data Layer',
      },
      INFRASTRUCTURE_TITLE: 'Infrastructure & Observability',
      MESSAGING: {
        DESCRIPTION: 'Apache Kafka with outbox pattern and dead letter queues',
        TITLE: 'Messaging',
      },
      MONITORING: {
        DESCRIPTION:
          'OpenTelemetry, Prometheus, Grafana for full observability',
        TITLE: 'Monitoring',
      },
      SERVICE_ARCHITECTURE: 'Service Architecture',
      SERVICE_ITEMS: [
        '9 Microservices with distinct responsibilities',
        'Domain-Driven Design with bounded contexts',
        'Clean Architecture with layered approach',
        'gRPC & REST APIs for service communication',
      ],
      TITLE: 'Architecture Highlights',
    },
    CTA: {
      BROWSE_SERVICES: 'Browse Services',
      DESCRIPTION:
        'Dive into the technical implementation and see these concepts in action',
      TITLE: 'Explore the Architecture',
      VIEW_SOURCE: 'View Source Code',
    },
    HERO_DESCRIPTION:
      'Go Micro Commerce is a comprehensive exploration of distributed systems architecture, built to demonstrate advanced patterns and technologies in modern software development.',
    KEY_FEATURES_TITLE: 'Key Features & Concepts',
    PROJECT_MISSION: {
      DESCRIPTION:
        'This application is primarily intended for exploring technical concepts. The goal is to experiment with different technologies, software architecture designs, and all the essential components involved in building distributed systems in Golang.',
      NOTE: 'While this project simulates an e-commerce platform, its primary purpose is educational and technical exploration rather than production deployment.',
      TITLE: 'Project Mission',
    },
    TECHNICAL_GOALS_TITLE: 'Technical Learning Goals',
    TITLE: 'About This Project',
  },
  HOME: {
    GET_STARTED: {
      ARCHITECTURE: {
        CTA: 'View Architecture',
        DESCRIPTION:
          'Explore the distributed systems architecture and design patterns',
        TITLE: 'Architecture',
      },
      LEARN_MORE: {
        CTA: 'About Project',
        DESCRIPTION:
          'Discover the technical concepts and implementation details',
        TITLE: 'Learn More',
      },
      PRODUCT: {
        CTA: 'Browse Products',
        DESCRIPTION: 'View available products and place your orders',
        TITLE: 'Products',
      },
      TITLE: 'Get Started',
    },
    SERVICES_SECTION: {
      DESCRIPTION:
        'Built with Domain-Driven Design principles, featuring distributed systems patterns and event-driven architecture for scalable e-commerce operations.',
      TITLE: 'Microservices Architecture',
    },
    TECH_STACK: {
      DESCRIPTION:
        'Built with modern, production-ready technologies powering our distributed architecture',
      TITLE: 'Technology Stack',
    },
  },
} as const
