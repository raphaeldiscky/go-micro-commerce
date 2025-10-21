import { createFileRoute } from '@tanstack/react-router'
import { RevenueOverview } from '../../components/dashboard/analytics/RevenueOverview'
import { DashboardHeader } from '../../components/layout/DashboardHeader'
import { PATH_DASHBOARD } from '../../constants/routes'

export const Route = createFileRoute('/dashboard/revenue')({
  component: RevenuePage,
})

function RevenuePage() {
  return (
    <div>
      <DashboardHeader
        title="Revenue Analytics"
        breadcrumbs={[
          { label: 'Dashboard', href: PATH_DASHBOARD.root },
          { label: 'Revenue' },
        ]}
      />

      <div className="space-y-4 p-6">
        <RevenueOverview />
      </div>
    </div>
  )
}
