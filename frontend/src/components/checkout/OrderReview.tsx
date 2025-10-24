import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useCheckoutSession } from '../../store/checkoutSessionStore'
import { CheckoutItemRow } from './CheckoutItemRow'

export function OrderReview() {
  const data = useCheckoutSession()

  const checkoutItems = data?.items ?? []

  if (checkoutItems.length === 0) {
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
        <CardTitle>Order Review ({checkoutItems.length} items)</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {checkoutItems.map((item) => (
          <div key={item.id} className="space-y-2">
            <CheckoutItemRow item={item} />
          </div>
        ))}
      </CardContent>
    </Card>
  )
}
