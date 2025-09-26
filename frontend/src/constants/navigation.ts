import { Home, Info, MessageCircle, Settings } from 'lucide-react'
import type { NavigationItem } from './types'

// Navigation menu items
export const NAVIGATION_ITEMS: Array<NavigationItem> = [
  { icon: Home, name: 'Home', path: '/' },
  { icon: Settings, name: 'Services', path: '/services' },
  { icon: MessageCircle, name: 'Chat Demo', path: '/chat' },
  { icon: Info, name: 'About', path: '/about' },
]

// Quick links for footer (subset of navigation)
export const QUICK_LINKS: Array<NavigationItem> = [
  { icon: Home, name: 'Home', path: '/' },
  { icon: Settings, name: 'Services', path: '/services' },
  { icon: MessageCircle, name: 'Chat Demo', path: '/chat' },
  { icon: Info, name: 'About', path: '/about' },
]
