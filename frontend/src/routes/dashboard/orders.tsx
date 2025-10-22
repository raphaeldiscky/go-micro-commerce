import { Card, CardContent } from '@/components/ui/card'
import type { OrderStatus } from '@/data/orders'
import { useOrders } from '@/hooks/dashboard'
import { createFileRoute } from '@tanstack/react-router'
import { Loader2 } from 'lucide-react'
import { useState } from 'react'
import { OrderFilters } from '../../components/dashboard/orders/OrderFilters'
import { OrderTable } from '../../components/dashboard/orders/OrderTable'
import { CursorPagination } from '../../components/dashboard/shared/CursorPagination'
import { DashboardHeader } from '../../components/layout/DashboardHeader'
import { PATH_DASHBOARD } from '../../constants/routes'

export const Route = createFileRoute('/dashboard/orders')({
  component: OrdersPage,
})

function OrdersPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<OrderStatus | 'all'>('all')
  const [currentCursor, setCurrentCursor] = useState<string | undefined>()
  const [cursorHistory, setCursorHistory] = useState<Array<string>>([])

  const { data: paginatedData, isLoading } = useOrders({
    searchQuery,
    status: statusFilter,
    cursor: currentCursor,
  })

  const handleReset = () => {
    setSearchQuery('')
    setStatusFilter('all')
    setCurrentCursor(undefined)
    setCursorHistory([])
  }

  const handleNextPage = () => {
    if (paginatedData?.endCursor) {
      // Add current cursor to history before moving forward (if not first page)
      if (currentCursor) {
        setCursorHistory((prev) => [...prev, currentCursor])
      }
      setCurrentCursor(paginatedData.endCursor)
    }
  }

  const handlePreviousPage = () => {
    if (cursorHistory.length > 0) {
      // Go back to previous cursor
      const previousCursors = [...cursorHistory]
      const previousCursor = previousCursors.pop()
      setCursorHistory(previousCursors)
      setCurrentCursor(previousCursor)
    } else {
      // Go to first page
      setCurrentCursor(undefined)
      setCursorHistory([])
    }
  }

  // Reset cursor when filters change
  const handleSearchChange = (value: string) => {
    setSearchQuery(value)
    setCurrentCursor(undefined)
    setCursorHistory([])
  }

  const handleStatusChange = (value: OrderStatus | 'all') => {
    setStatusFilter(value)
    setCurrentCursor(undefined)
    setCursorHistory([])
  }

  return (
    <div>
      <DashboardHeader
        title="Orders Management"
        breadcrumbs={[
          { label: 'Dashboard', href: PATH_DASHBOARD.root },
          { label: 'Orders' },
        ]}
      />

      <div className="space-y-4 p-6">
        <OrderFilters
          searchQuery={searchQuery}
          statusFilter={statusFilter}
          onSearchChange={handleSearchChange}
          onStatusChange={handleStatusChange}
          onReset={handleReset}
        />

        <Card>
          <CardContent className="p-0">
            {isLoading ? (
              <div className="flex items-center justify-center p-8">
                <Loader2 className="h-8 w-8 animate-spin" />
              </div>
            ) : (
              <OrderTable orders={paginatedData?.data || []} />
            )}
          </CardContent>
        </Card>

        {paginatedData && (
          <CursorPagination
            hasNextPage={paginatedData.hasNextPage}
            hasPreviousPage={paginatedData.hasPreviousPage}
            onNextPage={handleNextPage}
            onPreviousPage={handlePreviousPage}
            disabled={isLoading}
          />
        )}
      </div>
    </div>
  )
}
