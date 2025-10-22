import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { mockOrderStats, mockRevenueStats } from '@/mocks/analytics'
import { createFileRoute } from '@tanstack/react-router'
import { DollarSign, Package, ShoppingCart, Users } from 'lucide-react'
import { OrderOverview } from '../../components/dashboard/analytics/OrderOverview'
import { RevenueOverview } from '../../components/dashboard/analytics/RevenueOverview'
import { DashboardHeader } from '../../components/layout/DashboardHeader'

export const Route = createFileRoute('/dashboard/')({
  component: DashboardPage,
})

function DashboardPage() {
  const { totalRevenue } = mockRevenueStats
  const { totalOrders } = mockOrderStats

  return (
    <div>
      <DashboardHeader title="Overview" />

      <div className="space-y-6 p-6">
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Total Revenue
              </CardTitle>
              <DollarSign className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                ${totalRevenue.toLocaleString()}
              </div>
              <p className="text-xs text-muted-foreground">Last 30 days</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Total Orders
              </CardTitle>
              <ShoppingCart className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{totalOrders}</div>
              <p className="text-xs text-muted-foreground">All time</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Active Products
              </CardTitle>
              <Package className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">234</div>
              <p className="text-xs text-muted-foreground">In stock</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Users</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">1,429</div>
              <p className="text-xs text-muted-foreground">Registered</p>
            </CardContent>
          </Card>
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          <div>
            <h2 className="mb-4 text-lg font-semibold">Revenue</h2>
            <RevenueOverview />
          </div>

          <div>
            <h2 className="mb-4 text-lg font-semibold">Orders</h2>
            <OrderOverview />
          </div>
        </div>
      </div>
    </div>
  )
}
