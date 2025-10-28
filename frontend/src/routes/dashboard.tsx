import { DashboardSidebar } from '@/components/layout'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { createFileRoute, Outlet, redirect } from '@tanstack/react-router'
import { PATH_AUTH, PATH_ROOT } from '../constants/routes'
import { useAuthStore } from '../store/authStore'

export const Route = createFileRoute('/dashboard')({
  beforeLoad: () => {
    const user = useAuthStore.getState().user

    if (!user) {
      throw redirect({
        to: PATH_AUTH.login,
      })
    }

    if (!user.roles.includes('admin')) {
      throw redirect({
        to: PATH_ROOT.page404,
      })
    }
  },
  component: () => (
    <SidebarProvider>
      <DashboardSidebar />
      <SidebarInset>
        <main className="flex-1">
          <Outlet />
        </main>
      </SidebarInset>
    </SidebarProvider>
  ),
})
