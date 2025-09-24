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
    TITLE: 'About This Project',
    HERO_DESCRIPTION:
      'Go Micro Commerce is a comprehensive exploration of distributed systems architecture, built to demonstrate advanced patterns and technologies in modern software development.',
    PROJECT_MISSION: {
      TITLE: 'Project Mission',
      DESCRIPTION:
        'This application is primarily intended for exploring technical concepts. The goal is to experiment with different technologies, software architecture designs, and all the essential components involved in building distributed systems in Golang.',
      NOTE: 'While this project simulates an e-commerce platform, its primary purpose is educational and technical exploration rather than production deployment.',
    },
    KEY_FEATURES_TITLE: 'Key Features & Concepts',
    TECHNICAL_GOALS_TITLE: 'Technical Learning Goals',
    ARCHITECTURE_HIGHLIGHTS: {
      TITLE: 'Architecture Highlights',
      SERVICE_ARCHITECTURE: 'Service Architecture',
      SERVICE_ITEMS: [
        '9 Microservices with distinct responsibilities',
        'Domain-Driven Design with bounded contexts',
        'Clean Architecture with layered approach',
        'gRPC & REST APIs for service communication',
      ],
      ADVANCED_PATTERNS: 'Advanced Patterns',
      ADVANCED_ITEMS: [
        'Saga Orchestration for distributed transactions',
        'Event Sourcing with Kafka integration',
        'CQRS for scalable data operations',
        'Temporal Workflows for complex business logic',
      ],
      INFRASTRUCTURE_TITLE: 'Infrastructure & Observability',
      DATA_LAYER: {
        TITLE: 'Data Layer',
        DESCRIPTION:
          'PostgreSQL, Redis, Elasticsearch for comprehensive data management',
      },
      MESSAGING: {
        TITLE: 'Messaging',
        DESCRIPTION: 'Apache Kafka with outbox pattern and dead letter queues',
      },
      MONITORING: {
        TITLE: 'Monitoring',
        DESCRIPTION:
          'OpenTelemetry, Prometheus, Grafana for full observability',
      },
    },
    CTA: {
      TITLE: 'Explore the Architecture',
      DESCRIPTION:
        'Dive into the technical implementation and see these concepts in action',
      VIEW_SOURCE: 'View Source Code',
      BROWSE_SERVICES: 'Browse Services',
    },
  },
  HOME: {
    SERVICES_SECTION: {
      TITLE: 'Microservices Architecture',
      DESCRIPTION:
        'Built with Domain-Driven Design principles, featuring distributed systems patterns and event-driven architecture for scalable e-commerce operations.',
    },
    TECH_STACK: {
      TITLE: 'Technology Stack',
      DESCRIPTION:
        'Built with modern, production-ready technologies powering our distributed architecture',
    },
    GET_STARTED: {
      TITLE: 'Get Started',
      ARCHITECTURE: {
        TITLE: 'Architecture',
        DESCRIPTION:
          'Explore the distributed systems architecture and design patterns',
        CTA: 'View Architecture',
      },
      LIVE_DEMO: {
        TITLE: 'Live Demo',
        DESCRIPTION: 'Try the real-time chat with WebSocket connections',
        CTA: 'Try Chat Demo',
      },
      LEARN_MORE: {
        TITLE: 'Learn More',
        DESCRIPTION:
          'Discover the technical concepts and implementation details',
        CTA: 'About Project',
      },
    },
  },
} as const
