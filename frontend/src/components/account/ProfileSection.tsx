import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type { ProfileFormValues } from '@/schemas/account'
import { passwordSchema, profileSchema } from '@/schemas/account'
import { useAccountStore } from '@/store/accountStore'
import { useUser } from '@/store/authStore'
import { useForm } from '@tanstack/react-form'
import { Loader2, Lock, Mail, User } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

export function ProfileSection() {
  const user = useUser()
  const updateProfile = useAccountStore((state) => state.updateProfile)
  const changePassword = useAccountStore((state) => state.changePassword)
  const isUpdating = useAccountStore((state) => state.isUpdating)

  const [activeTab, setActiveTab] = useState('profile')

  const defaultValues: ProfileFormValues = {
    firstName: user?.firstName || '',
    lastName: user?.lastName || '',
    email: user?.email || '',
  }
  const profileForm = useForm({
    defaultValues,
    validators: {
      onChange: profileSchema,
    },
    onSubmit: async ({ value }) => {
      try {
        await updateProfile(value)
        toast.success('Profile updated successfully')
      } catch (error) {
        console.error('Profile update failed:', error)
        toast.error('Failed to update profile')
      }
    },
  })

  // Password form
  const passwordForm = useForm({
    defaultValues: {
      currentPassword: '',
      newPassword: '',
      confirmPassword: '',
    },
    validators: {
      onChange: passwordSchema,
    },
    onSubmit: async ({ value }) => {
      try {
        await changePassword(value)
        toast.success('Password changed successfully')
        passwordForm.reset()
      } catch (error) {
        console.error('Password change failed:', error)
        toast.error('Failed to change password')
      }
    },
  })

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <User className="h-5 w-5" />
            Profile Information
          </CardTitle>
          <CardDescription>
            Manage your personal information and account settings
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="profile">Edit Profile</TabsTrigger>
              <TabsTrigger value="password">Change Password</TabsTrigger>
            </TabsList>

            <TabsContent value="profile" className="space-y-6">
              <div className="flex items-center gap-6">
                <Avatar className="h-20 w-20">
                  <AvatarImage alt={user?.firstName} />
                  <AvatarFallback className="text-lg">
                    {user?.firstName.charAt(0).toUpperCase() || 'U'}
                  </AvatarFallback>
                </Avatar>
                <div className="space-y-2">
                  <h3 className="text-lg font-medium">Profile Picture</h3>
                  <Button variant="outline" size="sm">
                    Change Avatar
                  </Button>
                  <p className="text-sm text-muted-foreground">
                    JPG, GIF or PNG. Max size of 2MB
                  </p>
                </div>
              </div>

              <Separator />

              <form
                onSubmit={(e) => {
                  e.preventDefault()
                  e.stopPropagation()
                  profileForm.handleSubmit()
                }}
                className="space-y-4"
              >
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <profileForm.Field
                    name="firstName"
                    validators={{
                      onChange: ({ value }) =>
                        !value.trim() ? 'First name is required' : undefined,
                    }}
                  >
                    {(field) => (
                      <div className="space-y-2">
                        <Label htmlFor="firstName">First Name</Label>
                        <Input
                          id="firstName"
                          value={field.state.value}
                          onChange={(e) => field.handleChange(e.target.value)}
                          placeholder="Enter your first name"
                        />
                      </div>
                    )}
                  </profileForm.Field>

                  <profileForm.Field
                    name="lastName"
                    validators={{
                      onChange: ({ value }) =>
                        !value.trim() ? 'Last name is required' : undefined,
                    }}
                  >
                    {(field) => (
                      <div className="space-y-2">
                        <Label htmlFor="lastName">Last Name</Label>
                        <Input
                          id="lastName"
                          value={field.state.value}
                          onChange={(e) => field.handleChange(e.target.value)}
                          placeholder="Enter your last name"
                        />
                      </div>
                    )}
                  </profileForm.Field>
                </div>

                <profileForm.Field
                  name="email"
                  validators={{
                    onChange: ({ value }) => {
                      if (!value.trim()) return 'Email is required'
                      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
                      if (!emailRegex.test(value))
                        return 'Please enter a valid email address'
                      return undefined
                    },
                  }}
                >
                  {(field) => (
                    <div className="space-y-2">
                      <Label
                        htmlFor="email"
                        className="flex items-center gap-2"
                      >
                        <Mail className="h-4 w-4" />
                        Email Address
                      </Label>
                      <Input
                        id="email"
                        type="email"
                        value={field.state.value}
                        onChange={(e) => field.handleChange(e.target.value)}
                        placeholder="Enter your email"
                      />
                    </div>
                  )}
                </profileForm.Field>

                <div className="flex justify-end">
                  <Button type="submit" disabled={isUpdating}>
                    {isUpdating ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Updating...
                      </>
                    ) : (
                      'Save Changes'
                    )}
                  </Button>
                </div>
              </form>
            </TabsContent>

            <TabsContent value="password" className="space-y-6">
              <form
                onSubmit={(e) => {
                  e.preventDefault()
                  e.stopPropagation()
                  passwordForm.handleSubmit()
                }}
                className="space-y-4"
              >
                <passwordForm.Field
                  name="currentPassword"
                  validators={{
                    onChange: ({ value }) =>
                      !value.trim()
                        ? 'Current password is required'
                        : undefined,
                  }}
                >
                  {(field) => (
                    <div className="space-y-2">
                      <Label
                        htmlFor="currentPassword"
                        className="flex items-center gap-2"
                      >
                        <Lock className="h-4 w-4" />
                        Current Password
                      </Label>
                      <Input
                        id="currentPassword"
                        type="password"
                        value={field.state.value}
                        onChange={(e) => field.handleChange(e.target.value)}
                        placeholder="Enter current password"
                      />
                    </div>
                  )}
                </passwordForm.Field>

                <passwordForm.Field
                  name="newPassword"
                  validators={{
                    onChange: ({ value }) => {
                      if (!value.trim()) return 'New password is required'
                      if (value.length < 8)
                        return 'Password must be at least 8 characters'
                      if (!/(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/.test(value)) {
                        return 'Password must contain at least one uppercase letter, one lowercase letter, and one number'
                      }
                      return undefined
                    },
                  }}
                >
                  {(field) => (
                    <div className="space-y-2">
                      <Label htmlFor="newPassword">New Password</Label>
                      <Input
                        id="newPassword"
                        type="password"
                        value={field.state.value}
                        onChange={(e) => field.handleChange(e.target.value)}
                        placeholder="Enter new password"
                      />

                      <p className="text-xs text-muted-foreground">
                        Must be at least 8 characters with uppercase, lowercase,
                        and numbers
                      </p>
                    </div>
                  )}
                </passwordForm.Field>

                <passwordForm.Field
                  name="confirmPassword"
                  validators={{
                    onChange: ({ value, fieldApi }) => {
                      if (!value.trim())
                        return 'Please confirm your new password'
                      if (
                        value !== fieldApi.form.getFieldValue('newPassword')
                      ) {
                        return 'Passwords do not match'
                      }
                      return undefined
                    },
                  }}
                >
                  {(field) => (
                    <div className="space-y-2">
                      <Label htmlFor="confirmPassword">
                        Confirm New Password
                      </Label>
                      <Input
                        id="confirmPassword"
                        type="password"
                        value={field.state.value}
                        onChange={(e) => field.handleChange(e.target.value)}
                        placeholder="Confirm new password"
                      />
                    </div>
                  )}
                </passwordForm.Field>

                <div className="flex justify-end">
                  <Button type="submit" disabled={isUpdating}>
                    {isUpdating ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Changing Password...
                      </>
                    ) : (
                      'Change Password'
                    )}
                  </Button>
                </div>
              </form>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  )
}
