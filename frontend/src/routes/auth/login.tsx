import { createFileRoute, redirect } from '@tanstack/react-router'
import { AuthLayout } from '../../components/AuthLayout'
import { LoginForm } from '../../components/LoginForm'

export const Route = createFileRoute('/auth/login')({
  beforeLoad: () => {
    // Redirect to home if already authenticated
    // Note: This will be enhanced when we integrate the auth context
    const token = localStorage.getItem('access_token')
    if (token) {
      throw redirect({
        to: '/',
      })
    }
  },
  component: LoginPage,
})

function LoginPage() {
  return (
    <AuthLayout title="Sign In">
      <LoginForm />
    </AuthLayout>
  )
}
