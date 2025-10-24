import { TanStackDevtools } from '@tanstack/react-devtools'
import type { QueryClient } from '@tanstack/react-query'
import type { ErrorComponentProps } from '@tanstack/react-router'
import {
  createRootRouteWithContext,
  Outlet,
  useLocation,
} from '@tanstack/react-router'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'
import { AlertCircle, RefreshCw } from 'lucide-react'
import Footer from '../components/layout/Footer'
import Header from '../components/layout/Header'
import { Button } from '../components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card'
import { Toaster } from '../components/ui/sonner'
import { PATH_DASHBOARD } from '../constants'
import TanStackQueryDevtools from '../integrations/tanstack-query/devtools'
import { AuthProvider, NotificationProvider } from '../providers'

interface MyRouterContext {
  queryClient: QueryClient
}

function ErrorComponent({ error, reset }: ErrorComponentProps) {
  const errorMessage =
    error instanceof Error ? error.message : 'An unexpected error occurred'

  return (
    <div className="min-h-screen bg-gray-50/40 flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 h-12 w-12 rounded-full bg-red-100 flex items-center justify-center">
            <AlertCircle className="h-6 w-6 text-red-600" />
          </div>
          <CardTitle className="text-red-600">Something went wrong</CardTitle>
        </CardHeader>
        <CardContent className="text-center space-y-4">
          <p className="text-muted-foreground text-sm">{errorMessage}</p>
          <div className="space-y-2">
            <Button onClick={reset} className="w-full">
              <RefreshCw className="h-4 w-4 mr-2" />
              Try Again
            </Button>
            <Button
              onClick={() => window.location.reload()}
              variant="outline"
              className="w-full"
            >
              Reload Page
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export const Route = createRootRouteWithContext<MyRouterContext>()({
  errorComponent: ErrorComponent,
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
