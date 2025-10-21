import { OrderCard } from '@/components/orders/OrderCard'
import { OrderFilters } from '@/components/orders/OrderFilters'
import { OrderPagination } from '@/components/orders/OrderPagination'
import { OrderSkeleton } from '@/components/orders/OrderSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import {
  useOrderStore,
  useOrders,
  useOrdersError,
  useOrdersInitialized,
  useOrdersLoading,
} from '@/store/orderStore'
import { AlertCircleIcon, InboxIcon } from 'lucide-react'
import { useEffect } from 'react'

interface OrderListProps {
  className?: string
}

export function OrderList({ className }: OrderListProps) {
  const orders = useOrders()
  const isLoading = useOrdersLoading()
  const error = useOrdersError()
  const hasInitialized = useOrdersInitialized()
  const { initialize } = useOrderStore()

  useEffect(() => {
    if (!hasInitialized) {
      initialize()
    }
  }, [hasInitialized, initialize])

  if (error) {
    return (
      <div className={cn('w-full max-w-4xl mx-auto p-6', className)}>
        <Alert variant="destructive">
          <AlertCircleIcon className="h-4 w-4" />
          <AlertDescription>
            Failed to load orders: {error}. Please try again later.
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  const hasOrders = orders.length > 0
  const showEmptyState = !isLoading && !hasOrders && hasInitialized

  return (
    <div className={cn('w-full max-w-4xl mx-auto space-y-6 p-6', className)}>
      {/* Filters */}
      <OrderFilters />

      {/* Loading State */}
      {isLoading && !hasOrders && <OrderSkeleton count={5} />}

      {/* Orders List */}
      {hasOrders && (
        <ScrollArea className="h-[calc(100vh-280px)] min-h-0">
          <div className="space-y-4">
            {orders.map((order) => (
              <OrderCard key={order.id} order={order} />
            ))}
          </div>
        </ScrollArea>
      )}

      {/* Empty State */}
      {showEmptyState && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <InboxIcon className="h-16 w-16 text-muted-foreground mb-4" />
          <h3 className="text-lg font-semibold mb-2">No orders found</h3>
          <p className="text-muted-foreground max-w-md">
            No orders match your current filters. Try adjusting your search
            criteria or clear all filters to see all orders.
          </p>
        </div>
      )}

      {/* Pagination */}
      {hasOrders && <OrderPagination className="pt-4 border-t" />}

      {/* Loading overlay for pagination */}
      {isLoading && hasOrders && (
        <div className="flex items-center justify-center py-4">
          <OrderSkeleton count={2} />
        </div>
      )}
    </div>
  )
}
