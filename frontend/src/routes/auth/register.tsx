import { createFileRoute, redirect } from '@tanstack/react-router'
import { AuthLayout } from '../../components/AuthLayout'
import { SignupForm } from '../../components/SignupForm'

export const Route = createFileRoute('/auth/register')({
  beforeLoad: () => {
    // Redirect to home if already authenticated
    const token = localStorage.getItem('access_token')
    if (token) {
      throw redirect({
        to: '/',
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
