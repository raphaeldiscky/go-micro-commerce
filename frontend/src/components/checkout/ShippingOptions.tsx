import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { mockShippingOptions } from '@/mocks/shipping'
import { useCartData } from '@/store/cartStore'
import { useCheckoutSessionStore } from '@/store/checkoutSessionStore'
import type { ShippingOptionUI } from '@/types/cart'
import { Clock, Truck } from 'lucide-react'
import { useMemo } from 'react'

export function ShippingOptions() {
  const { selectedAddress, selectedShippingOption, setShippingMethod } =
    useCheckoutSessionStore()

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

  // Calculate subtotal from selected cart items
  const subtotal = selectedItems.reduce(
    (total, item) => total + (item.product?.price ?? 0) * item.quantity,
    0,
  )
  const isDisabled = !selectedAddress

  const handleShippingChange = (optionId: string) => {
    const option = mockShippingOptions.find((opt) => opt.id === optionId)
    if (option) {
      setShippingMethod(optionId, option)
    }
  }

  // Filter shipping options based on cart total
  const availableShippingOptions = mockShippingOptions.filter((option) => {
    if (option.price === 0 && subtotal >= 100) {
      return true // Free shipping available for orders over $100
    }
    return option.price > 0
  })

  return (
    <Card className={isDisabled ? 'opacity-60' : ''}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Truck className="h-5 w-5" />
          Shipping Method
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {isDisabled && (
          <p className="text-sm text-muted-foreground mb-4">
            Please select a delivery address first
          </p>
        )}
        <RadioGroup
          value={selectedShippingOption?.id || ''}
          onValueChange={handleShippingChange}
          disabled={isDisabled}
        >
          {availableShippingOptions.map((option: ShippingOptionUI) => (
            <div key={option.id} className="space-y-2">
              <div className="flex items-start space-x-3">
                <RadioGroupItem
                  id={option.id}
                  value={option.id}
                  className="mt-1"
                  disabled={isDisabled}
                />
                <div className="flex-1 space-y-1">
                  <Label
                    htmlFor={option.id}
                    className="flex items-center justify-between font-medium cursor-pointer"
                  >
                    <span>{option.name}</span>
                    <span className="font-normal">
                      {option.price === 0 ? (
                        <span className="text-green-600 font-medium">FREE</span>
                      ) : (
                        `+${option.price.toFixed(2)}`
                      )}
                    </span>
                  </Label>
                  {option.description && (
                    <p className="text-sm text-muted-foreground">
                      {option.description}
                    </p>
                  )}
                  <div className="flex items-center gap-1 text-xs text-muted-foreground">
                    <Clock className="h-3 w-3" />
                    <span>
                      {option.estimatedDays.min === option.estimatedDays.max
                        ? `${option.estimatedDays.min} business day`
                        : `${option.estimatedDays.min}-${option.estimatedDays.max} business days`}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </RadioGroup>

        {!isDisabled && !selectedShippingOption && (
          <p className="text-sm text-muted-foreground">
            Please select a shipping method to continue
          </p>
        )}
      </CardContent>
    </Card>
  )
}
