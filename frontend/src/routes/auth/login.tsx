import { PATH_AUTH, PATH_ROOT } from '@/constants/routes'
import type { FileRoutesByPath } from '@tanstack/react-router'
import { createFileRoute, redirect } from '@tanstack/react-router'
import { LoginForm } from '../../components/auth/LoginForm'
import { AuthLayout } from '../../components/layout/AuthLayout'

export const Route = createFileRoute(PATH_AUTH.login as keyof FileRoutesByPath)(
  {
    beforeLoad: () => {
      // Redirect to home if already authenticated
      // Note: This will be enhanced when we integrate the auth context
      const token = localStorage.getItem('access_token')
      if (token) {
        throw redirect({
          to: PATH_ROOT.home,
        })
      }
    },
    component: LoginPage,
  },
)

function LoginPage() {
  return (
    <AuthLayout title="Sign In">
      <LoginForm />
    </AuthLayout>
  )
}
