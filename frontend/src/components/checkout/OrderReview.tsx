import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { useSelectedItems } from '@/store/cartStore'
import type { CartItem } from '@/types/cart'
import { CartItemRow } from '../cart/CartItemRow'

export function OrderReview() {
  const selectedItems = useSelectedItems()

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
        {selectedItems.map((item: CartItem) => (
          <div key={item.id} className="space-y-2">
            <CartItemRow item={item} />
            {selectedItems.indexOf(item) < selectedItems.length - 1 && (
              <Separator />
            )}
          </div>
        ))}
      </CardContent>
    </Card>
  )
}
