import { OrderStatusBadge } from '@/components/orders/OrderStatusBadge'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { PATH } from '@/constants/routes'
import { cn } from '@/lib/utils'
import { fCurrency } from '@/lib/utils/number'
import type { Order } from '@/types/__generated__/graphql'
import { useNavigate } from '@tanstack/react-router'
import { format } from 'date-fns'
import Decimal from 'decimal.js'
import {
  CalendarIcon,
  CreditCardIcon,
  DollarSignIcon,
  UserIcon,
} from 'lucide-react'

interface OrderCardProps {
  order: Order
  className?: string
}

export function OrderCard({ order, className }: OrderCardProps) {
  const navigate = useNavigate()

  const formatDate = (dateString: string) => {
    try {
      return format(new Date(dateString), 'MMM dd, yyyy HH:mm')
    } catch {
      return dateString
    }
  }

  const handleClick = () => {
    navigate({ to: PATH.orders.detail(order.id) })
  }

  return (
    <Card
      className={cn(
        'w-full transition-all hover:shadow-lg cursor-pointer hover:bg-accent/50',
        className,
      )}
      onClick={handleClick}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <h3 className="font-semibold text-lg leading-tight">
              Order {order.id.slice(-12)}
            </h3>
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <div className="flex items-center gap-1">
                <CalendarIcon className="h-3 w-3" />
                <span>{formatDate(order.createdAt)}</span>
              </div>
              <div className="flex items-center gap-1">
                <UserIcon className="h-3 w-3" />
                <span>{order.customerId.slice(-8)}</span>
              </div>
            </div>
          </div>
          <OrderStatusBadge status={order.status} />
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        <div className="flex items-center gap-2 text-sm">
          <Badge variant="outline" className="flex items-center gap-1">
            <CreditCardIcon className="h-3 w-3" />
            {order.paymentGateway}
          </Badge>
          <Badge variant="outline" className="flex items-center gap-1">
            <DollarSignIcon className="h-3 w-3" />
            {order.currency}
          </Badge>
        </div>

        <Separator />

        <div className="space-y-2">
          <h4 className="text-sm font-medium text-muted-foreground">
            Order Summary
          </h4>
          <div className="space-y-1 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">Subtotal</span>
              <span>{fCurrency(order.subtotal, order.currency)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Shipping</span>
              <span>{fCurrency(order.shippingCost, order.currency)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Tax</span>
              <span>{fCurrency(order.totalTax, order.currency)}</span>
            </div>
            {Decimal(order.totalDiscount).greaterThan(new Decimal(0)) && (
              <div className="flex justify-between text-green-600">
                <span>Discount</span>
                <span>-{fCurrency(order.totalDiscount, order.currency)}</span>
              </div>
            )}
            <Separator className="my-2" />
            <div className="flex justify-between font-semibold text-base">
              <span>Total</span>
              <span className="text-primary">
                {fCurrency(order.totalPrice, order.currency)}
              </span>
            </div>
          </div>
        </div>

        <div className="pt-2">
          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <span>ID: {order.idempotencyKey.slice(-12)}</span>
            <span>Updated: {formatDate(order.updatedAt)}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
