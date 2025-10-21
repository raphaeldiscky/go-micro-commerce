import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { useOrderStore, useOrdersPagination } from '@/store/orderStore'
import { ChevronLeftIcon, ChevronRightIcon, Loader2Icon } from 'lucide-react'

interface OrderPaginationProps {
  className?: string
}

export function OrderPagination({ className }: OrderPaginationProps) {
  const pagination = useOrdersPagination()
  const { fetchOrders, isLoading } = useOrderStore()

  const handleLoadPrevious = async () => {
    if (pagination.previousCursor && !isLoading) {
      await fetchOrders(pagination.previousCursor)
    }
  }

  const handleLoadNext = async () => {
    if (pagination.nextCursor && !isLoading) {
      await fetchOrders(pagination.nextCursor)
    }
  }

  const hasNavigation = pagination.hasNextPage || pagination.hasPreviousPage

  if (!hasNavigation) {
    return null
  }

  return (
    <div className={cn('flex items-center justify-between', className)}>
      <div className="text-sm text-muted-foreground">
        {pagination.hasNextPage && <span>More orders available</span>}
      </div>

      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={handleLoadPrevious}
          disabled={!pagination.hasPreviousPage || isLoading}
          className="flex items-center gap-2"
        >
          <ChevronLeftIcon className="h-4 w-4" />
          Previous
        </Button>

        <Button
          variant="default"
          size="sm"
          onClick={handleLoadNext}
          disabled={!pagination.hasNextPage || isLoading}
          className="flex items-center gap-2"
        >
          {isLoading ? (
            <>
              <Loader2Icon className="h-4 w-4 animate-spin" />
              Loading...
            </>
          ) : (
            <>
              Next
              <ChevronRightIcon className="h-4 w-4" />
            </>
          )}
        </Button>
      </div>
    </div>
  )
}
