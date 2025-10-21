import { TanStackDevtools } from '@tanstack/react-devtools'
import type { QueryClient } from '@tanstack/react-query'
import {
  createRootRouteWithContext,
  Outlet,
  useLocation,
} from '@tanstack/react-router'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'
import Footer from '../components/layout/Footer'
import Header from '../components/layout/Header'
import { Toaster } from '../components/ui/sonner'
import { PATH_DASHBOARD } from '../constants'
import TanStackQueryDevtools from '../integrations/tanstack-query/devtools'
import { AuthProvider, NotificationProvider } from '../providers'

interface MyRouterContext {
  queryClient: QueryClient
}

export const Route = createRootRouteWithContext<MyRouterContext>()({
  component: () => {
    const location = useLocation()
    const currentPathname = location.pathname
    const isDashboard = currentPathname.startsWith(PATH_DASHBOARD.root)
    return (
      <AuthProvider>
        <NotificationProvider>
          <div className="min-h-screen flex flex-col">
            {!isDashboard && <Header />}
            <main className="flex-1">
              <Outlet />
            </main>
            {!isDashboard && <Footer />}
            <TanStackDevtools
              config={{
                position: 'bottom-left',
              }}
              plugins={[
                {
                  name: 'Tanstack Router',
                  render: <TanStackRouterDevtoolsPanel />,
                },
                TanStackQueryDevtools,
              ]}
            />
            <Toaster position="bottom-left" />
          </div>
        </NotificationProvider>
      </AuthProvider>
    )
  },
})
