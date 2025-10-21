import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { mockPaymentGateways } from '@/data/mockData'
import { useCartStore } from '@/store/cartStore'
import type { PaymentGateway } from '@/types/cart'
import { Building2, CreditCard } from 'lucide-react'
import { useEffect } from 'react'

const getGatewayIcon = (type: PaymentGateway['type']) => {
  switch (type) {
    case 'stripe':
      return CreditCard
    case 'paypal':
      return CreditCard
    default:
      return Building2
  }
}

export function PaymentGatewaySelector() {
  const {
    selectedAddress,
    selectedShippingOption,
    selectedPaymentMethod,
    selectedPaymentGateway,
    setPaymentGateway,
  } = useCartStore()

  const isDisabled =
    !selectedAddress || !selectedShippingOption || !selectedPaymentMethod

  // Filter available gateways based on payment method
  const availableGateways = selectedPaymentMethod?.supportedGateways
    ? selectedPaymentMethod.supportedGateways
    : mockPaymentGateways

  // Auto-select first available gateway when payment method is selected and no gateway is selected
  useEffect(() => {
    if (selectedPaymentMethod && !selectedPaymentGateway && availableGateways.length > 0) {
      setPaymentGateway(availableGateways[0])
    }
  }, [selectedPaymentMethod, selectedPaymentGateway, availableGateways, setPaymentGateway])

  const handleGatewayChange = (gatewayId: string) => {
    const gateway = availableGateways.find((g) => g.id === gatewayId)
    if (gateway) {
      setPaymentGateway(gateway)
    }
  }

  return (
    <Card className={isDisabled ? 'opacity-60' : ''}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Building2 className="h-5 w-5" />
          Payment Gateway
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {isDisabled && (
          <p className="text-sm text-muted-foreground mb-4">
            Please complete previous steps first
          </p>
        )}
        <RadioGroup
          value={selectedPaymentGateway?.id || ''}
          onValueChange={handleGatewayChange}
          disabled={isDisabled}
        >
          {availableGateways.map((gateway: PaymentGateway) => {
            const Icon = getGatewayIcon(gateway.type)
            return (
              <div key={gateway.id} className="space-y-2">
                <div className="flex items-start space-x-3">
                  <RadioGroupItem
                    id={gateway.id}
                    value={gateway.id}
                    className="mt-1"
                    disabled={isDisabled}
                  />
                  <div className="flex-1 space-y-1">
                    <Label
                      htmlFor={gateway.id}
                      className="flex items-center justify-between font-medium cursor-pointer"
                    >
                      <div className="flex items-center gap-2">
                        <Icon className="h-4 w-4" />
                        <span>{gateway.name}</span>
                      </div>
                    </Label>
                  </div>
                </div>
              </div>
            )
          })}
        </RadioGroup>

        {!isDisabled && !selectedPaymentGateway && (
          <p className="text-sm text-muted-foreground">
            Please select a payment gateway to continue
          </p>
        )}
      </CardContent>
    </Card>
  )
}
