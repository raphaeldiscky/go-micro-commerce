import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { MockOrder } from '@/lib/mock-data/orders'
import { formatDate } from '@/lib/utils/date'
import { Eye } from 'lucide-react'

interface OrderTableProps {
  orders: Array<MockOrder>
}

function getStatusColor(status: MockOrder['status']) {
  switch (status) {
    case 'completed':
      return 'border-green-200 text-green-700 dark:border-green-800 dark:text-green-400'
    case 'processing':
      return 'border-blue-200 text-blue-700 dark:border-blue-800 dark:text-blue-400'
    case 'pending':
      return 'border-amber-200 text-amber-700 dark:border-amber-800 dark:text-amber-400'
    case 'cancelled':
      return 'border-red-200 text-red-700 dark:border-red-800 dark:text-red-400'
    default:
      return 'border-gray-200 text-gray-700 dark:border-gray-800 dark:text-gray-400'
  }
}

export function OrderTable({ orders }: OrderTableProps) {
  if (orders.length === 0) {
    return (
      <div className="flex h-64 items-center justify-center rounded-lg border border-dashed">
        <p className="text-muted-foreground">No orders found</p>
      </div>
    )
  }

  return (
    <div className="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Order ID</TableHead>
            <TableHead>Customer</TableHead>
            <TableHead>Items</TableHead>
            <TableHead>Total</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Date</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {orders.map((order) => (
            <TableRow key={order.id}>
              <TableCell className="font-medium">{order.id}</TableCell>
              <TableCell>
                <div>
                  <div className="font-medium">{order.customerName}</div>
                  <div className="text-sm text-muted-foreground">
                    {order.customerEmail}
                  </div>
                </div>
              </TableCell>
              <TableCell>{order.items}</TableCell>
              <TableCell>${order.total.toFixed(2)}</TableCell>
              <TableCell>
                <Badge
                  className={getStatusColor(order.status)}
                  variant="outline"
                >
                  {order.status}
                </Badge>
              </TableCell>
              <TableCell>{formatDate(order.createdAt)}</TableCell>
              <TableCell className="text-right">
                <button className="inline-flex h-8 w-8 items-center justify-center rounded-md hover:bg-accent">
                  <Eye className="h-4 w-4" />
                </button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
