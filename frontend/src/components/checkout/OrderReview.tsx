import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { useCartData } from '@/store/cartStore'
import { useMemo } from 'react'
import { CheckoutItemRow } from './CheckoutItemRow'

export function OrderReview() {
  // Get raw state with shallow comparison
  const { items: cartItems, productsMap } = useCartData()

  // Transform in useMemo - only recalculates when dependencies change
  const selectedItems = useMemo(() => {
    return cartItems
      .map((item) => ({
        ...item,
        product: productsMap.get(item.productId),
      }))
      .filter((item) => item.selectedForCheckout)
  }, [cartItems, productsMap])

  if (selectedItems.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Order Review</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground">
            No items selected for checkout
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Order Review ({selectedItems.length} items)</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {selectedItems.map((item) => (
          <div key={item.id} className="space-y-2">
            <CheckoutItemRow item={item} />
            {selectedItems.indexOf(item) < selectedItems.length - 1 && (
              <Separator />
            )}
          </div>
        ))}
      </CardContent>
    </Card>
  )
}
