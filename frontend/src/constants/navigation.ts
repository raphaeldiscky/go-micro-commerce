import { Home, Info, MessageCircle, Package, Settings } from 'lucide-react'
import type { NavigationItem } from './types'

// Navigation menu items
export const NAVIGATION_ITEMS: Array<NavigationItem> = [
  { icon: Home, name: 'Home', path: '/' },
  { icon: Package, name: 'Products', path: '/products' },
  { icon: MessageCircle, name: 'Chat Demo', path: '/chat' },
  { icon: Settings, name: 'Services', path: '/services' },
  { icon: Info, name: 'About', path: '/about' },
]

// Features dropdown items
export const FEATURES_ITEMS: Array<NavigationItem> = [
  {
    description: 'Browse products with cursor pagination',
    icon: Package,
    name: 'Products',
    path: '/products',
  },
  {
    description: 'Real-time messaging with WebSocket',
    icon: MessageCircle,
    name: 'Chat Demo',
    path: '/chat',
  },
]

// Quick links for footer (subset of navigation)
export const QUICK_LINKS: Array<NavigationItem> = [
  { icon: Home, name: 'Home', path: '/' },
  { icon: Package, name: 'Products', path: '/products' },
  { icon: Settings, name: 'Services', path: '/services' },
  { icon: MessageCircle, name: 'Chat Demo', path: '/chat' },
  { icon: Info, name: 'About', path: '/about' },
]
