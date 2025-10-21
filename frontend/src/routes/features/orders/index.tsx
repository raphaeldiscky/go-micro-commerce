import { OrderList } from '@/components/orders'
import { createFileRoute } from '@tanstack/react-router'
import { ReceiptIcon } from 'lucide-react'

export const Route = createFileRoute('/features/orders/')({
  component: OrdersPage,
})

function OrdersPage() {
  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-2">
          <ReceiptIcon className="h-8 w-8 text-primary" />
          <div className="space-y-1">
            <h1 className="text-3xl font-bold tracking-tight">
              Order Transactions
            </h1>
            <p className="text-muted-foreground">
              View and manage all your order transactions in one place
            </p>
          </div>
        </div>
      </div>

      {/* Orders List */}
      <OrderList />
    </div>
  )
}
