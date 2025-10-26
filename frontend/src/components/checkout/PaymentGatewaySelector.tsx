import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { mockPaymentGateways } from '@/mocks/shipping'
import { useCheckoutSessionStore } from '@/store/checkoutSessionStore'
import type { PaymentGatewayUI } from '@/types/cart'
import { Building2 } from 'lucide-react'
import { useEffect } from 'react'
import { BsPaypal, BsStripe } from 'react-icons/bs'

const getGatewayIcon = (type: PaymentGatewayUI['id']) => {
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
    selectedPaymentGatewayData,
    selectedPaymentGateway,
    setPaymentGateway,
  } = useCheckoutSessionStore()

  // Reconstruct selectedPaymentGatewayData from selectedPaymentGateway after fetch
  useEffect(() => {
    if (selectedPaymentGateway && !selectedPaymentGatewayData) {
      const gateway = mockPaymentGateways.find(
        (gw) => gw.id === selectedPaymentGateway,
      )
      if (gateway) {
        // Just update the UI object in store (not calling backend)
        useCheckoutSessionStore.setState({
          selectedPaymentGatewayData: gateway,
        })
      }
    }
  }, [selectedPaymentGateway, selectedPaymentGatewayData])

  const handleGatewayChange = (gatewayId: string) => {
    const gateway = mockPaymentGateways.find((gw) => gw.id === gatewayId)
    if (gateway) {
      setPaymentGateway(gatewayId, gateway)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Building2 className="h-5 w-5" />
          Payment Gateway
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <RadioGroup
          value={selectedPaymentGatewayData?.id || selectedPaymentGateway || ''}
          onValueChange={handleGatewayChange}
        >
          {mockPaymentGateways.map((gateway: PaymentGatewayUI) => {
            const Icon = getGatewayIcon(gateway.id)
            return (
              <div key={gateway.id} className="space-y-1">
                <div className="flex items-center space-x-3">
                  <RadioGroupItem id={gateway.id} value={gateway.id} />
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

        {!selectedPaymentGatewayData && (
          <p className="text-sm text-muted-foreground">
            Please select a payment gateway to continue
          </p>
        )}
      </CardContent>
    </Card>
  )
}
