import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { cn } from '@/lib/utils'
import { timestampToDate } from '@/lib/utils/date'
import { fCurrency } from '@/lib/utils/number'
import type { Product } from '@/proto/product/v1/product_pb'
import { useCartStore } from '@/store/cartStore'
import { format } from 'date-fns'
import {
  Check,
  Loader2,
  Minus,
  Package,
  Plus,
  ShoppingCart,
} from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

interface ProductCardProps {
  product: Product
}

export function ProductCard({ product }: ProductCardProps) {
  const [quantity, setQuantity] = useState(1)
  const [isAdding, setIsAdding] = useState(false)
  const [showSuccess, setShowSuccess] = useState(false)
  const { addItem } = useCartStore()

  const availableQuantity = Number(product.quantity - product.reservedQuantity)
  const isLowStock = availableQuantity < 10 && availableQuantity > 0
  const isOutOfStock = availableQuantity === 0

  const handleAddToCart = () => {
    if (isOutOfStock || quantity > availableQuantity || availableQuantity <= 0)
      return

    setIsAdding(true)
    try {
      addItem(product.id, quantity)
      setQuantity(1)
      console.log('Product added to cart:', product.name, 'Quantity:', quantity)

      // Show success animation
      setShowSuccess(true)

      // Reset success state after animation
      setTimeout(() => {
        setShowSuccess(false)
      }, 2000)
    } catch (error) {
      console.error('Failed to add to cart:', error)
      toast.error('Failed to add item to cart. Please try again.')
    } finally {
      setIsAdding(false)
    }
  }

  const handleQuantityChange = (newQuantity: number) => {
    if (newQuantity > 0 && newQuantity <= availableQuantity) {
      setQuantity(newQuantity)
    }
  }

  return (
    <Card className="hover:shadow-lg transition-shadow">
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="flex items-center space-x-2">
            <Package className="h-5 w-5 text-blue-600 dark:text-blue-400" />
            <div>
              <CardTitle className="text-lg">{product.name}</CardTitle>
              <CardDescription className="text-xs font-mono">
                ID: {product.id.slice(0, 8)}...
              </CardDescription>
            </div>
          </div>
          {isOutOfStock && (
            <Badge variant="destructive" className="text-xs">
              Out of Stock
            </Badge>
          )}
          {isLowStock && (
            <Badge variant="secondary" className="text-xs">
              Low Stock
            </Badge>
          )}
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Price */}
        <div>
          <p className="text-2xl font-bold text-gray-900 dark:text-white">
            {fCurrency(product.price, 'IDR')}
          </p>
        </div>

        {/* Inventory Info */}
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span className="text-gray-600 dark:text-gray-400">
              Total Quantity:
            </span>
            <span className="font-medium">{product.quantity.toString()}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-gray-600 dark:text-gray-400">Reserved:</span>
            <span className="font-medium text-orange-600 dark:text-orange-400">
              {product.reservedQuantity.toString()}
            </span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-gray-600 dark:text-gray-400">Available:</span>
            <span className="font-medium text-green-600 dark:text-green-400">
              {availableQuantity}
            </span>
          </div>
        </div>

        {/* Add to Cart Section */}
        <div className="pt-4 border-t border-gray-200 dark:border-gray-700 space-y-3">
          {isOutOfStock ? (
            <Button disabled className="w-full" variant="outline">
              <ShoppingCart className="h-4 w-4 mr-2" />
              Out of Stock
            </Button>
          ) : (
            <div className="space-y-3">
              {/* Quantity Selector */}
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">Quantity:</span>
                <div className="flex items-center border rounded-md">
                  <Button
                    disabled={quantity <= 1}
                    onClick={() => handleQuantityChange(quantity - 1)}
                    size="sm"
                    variant="ghost"
                    className="h-8 w-8 p-0"
                  >
                    <Minus className="h-3 w-3" />
                  </Button>
                  <div className="w-12 text-center text-sm font-medium">
                    {quantity}
                  </div>
                  <Button
                    disabled={quantity >= availableQuantity}
                    onClick={() => handleQuantityChange(quantity + 1)}
                    size="sm"
                    variant="ghost"
                    className="h-8 w-8 p-0"
                  >
                    <Plus className="h-3 w-3" />
                  </Button>
                </div>
              </div>

              {/* Add to Cart Button */}
              <Button
                disabled={isAdding || availableQuantity <= 0}
                onClick={handleAddToCart}
                className={cn(
                  'w-full transition-all duration-300',
                  showSuccess &&
                    'bg-green-600 hover:bg-green-700 border-green-600',
                )}
                size="lg"
              >
                {isAdding ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Adding...
                  </>
                ) : showSuccess ? (
                  <>
                    <Check className="h-4 w-4" />
                    Added!
                  </>
                ) : (
                  <>
                    <ShoppingCart className="h-4 w-4 mr-2" />
                    Add to Cart
                  </>
                )}
              </Button>

              {isLowStock && (
                <p className="text-xs text-amber-600 dark:text-amber-400 text-center">
                  Only {availableQuantity} items available
                </p>
              )}
            </div>
          )}
        </div>

        {/* Metadata */}
        <div className="pt-4 border-t border-gray-200 dark:border-gray-700 space-y-1">
          <div className="flex justify-between text-xs text-gray-500 dark:text-gray-400">
            <span>Version:</span>
            <span className="font-mono">{product.version.toString()}</span>
          </div>
          {product.createdAt && (
            <div className="flex justify-between text-xs text-gray-500 dark:text-gray-400">
              <span>Created:</span>
              <span>
                {format(timestampToDate(product.createdAt)!, 'MMM d, yyyy')}
              </span>
            </div>
          )}
          {product.updatedAt && (
            <div className="flex justify-between text-xs text-gray-500 dark:text-gray-400">
              <span>Updated:</span>
              <span>
                {format(timestampToDate(product.updatedAt)!, 'MMM d, yyyy')}
              </span>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
