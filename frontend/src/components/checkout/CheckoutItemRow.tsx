import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils/index'
import { fCurrency } from '@/lib/utils/number'
import type { EnrichedCartItem } from '@/store/cartStore'
import { Package } from 'lucide-react'

interface CheckoutItemRowProps {
  item: EnrichedCartItem
  className?: string
}

export function CheckoutItemRow({ item, className }: CheckoutItemRowProps) {
  if (!item.product) {
    return (
      <div
        className={cn(
          'flex items-center gap-4 p-4 border-b animate-pulse',
          className,
        )}
      >
        <div className="h-4 w-4 bg-muted rounded" />
        <div className="h-16 w-16 rounded-md bg-muted" />
        <div className="flex-1 space-y-2">
          <div className="h-4 bg-muted rounded w-3/4" />
          <div className="h-3 bg-muted rounded w-1/2" />
        </div>
      </div>
    )
  }

  const availableQuantity = Number(
    item.product.quantity - item.product.reservedQuantity,
  )
  const isLowStock = availableQuantity < 10 && availableQuantity > 0
  const isOutOfStock = availableQuantity === 0

  return (
    <div
      className={cn(
        'flex items-center gap-4 p-4 transition-all duration-200',
        isOutOfStock && 'opacity-60',
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
              {item.product.name}
            </h4>
            <p className="text-xs text-muted-foreground mt-1">
              <span className="font-mono">ID: {item.product.id}</span>
            </p>
            <div className="flex items-center gap-2 mt-2">
              <span className="font-semibold text-sm">
                {fCurrency(item.product.price)}
              </span>
              {isLowStock && (
                <Badge variant="secondary" className="text-xs">
                  Low Stock
                </Badge>
              )}
              {isOutOfStock && (
                <Badge variant="destructive" className="text-xs">
                  Out of Stock
                </Badge>
              )}
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
            Subtotal: {fCurrency(item.product.price * item.quantity)}
          </span>
        </div>
      </div>
    </div>
  )
}
