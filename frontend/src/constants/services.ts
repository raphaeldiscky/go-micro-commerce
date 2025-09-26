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
    description: 'Central entry point with load balancing and rate limiting',
    features: ['Load Balancing', 'Rate Limiting', 'Service Discovery'],
    icon: ExternalLink,
    name: 'API Gateway',
  },
  {
    description: 'JWT authentication and authorization management',
    features: ['JWT Tokens', 'User Management', 'Role-based Access'],
    icon: Lock,
    name: 'Auth Service',
  },
  {
    description: 'Product catalog management with gRPC support',
    features: ['Product CRUD', 'gRPC API', 'Inventory Tracking'],
    icon: Package,
    name: 'Product Service',
  },
  {
    description: 'Order processing with saga orchestration patterns',
    features: ['Order Management', 'Saga Pattern', 'Workflow Engine'],
    icon: FileText,
    name: 'Order Service',
  },
  {
    description: 'Secure payment processing and transaction management',
    features: ['Payment Processing', 'Transaction History', 'Refunds'],
    icon: CreditCard,
    name: 'Payment Service',
  },
  {
    description: 'Real-time messaging with WebSocket connections',
    features: ['WebSocket', 'Real-time Messaging'],
    icon: MessageCircle,
    name: 'Chat Service',
  },
]

// Technology stack information
export const TECHNOLOGIES: Array<TechnologyInfo> = [
  { description: 'Backend Services', icon: Go, name: 'Go' },
  { description: 'Frontend UI', icon: React, name: 'React' },
  { description: 'Database', icon: PostgreSQL, name: 'PostgreSQL' },
  { description: 'Caching', icon: Redis, name: 'Redis' },
  { description: 'Event Streaming', icon: Kafka, name: 'Kafka' },
  { description: 'Containerization', icon: Docker, name: 'Docker' },
  { description: 'Orchestration', icon: Kubernetes, name: 'Kubernetes' },
  { description: 'API Communication', icon: LinkIcon, name: 'gRPC' },
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
