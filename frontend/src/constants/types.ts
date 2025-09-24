// TypeScript interfaces for all constants

export interface NavigationItem {
  name: string
  path: string
  icon: React.ElementType
}

export interface ServiceInfo {
  name: string
  description: string
  icon: React.ElementType
  features: string[]
}

export interface TechnologyInfo {
  name: string
  description: string
  icon: React.ElementType
}

export interface TechnologyLink {
  name: string
  url: string
}

export interface SocialLink {
  name: string
  url: string
  icon: React.ElementType
  ariaLabel: string
}

export interface KeyFeature {
  title: string
  description: string
  icon: React.ElementType
}

export interface TechnicalGoalSection {
  category: string
  goals: string[]
}
