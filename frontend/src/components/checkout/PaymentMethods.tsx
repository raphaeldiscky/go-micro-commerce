import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { mockPaymentMethods } from '@/mocks/shipping'
import { useCartStore } from '@/store/cartStore'
import type { PaymentMethod } from '@/types/cart'
import { Building, CreditCard, Package, Smartphone, Wallet } from 'lucide-react'

const getPaymentIcon = (type: PaymentMethod['type']) => {
  switch (type) {
    case 'card':
      return CreditCard
    case 'ewallet':
      return Smartphone
    case 'bank_transfer':
      return Building
    case 'cod':
      return Package
    default:
      return Wallet
  }
}

export function PaymentMethods() {
  const {
    selectedAddress,
    selectedShippingOption,
    selectedPaymentMethod,
    setPaymentMethod,
  } = useCartStore()

  const isDisabled = !selectedAddress || !selectedShippingOption

  const handlePaymentChange = (methodId: string) => {
    const method = mockPaymentMethods.find((m) => m.id === methodId)
    if (method) {
      setPaymentMethod(method)
    }
  }

  return (
    <Card className={isDisabled ? 'opacity-60' : ''}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <CreditCard className="h-5 w-5" />
          Payment Method
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {isDisabled && (
          <p className="text-sm text-muted-foreground mb-4">
            Please select an address and shipping method first
          </p>
        )}
        <RadioGroup
          value={selectedPaymentMethod?.id || ''}
          onValueChange={handlePaymentChange}
          disabled={isDisabled}
        >
          {mockPaymentMethods.map((method: PaymentMethod) => {
            const Icon = getPaymentIcon(method.type)
            return (
              <div key={method.id} className="space-y-1">
                <div className="flex items-center space-x-3">
                  <RadioGroupItem id={method.id} value={method.id} />
                  <Label
                    htmlFor={method.id}
                    className="font-medium cursor-pointer flex items-center gap-2"
                  >
                    <Icon className="h-4 w-4" />
                    <span>{method.name}</span>
                  </Label>
                </div>
                {method.description && (
                  <p className="text-sm text-muted-foreground ml-6 pl-7">
                    {method.description}
                  </p>
                )}
              </div>
            )
          })}
        </RadioGroup>

        {!selectedPaymentMethod && (
          <p className="text-sm text-muted-foreground">
            Please select a payment method to continue
          </p>
        )}
      </CardContent>
    </Card>
  )
}
