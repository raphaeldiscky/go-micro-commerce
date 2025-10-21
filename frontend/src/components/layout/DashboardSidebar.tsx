import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'
import { PATH_DASHBOARD, PATH_ROOT } from '@/constants/routes'
import { Link, useLocation } from '@tanstack/react-router'
import {
  ArrowLeft,
  BarChart3,
  DollarSign,
  HomeIcon,
  Package,
  ShoppingCart,
  Users,
} from 'lucide-react'

interface NavItem {
  title: string
  href: string
  icon: React.ComponentType<{ className?: string }>
}

const analyticsItems: Array<NavItem> = [
  {
    href: PATH_DASHBOARD.root,
    icon: HomeIcon,
    title: 'Overview',
  },
  {
    href: PATH_DASHBOARD.revenue,
    icon: DollarSign,
    title: 'Revenue',
  },
]

const managementItems: Array<NavItem> = [
  {
    href: PATH_DASHBOARD.orders,
    icon: ShoppingCart,
    title: 'Orders',
  },
  {
    href: PATH_DASHBOARD.products,
    icon: Package,
    title: 'Products',
  },
  {
    href: PATH_DASHBOARD.users,
    icon: Users,
    title: 'Users',
  },
]

export function DashboardSidebar() {
  const location = useLocation()

  return (
    <Sidebar>
      <SidebarHeader>
        <div className="flex items-center gap-2 px-2 py-2">
          <BarChart3 className="h-6 w-6 text-primary" />
          <span className="text-lg font-semibold">Admin Dashboard</span>
        </div>
        <SidebarMenuButton asChild>
          <Link to={PATH_ROOT.home}>
            <ArrowLeft />
            <span>Back to Home</span>
          </Link>
        </SidebarMenuButton>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Analytics</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {analyticsItems.map((item) => {
                const isActive = location.pathname === item.href
                const Icon = item.icon

                return (
                  <SidebarMenuItem key={item.href}>
                    <SidebarMenuButton asChild isActive={isActive}>
                      <Link to={item.href}>
                        <Icon />
                        <span>{item.title}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                )
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarGroup>
          <SidebarGroupLabel>Management</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {managementItems.map((item) => {
                const isActive = location.pathname === item.href
                const Icon = item.icon

                return (
                  <SidebarMenuItem key={item.href}>
                    <SidebarMenuButton asChild isActive={isActive}>
                      <Link to={item.href}>
                        <Icon />
                        <span>{item.title}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                )
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <p className="px-2 text-xs text-muted-foreground">Admin Panel v1.0.0</p>
      </SidebarFooter>
    </Sidebar>
  )
}
