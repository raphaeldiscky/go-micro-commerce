import { PATH_AUTH } from '@/constants/routes'
import { useRegister } from '@/hooks/auth/useAuth'
import { useForm } from '@tanstack/react-form'
import { Link } from '@tanstack/react-router'
import { Eye, EyeOff, UserPlus } from 'lucide-react'
import { useState } from 'react'
import { Button } from '../ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card'
import { Input } from '../ui/input'
import { Label } from '../ui/label'

interface SignupFormProps {
  onSuccess?: () => void
}

export function SignupForm({ onSuccess }: SignupFormProps) {
  const [showPassword, setShowPassword] = useState(false)
  const registerMutation = useRegister()

  const form = useForm({
    defaultValues: {
      email: '',
      firstName: '',
      lastName: '',
      password: '',
      username: '',
    },
    onSubmit: async ({ value }) => {
      try {
        await registerMutation.mutateAsync(value)
        onSuccess?.()
      } catch (error) {
        // Error is handled by the mutation
      }
    },
  })

  const isLoading = registerMutation.isPending

  return (
    <Card className="w-full max-w-md mx-auto">
      <CardHeader className="text-center">
        <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
          <UserPlus className="h-6 w-6 text-primary" />
        </div>
        <CardTitle className="text-2xl font-bold">Create account</CardTitle>
        <p className="text-muted-foreground">
          Sign up to get started with Go Micro Commerce
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
          {/* Name Fields */}
          <div className="grid grid-cols-2 gap-4">
            <form.Field
              name="firstName"
              validators={{
                onChange: ({ value }) => {
                  if (!value) return 'First name is required'
                  if (value.length < 1) return 'First name is too short'
                  if (value.length > 50) return 'First name is too long'
                  return undefined
                },
              }}
            >
              {(field) => (
                <div className="space-y-2">
                  <Label htmlFor="firstName">First Name</Label>
                  <Input
                    className={
                      field.state.meta.errors.length > 0
                        ? 'border-destructive'
                        : ''
                    }
                    disabled={isLoading}
                    id="firstName"
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    placeholder="First name"
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

            <form.Field
              name="lastName"
              validators={{
                onChange: ({ value }) => {
                  if (!value) return 'Last name is required'
                  if (value.length < 1) return 'Last name is too short'
                  if (value.length > 50) return 'Last name is too long'
                  return undefined
                },
              }}
            >
              {(field) => (
                <div className="space-y-2">
                  <Label htmlFor="lastName">Last Name</Label>
                  <Input
                    className={
                      field.state.meta.errors.length > 0
                        ? 'border-destructive'
                        : ''
                    }
                    disabled={isLoading}
                    id="lastName"
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    placeholder="Last name"
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
          </div>

          {/* Username Field */}
          <form.Field
            name="username"
            validators={{
              onChange: ({ value }) => {
                if (!value) return 'Username is required'
                if (value.length < 3)
                  return 'Username must be at least 3 characters'
                if (value.length > 50) return 'Username is too long'
                if (!/^[a-zA-Z0-9_-]+$/.test(value))
                  return 'Username can only contain letters, numbers, hyphens and underscores'
                return undefined
              },
            }}
          >
            {(field) => (
              <div className="space-y-2">
                <Label htmlFor="username">Username</Label>
                <Input
                  className={
                    field.state.meta.errors.length > 0
                      ? 'border-destructive'
                      : ''
                  }
                  disabled={isLoading}
                  id="username"
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  placeholder="Choose a username"
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
                if (value.length < 8)
                  return 'Password must be at least 8 characters'
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
                    placeholder="Create a password"
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
                <p className="text-xs text-muted-foreground">
                  Password must be at least 8 characters long
                </p>
              </div>
            )}
          </form.Field>

          {/* Error Message */}
          {registerMutation.isError && (
            <div className="p-3 text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-md">
              {registerMutation.error.message ||
                'Registration failed. Please try again.'}
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
                Creating account...
              </div>
            ) : (
              <div className="flex items-center gap-2">
                <UserPlus className="h-4 w-4" />
                Create Account
              </div>
            )}
          </Button>

          {/* Link to Login */}
          <div className="text-center text-sm">
            <span className="text-muted-foreground">
              Already have an account?{' '}
            </span>
            <Link
              className="text-primary hover:underline font-medium"
              to={PATH_AUTH.login}
            >
              Sign in
            </Link>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
