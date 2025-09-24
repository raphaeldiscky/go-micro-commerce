import {
  Docker,
  Go,
  Kafka,
  Kubernetes,
  PostgreSQL,
  React,
  Redis,
} from 'developer-icons'
import {
  CreditCard,
  ExternalLink,
  FileText,
  Link as LinkIcon,
  Lock,
  MessageCircle,
  Package,
} from 'lucide-react'
import type { ServiceInfo, TechnologyInfo, TechnologyLink } from './types'
import { TECH_URLS } from './urls'

// Microservices information
export const SERVICES: Array<ServiceInfo> = [
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

// Technology stack information
export const TECHNOLOGIES: Array<TechnologyInfo> = [
  { name: 'Go', icon: Go, description: 'Backend Services' },
  { name: 'React', icon: React, description: 'Frontend UI' },
  { name: 'PostgreSQL', icon: PostgreSQL, description: 'Database' },
  { name: 'Redis', icon: Redis, description: 'Caching' },
  { name: 'Kafka', icon: Kafka, description: 'Event Streaming' },
  { name: 'Docker', icon: Docker, description: 'Containerization' },
  { name: 'Kubernetes', icon: Kubernetes, description: 'Orchestration' },
  { name: 'gRPC', icon: LinkIcon, description: 'API Communication' },
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
