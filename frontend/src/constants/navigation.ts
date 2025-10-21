import { Home, Info, Package, Receipt, Settings } from 'lucide-react'
import { PATH, PATH_ROOT } from './routes'
import type { NavigationItem } from './types'

// Navigation menu items
export const NAVIGATION_ITEMS: Array<NavigationItem> = [
  { icon: Home, name: 'Home', path: PATH_ROOT.home },
  { icon: Package, name: 'Products', path: PATH.products.root },
  { icon: Receipt, name: 'Orders', path: PATH.orders.root },
  { icon: Settings, name: 'Services', path: PATH_ROOT.services },
  { icon: Info, name: 'About', path: PATH_ROOT.about },
]

// Explore dropdown items
export const EXPLORE_ITEMS: Array<NavigationItem> = [
  {
    description: 'Browse products',
    icon: Package,
    name: 'Products',
    path: PATH.products.root,
  },
  {
    description: 'View order transactions and history',
    icon: Receipt,
    name: 'Orders',
    path: PATH.orders.root,
  },
]

// Quick links for footer (subset of navigation)
export const QUICK_LINKS: Array<NavigationItem> = [
  { icon: Home, name: 'Home', path: PATH_ROOT.home },
  { icon: Package, name: 'Products', path: PATH.products.root },
  { icon: Receipt, name: 'Orders', path: PATH.orders.root },
  { icon: Info, name: 'About', path: PATH_ROOT.about },
]
