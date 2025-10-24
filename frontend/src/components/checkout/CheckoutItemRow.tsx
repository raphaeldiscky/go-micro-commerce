import { cn } from '@/lib/utils/index'
import { fCurrency } from '@/lib/utils/number'
import { Package } from 'lucide-react'
import type { CheckoutSessionItem } from '../../types/__generated__/graphql'

interface CheckoutItemRowProps {
  item: CheckoutSessionItem
  className?: string
}

export function CheckoutItemRow({ item, className }: CheckoutItemRowProps) {
  return (
    <div
      className={cn(
        'flex items-center gap-4 p-4 transition-all duration-200',
        className,
      )}
    >
      {/* Product Image */}
      <div className="flex-shrink-0">
        <div className="h-16 w-16 rounded-md bg-muted flex items-center justify-center overflow-hidden">
          <Package className="h-8 w-8 text-muted-foreground" />
        </div>
      </div>

      {/* Product Details */}
      <div className="flex-1 min-w-0">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0 flex-1">
            <h4 className="font-medium text-sm leading-tight truncate">
              {item.productName}
            </h4>
            <p className="text-xs text-muted-foreground mt-1">
              <span className="font-mono">ID: {item.productId}</span>
            </p>
            <div className="flex items-center gap-2 mt-2">
              <span className="font-semibold text-sm">
                {fCurrency(item.unitPrice)}
              </span>
            </div>
          </div>
        </div>

        {/* Quantity Controls */}
        <div className="flex items-center gap-2 mt-3">
          <span className="text-xs text-muted-foreground">
            Qty: {item.quantity}
          </span>
        </div>

        {/* Item Subtotal */}
        <div className="mt-2">
          <span className="text-sm font-medium">
            Subtotal: {fCurrency(Number(item.unitPrice) * item.quantity)}
          </span>
        </div>
      </div>
    </div>
  )
}
