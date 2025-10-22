import { Card, CardContent } from '@/components/ui/card'
import { useProducts } from '@/hooks/dashboard'
import type { ProductStatus } from '@/mocks/products'
import { createFileRoute } from '@tanstack/react-router'
import { Loader2 } from 'lucide-react'
import { useState } from 'react'
import { ProductFilters } from '../../components/dashboard/products/ProductFilters'
import { ProductTable } from '../../components/dashboard/products/ProductTable'
import { CursorPagination } from '../../components/dashboard/shared/CursorPagination'
import { DashboardHeader } from '../../components/layout/DashboardHeader'
import { PATH_DASHBOARD } from '../../constants/routes'

export const Route = createFileRoute('/dashboard/products')({
  component: ProductsPage,
})

function ProductsPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<ProductStatus | 'all'>('all')
  const [currentCursor, setCurrentCursor] = useState<string | undefined>()
  const [cursorHistory, setCursorHistory] = useState<Array<string>>([])

  const { data: paginatedData, isLoading } = useProducts({
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

  const handleStatusChange = (value: ProductStatus | 'all') => {
    setStatusFilter(value)
    setCurrentCursor(undefined)
    setCursorHistory([])
  }

  return (
    <div>
      <DashboardHeader
        title="Products Management"
        breadcrumbs={[
          { label: 'Dashboard', href: PATH_DASHBOARD.root },
          { label: 'Products' },
        ]}
      />

      <div className="space-y-4 p-6">
        <ProductFilters
          searchQuery={searchQuery}
          statusFilter={statusFilter}
          onSearchChange={handleSearchChange}
          onStatusChange={handleStatusChange}
          onReset={handleReset}
        />

        <Card>
          <CardContent className="p-0">
            {isLoading ? (
              <div className="flex items-center justify-center py-12">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              </div>
            ) : (
              <ProductTable products={paginatedData?.data || []} />
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
