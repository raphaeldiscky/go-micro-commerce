import { PATH_AUTH, PATH_ROOT } from '@/constants/routes'
import type { FileRoutesByPath } from '@tanstack/react-router'
import { createFileRoute, redirect } from '@tanstack/react-router'
import { SignupForm } from '../../components/auth/SignupForm'
import { AuthLayout } from '../../components/layout/AuthLayout'

export const Route = createFileRoute(
  PATH_AUTH.register as keyof FileRoutesByPath,
)({
  beforeLoad: () => {
    // Redirect to home if already authenticated
    const token = localStorage.getItem('access_token')
    if (token) {
      throw redirect({
        to: PATH_ROOT.home,
      })
    }
  },
  component: RegisterPage,
})

function RegisterPage() {
  return (
    <AuthLayout title="Create Account">
      <SignupForm />
    </AuthLayout>
  )
}
