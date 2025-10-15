import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { formatCurrency } from '@/data/mockData'
import { cn } from '@/lib/utils/index'
import { useCartStore } from '@/store/cartStore'
import type { CartItem } from '@/types/cart'
import { Minus, Package, Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'

interface CartItemRowProps {
  item: CartItem
  className?: string
}

export function CartItemRow({ item, className }: CartItemRowProps) {
  const [isRemoving, setIsRemoving] = useState(false)
  const { updateQuantity, removeItem, toggleSelection } = useCartStore()

  const availableQuantity =
    item.product.quantity - item.product.reservedQuantity
  const isLowStock = availableQuantity < 10 && availableQuantity > 0
  const isOutOfStock = availableQuantity === 0

  const handleQuantityChange = (newQuantity: number) => {
    if (newQuantity > 0 && newQuantity <= availableQuantity) {
      updateQuantity(item.id, newQuantity)
    }
  }

  const handleRemove = () => {
    setIsRemoving(true)
    // Add a small delay for better UX
    setTimeout(() => {
      removeItem(item.id)
    }, 150)
  }

  const handleToggleSelection = () => {
    toggleSelection(item.id)
  }

  if (isRemoving) {
    return (
      <div
        className={cn(
          'flex items-center gap-4 p-4 border-b transition-all duration-150 opacity-50 scale-95',
          className,
        )}
      />
    )
  }

  return (
    <div
      className={cn(
        'flex items-center gap-4 p-4 border-b transition-all duration-200 hover:bg-muted/50',
        isOutOfStock && 'opacity-60',
        className,
      )}
    >
      {/* Selection Checkbox */}
      <Checkbox
        checked={item.selected_for_checkout}
        disabled={isOutOfStock}
        onCheckedChange={handleToggleSelection}
      />

      {/* Product Image */}
      <div className="flex-shrink-0">
        <div className="h-16 w-16 rounded-md bg-muted flex items-center justify-center overflow-hidden">
          {item.product.image ? (
            <img
              alt={item.product.name}
              className="h-full w-full object-cover"
              src={item.product.image}
            />
          ) : (
            <Package className="h-8 w-8 text-muted-foreground" />
          )}
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
              {item.product.sku && (
                <span className="font-mono">SKU: {item.product.sku}</span>
              )}
            </p>
            <div className="flex items-center gap-2 mt-2">
              <span className="font-semibold text-sm">
                {formatCurrency(item.product.price)}
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
          <span className="text-xs text-muted-foreground">Qty:</span>
          <div className="flex items-center border rounded-md">
            <Button
              disabled={item.quantity <= 1 || isOutOfStock}
              onClick={() => handleQuantityChange(item.quantity - 1)}
              size="sm"
              variant="ghost"
              className="h-8 w-8 p-0"
            >
              <Minus className="h-3 w-3" />
            </Button>
            <div className="w-12 text-center text-sm font-medium">
              {item.quantity}
            </div>
            <Button
              disabled={item.quantity >= availableQuantity || isOutOfStock}
              onClick={() => handleQuantityChange(item.quantity + 1)}
              size="sm"
              variant="ghost"
              className="h-8 w-8 p-0"
            >
              <Plus className="h-3 w-3" />
            </Button>
          </div>
          {/* Available quantity indicator */}
          <span className="text-xs text-muted-foreground">
            ({availableQuantity} available)
          </span>
        </div>

        {/* Item Subtotal */}
        <div className="mt-2">
          <span className="text-sm font-medium">
            Subtotal: {formatCurrency(item.product.price * item.quantity)}
          </span>
        </div>
      </div>

      {/* Remove Button */}

      <Button
        disabled={isRemoving}
        onClick={handleRemove}
        size="sm"
        variant="ghost"
        className="flex-shrink-0 h-8 w-8 p-0 text-muted-foreground hover:text-destructive"
      >
        <Trash2 className="h-4 w-4" />
      </Button>
    </div>
  )
}
