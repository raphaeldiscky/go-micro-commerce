// TypeScript interfaces for all constants

export interface KeyFeature {
  description: string
  icon: React.ElementType
  title: string
}

export interface NavigationItem {
  description?: string
  icon: React.ElementType
  name: string
  path: string
}

export interface ServiceInfo {
  description: string
  features: Array<string>
  icon: React.ElementType
  name: string
}

export interface SocialLink {
  ariaLabel: string
  icon: React.ElementType
  name: string
  url: string
}

export interface TechnicalGoalSection {
  category: string
  goals: Array<string>
}

export interface TechnologyInfo {
  description: string
  icon: React.ElementType
  name: string
}

export interface TechnologyLink {
  name: string
  url: string
}
