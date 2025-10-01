import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { timestampToDate } from '@/lib/utils/date'
import type { Product } from '@/proto/product/v1/product_pb'
import { format } from 'date-fns'
import { Package } from 'lucide-react'

interface ProductCardProps {
  product: Product
}

export function ProductCard({ product }: ProductCardProps) {
  const availableQuantity = Number(product.quantity - product.reservedQuantity)
  const isLowStock = availableQuantity < 10 && availableQuantity > 0
  const isOutOfStock = availableQuantity === 0

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
            Rp {product.price.toFixed(2)}
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
