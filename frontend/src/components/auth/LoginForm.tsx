import { useLogin } from '@/hooks/useAuth'
import { useForm } from '@tanstack/react-form'
import { Link } from '@tanstack/react-router'
import { Eye, EyeOff, LogIn } from 'lucide-react'
import { useState } from 'react'
import { Button } from '../ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card'
import { Input } from '../ui/input'
import { Label } from '../ui/label'

interface LoginFormProps {
  onSuccess?: () => void
}

export function LoginForm({ onSuccess }: LoginFormProps) {
  const [showPassword, setShowPassword] = useState(false)
  const loginMutation = useLogin()

  const form = useForm({
    defaultValues: {
      email: '',
      password: '',
    },
    onSubmit: async ({ value }) => {
      try {
        await loginMutation.mutateAsync(value)
        onSuccess?.()
      } catch (error) {
        // Error is handled by the mutation
      }
    },
  })

  const isLoading = loginMutation.isPending

  return (
    <Card className="w-full max-w-md mx-auto">
      <CardHeader className="text-center">
        <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
          <LogIn className="h-6 w-6 text-primary" />
        </div>
        <CardTitle className="text-2xl font-bold">Welcome back</CardTitle>
        <p className="text-muted-foreground">
          Sign in to your account to continue
        </p>
      </CardHeader>
      <CardContent>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault()
            form.handleSubmit()
          }}
        >
          {/* Email Field */}
          <form.Field
            name="email"
            validators={{
              onChange: ({ value }) => {
                if (!value) return 'Email is required'
                if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value))
                  return 'Please enter a valid email address'
                return undefined
              },
            }}
          >
            {(field) => (
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  className={
                    field.state.meta.errors.length > 0
                      ? 'border-destructive'
                      : ''
                  }
                  disabled={isLoading}
                  id="email"
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  placeholder="Enter your email"
                  type="email"
                  value={field.state.value}
                />
                {field.state.meta.errors.length > 0 && (
                  <p className="text-sm text-destructive">
                    {field.state.meta.errors[0]}
                  </p>
                )}
              </div>
            )}
          </form.Field>

          {/* Password Field */}
          <form.Field
            name="password"
            validators={{
              onChange: ({ value }) => {
                if (!value) return 'Password is required'
                return undefined
              },
            }}
          >
            {(field) => (
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <div className="relative">
                  <Input
                    className={
                      field.state.meta.errors.length > 0
                        ? 'border-destructive pr-10'
                        : 'pr-10'
                    }
                    disabled={isLoading}
                    id="password"
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    placeholder="Enter your password"
                    type={showPassword ? 'text' : 'password'}
                    value={field.state.value}
                  />
                  <Button
                    className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                    disabled={isLoading}
                    onClick={() => setShowPassword(!showPassword)}
                    size="sm"
                    type="button"
                    variant="ghost"
                  >
                    {showPassword ? (
                      <EyeOff className="h-4 w-4 text-muted-foreground" />
                    ) : (
                      <Eye className="h-4 w-4 text-muted-foreground" />
                    )}
                  </Button>
                </div>
                {field.state.meta.errors.length > 0 && (
                  <p className="text-sm text-destructive">
                    {field.state.meta.errors[0]}
                  </p>
                )}
              </div>
            )}
          </form.Field>

          {/* Error Message */}
          {loginMutation.isError && (
            <div className="p-3 text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-md">
              {loginMutation.error.message || 'Login failed. Please try again.'}
            </div>
          )}

          {/* Submit Button */}
          <Button
            className="w-full"
            disabled={isLoading || !form.state.isValid}
            type="submit"
          >
            {isLoading ? (
              <div className="flex items-center gap-2">
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                Signing in...
              </div>
            ) : (
              <div className="flex items-center gap-2">
                <LogIn className="h-4 w-4" />
                Sign In
              </div>
            )}
          </Button>

          {/* Link to Register */}
          <div className="text-center text-sm">
            <span className="text-muted-foreground">
              Don't have an account?{' '}
            </span>
            <Link
              className="text-primary hover:underline font-medium"
              to="/auth/register"
            >
              Sign up
            </Link>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
