import {
  Home,
  Info,
  MessageCircle,
  Package,
  Receipt,
  Settings,
} from 'lucide-react'
import { PATH_FEATURES, PATH_ROOT } from './routes'
import type { NavigationItem } from './types'

// Navigation menu items
export const NAVIGATION_ITEMS: Array<NavigationItem> = [
  { icon: Home, name: 'Home', path: PATH_ROOT.home },
  { icon: Package, name: 'Products', path: PATH_FEATURES.products.root },
  { icon: Receipt, name: 'Orders', path: PATH_FEATURES.orders.root },
  { icon: MessageCircle, name: 'Chat Demo', path: PATH_FEATURES.chat.root },
  { icon: Settings, name: 'Services', path: PATH_ROOT.services },
  { icon: Info, name: 'About', path: PATH_ROOT.about },
]

// Features dropdown items
export const FEATURES_ITEMS: Array<NavigationItem> = [
  {
    description: 'Browse products with cursor pagination',
    icon: Package,
    name: 'Products',
    path: PATH_FEATURES.products.root,
  },
  {
    description: 'View order transactions and history',
    icon: Receipt,
    name: 'Orders',
    path: PATH_FEATURES.orders.root,
  },
  {
    description: 'Real-time messaging with WebSocket',
    icon: MessageCircle,
    name: 'Chat Demo',
    path: PATH_FEATURES.chat.root,
  },
]

// Quick links for footer (subset of navigation)
export const QUICK_LINKS: Array<NavigationItem> = [
  { icon: Home, name: 'Home', path: PATH_ROOT.home },
  { icon: Package, name: 'Products', path: PATH_FEATURES.products.root },
  { icon: Receipt, name: 'Orders', path: PATH_FEATURES.orders.root },
  { icon: Settings, name: 'Services', path: PATH_ROOT.services },
  { icon: MessageCircle, name: 'Chat Demo', path: PATH_FEATURES.chat.root },
  { icon: Info, name: 'About', path: PATH_ROOT.about },
]
