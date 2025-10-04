import { PATH_ROOT } from '@/constants/routes'
import { createFileRoute, redirect } from '@tanstack/react-router'
import { SignupForm } from '../../components/auth/SignupForm'
import { AuthLayout } from '../../components/layout/AuthLayout'

export const Route = createFileRoute('/auth/register')({
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
