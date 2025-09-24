import { Home, Info, MessageCircle, Settings } from 'lucide-react'
import type { NavigationItem } from './types'

// Navigation menu items
export const NAVIGATION_ITEMS: Array<NavigationItem> = [
  { name: 'Home', path: '/', icon: Home },
  { name: 'Services', path: '/services', icon: Settings },
  { name: 'Chat Demo', path: '/chat', icon: MessageCircle },
  { name: 'About', path: '/about', icon: Info },
]

// Quick links for footer (subset of navigation)
export const QUICK_LINKS: Array<NavigationItem> = [
  { name: 'Home', path: '/', icon: Home },
  { name: 'Services', path: '/services', icon: Settings },
  { name: 'Chat Demo', path: '/chat', icon: MessageCircle },
  { name: 'About', path: '/about', icon: Info },
]
