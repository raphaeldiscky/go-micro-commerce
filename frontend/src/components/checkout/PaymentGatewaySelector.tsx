import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { mockPaymentGateways } from '@/mocks/shipping'
import { useCartStore } from '@/store/cartStore'
import type { PaymentGateway } from '@/types/cart'
import { Building2 } from 'lucide-react'
import { useEffect } from 'react'
import { BsPaypal, BsStripe } from 'react-icons/bs'

const getGatewayIcon = (type: PaymentGateway['type']) => {
  switch (type) {
    case 'stripe':
      return BsStripe
    case 'paypal':
      return BsPaypal
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
    if (
      selectedPaymentMethod &&
      !selectedPaymentGateway &&
      availableGateways.length > 0
    ) {
      setPaymentGateway(availableGateways[0])
    }
  }, [
    selectedPaymentMethod,
    selectedPaymentGateway,
    availableGateways,
    setPaymentGateway,
  ])

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
              <div key={gateway.id} className="space-y-1">
                <div className="flex items-center space-x-3">
                  <RadioGroupItem
                    id={gateway.id}
                    value={gateway.id}
                    disabled={isDisabled}
                  />
                  <Label
                    htmlFor={gateway.id}
                    className="font-medium cursor-pointer flex items-center gap-2"
                  >
                    <Icon className="h-4.5 w-4.5" />
                    <span>{gateway.name}</span>
                  </Label>
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
