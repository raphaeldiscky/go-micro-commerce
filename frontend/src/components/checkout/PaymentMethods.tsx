import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { mockPaymentMethods } from '@/data/mockData'
import { useCartStore } from '@/store/cartStore'
import type { PaymentMethod } from '@/types/cart'
import { CreditCard, Smartphone, Building, Wallet, Package } from 'lucide-react'

const getPaymentIcon = (type: PaymentMethod['type']) => {
  switch (type) {
    case 'credit_card':
    case 'debit_card':
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
  const { selectedPaymentMethod, setPaymentMethod } = useCartStore()

  const handlePaymentChange = (methodId: string) => {
    const method = mockPaymentMethods.find((m) => m.id === methodId)
    if (method) {
      setPaymentMethod(method)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <CreditCard className="h-5 w-5" />
          Payment Method
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <RadioGroup
          value={selectedPaymentMethod?.id || ''}
          onValueChange={handlePaymentChange}
        >
          {mockPaymentMethods.map((method: PaymentMethod) => {
            const Icon = getPaymentIcon(method.type)
            return (
              <div key={method.id} className="space-y-2">
                <div className="flex items-start space-x-3">
                  <RadioGroupItem
                    id={method.id}
                    value={method.id}
                    className="mt-1"
                  />
                  <div className="flex-1 space-y-1">
                    <Label
                      htmlFor={method.id}
                      className="flex items-center justify-between font-medium cursor-pointer"
                    >
                      <div className="flex items-center gap-2">
                        <Icon className="h-4 w-4" />
                        <span>{method.name}</span>
                      </div>
                    </Label>
                    {method.description && (
                      <p className="text-sm text-muted-foreground ml-6">
                        {method.description}
                      </p>
                    )}
                  </div>
                </div>
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
