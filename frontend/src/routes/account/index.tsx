import { AccountStats } from '@/components/account/AccountStats'
import { AddressSection } from '@/components/account/AddressSection'
import { ProfileSection } from '@/components/account/ProfileSection'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useUser } from '@/store/authStore'
import { createFileRoute } from '@tanstack/react-router'
import { MapPin, User } from 'lucide-react'

export const Route = createFileRoute('/account/')({
  component: AccountPage,
})

function AccountPage() {
  const user = useUser()

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <div className="space-y-2">
          <h1 className="text-3xl font-bold tracking-tight">My Account</h1>
          <p className="text-muted-foreground">
            Manage your profile, addresses, and account settings
          </p>
          <p className="text-sm text-muted-foreground">
            Welcome back, {user?.firstName} {user?.lastName}
          </p>
        </div>
      </div>

      {/* Account Stats */}
      <div className="mb-8">
        <AccountStats />
      </div>

      {/* Account Management Tabs */}
      <Tabs defaultValue="profile" className="space-y-6">
        <TabsList className="grid w-full grid-cols-2 lg:w-96">
          <TabsTrigger value="profile" className="flex items-center gap-2">
            <User className="h-4 w-4" />
            Profile
          </TabsTrigger>
          <TabsTrigger value="addresses" className="flex items-center gap-2">
            <MapPin className="h-4 w-4" />
            Addresses
          </TabsTrigger>
        </TabsList>

        <TabsContent value="profile" className="space-y-6">
          <ProfileSection />
        </TabsContent>

        <TabsContent value="addresses" className="space-y-6">
          <AddressSection />
        </TabsContent>
      </Tabs>
    </div>
  )
}
