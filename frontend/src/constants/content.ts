import {
  Activity,
  Building2,
  MessageCircle,
  Radio,
  Rocket,
  RotateCcw,
} from 'lucide-react'
import type { KeyFeature, TechnicalGoalSection } from './types'

// About page content
export const KEY_FEATURES: Array<KeyFeature> = [
  {
    description:
      'Independent services with DDD, clean architecture, and bounded contexts',
    icon: Building2,
    title: 'Microservices Architecture',
  },
  {
    description:
      'Kafka KRaft cluster, Redis Pub/Sub, and Asynq for distributed messaging',
    icon: Radio,
    title: 'Event-Driven Communication',
  },
  {
    description:
      'Dual saga implementation with custom PostgreSQL and Temporal workflows',
    icon: RotateCcw,
    title: 'Saga Orchestration',
  },
  {
    description:
      'WebSocket chat, SSE notifications, and Redis-powered real-time updates',
    icon: MessageCircle,
    title: 'Real-time Features',
  },
  {
    description:
      'Kubernetes with GitOps, Terraform IaC, and Cloudflare Pages deployment',
    icon: Rocket,
    title: 'Production Infrastructure',
  },
  {
    description:
      'LGTM stack with Alloy collector for metrics, traces, and logs',
    icon: Activity,
    title: 'Full Observability',
  },
]

export const TECHNICAL_GOALS: Array<TechnicalGoalSection> = [
  {
    category: 'Architecture Exploration',
    goals: [
      'Domain-Driven Design with bounded contexts',
      'Event-driven architecture with Kafka KRaft',
      'Outbox and inbox patterns for reliable messaging',
      'Apollo Federation for GraphQL composition',
    ],
  },
  {
    category: 'Distributed Systems',
    goals: [
      'Saga orchestration',
      'Distributed task scheduling',
      'Circuit breaker and retry patterns',
      'Service discovery and load balancing',
    ],
  },
  {
    category: 'DevOps & Observability',
    goals: [
      'GitOps with ArgoCD and Kustomize',
      'Terraform IaC for GKE provisioning',
      'LGTM stack with Alloy telemetry collector',
      'Cloudflare Pages for frontend deployment',
    ],
  },
]

// Page titles and descriptions
export const PAGE_CONTENT = {
  ABOUT: {
    ARCHITECTURE_HIGHLIGHTS: {
      ADVANCED_ITEMS: [
        'Dual Saga implementation options (PostgreSQL + Temporal)',
        'Outbox and inbox patterns for reliable events',
        'Asynq distributed task scheduling',
        'Apollo Federation for GraphQL composition',
      ],
      ADVANCED_PATTERNS: 'Advanced Patterns',
      DATA_LAYER: {
        DESCRIPTION:
          'PostgreSQL per service, 6-node Redis Cluster, Elasticsearch',
        TITLE: 'Data Layer',
      },
      INFRASTRUCTURE_TITLE: 'Infrastructure & Observability',
      MESSAGING: {
        DESCRIPTION:
          '3-node Kafka KRaft cluster with DLQ and transactional outbox',
        TITLE: 'Messaging',
      },
      MONITORING: {
        DESCRIPTION: 'LGTM stack (Loki, Grafana, Tempo, Prometheus) with Alloy',
        TITLE: 'Monitoring',
      },
      SERVICE_ARCHITECTURE: 'Service Architecture',
      SERVICE_ITEMS: [
        'Microservices with distinct responsibilities',
        'Domain-Driven Design with bounded contexts',
        'Clean Architecture with layered approach',
        'gRPC, REST, GraphQL, WebSocket, and SSE APIs',
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
      'Go Micro Commerce is a comprehensive exploration of distributed systems architecture, built to demonstrate advanced patterns and production-grade infrastructure.',
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
          'Explore microservices with DDD, saga orchestration, and event-driven patterns',
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
        'Microservices with DDD, saga orchestration, event-driven patterns, and production-grade infrastructure.',
      TITLE: 'Microservices Architecture',
    },
    TECH_STACK: {
      DESCRIPTION:
        'Production-ready cloud-native  stack with Kubernetes, GitOps, and full observability',
      TITLE: 'Technology Stack',
    },
  },
} as const
