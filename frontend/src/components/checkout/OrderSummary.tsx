import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { formatCurrency } from '@/data/mockData'
import { useOrderSummary } from '@/store/cartStore'
import { Receipt, ShoppingCart, Truck } from 'lucide-react'

export function OrderSummary() {
  const orderSummary = useOrderSummary()

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Receipt className="h-5 w-5" />
          Order Summary
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-3">
          {/* Subtotal */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-muted-foreground">
              <ShoppingCart className="h-4 w-4" />
              <span>Subtotal</span>
            </div>
            <span className="font-medium">
              {formatCurrency(orderSummary.subtotal)}
            </span>
          </div>

          {/* Shipping */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-muted-foreground">
              <Truck className="h-4 w-4" />
              <span>Shipping</span>
            </div>
            <span className="font-medium">
              {orderSummary.shipping === 0 ? (
                <span className="text-green-600">FREE</span>
              ) : (
                formatCurrency(orderSummary.shipping)
              )}
            </span>
          </div>

          
          <Separator />

          {/* Total */}
          <div className="flex items-center justify-between text-lg font-bold">
            <span>Total</span>
            <span>{formatCurrency(orderSummary.total)}</span>
          </div>
        </div>

        
        {orderSummary.shipping === 0 && orderSummary.subtotal >= 100 && (
          <div className="text-xs text-green-600 bg-green-50 dark:bg-green-950 p-3 rounded-lg">
            <p>🎉 Free shipping applied! (Orders over $100)</p>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
