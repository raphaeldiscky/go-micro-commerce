import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useStripePaymentConfirmation } from '@/hooks/payment'
import { PaymentElement, useElements, useStripe } from '@stripe/react-stripe-js'
import { Loader2 } from 'lucide-react'
import { useState } from 'react'
import { PATH } from '../../constants'

interface StripePaymentElementProps {
  orderId: string
  amount: string
  currency: string
  onSuccess?: () => void
  onError?: (error: string) => void
}

/**
 * Stripe Payment Element component
 * Follows Stripe best practices using Payment Element (not Card Element)
 * Handles payment method collection and confirmation
 */
export function StripePaymentElement({
  orderId,
  amount,
  currency,
  onSuccess,
  onError,
}: StripePaymentElementProps) {
  const stripe = useStripe()
  const elements = useElements()
  const { confirmPayment, isProcessing, error } = useStripePaymentConfirmation()
  const [isReady, setIsReady] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!stripe || !elements) {
      return
    }

    const domain = window.location.origin

    const result = await confirmPayment({
      returnUrl: domain + PATH.payment.success(orderId),
      onSuccess,
      onError,
    })

    if (!result.success && result.error) {
      console.error('Payment confirmation failed:', result.error)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Payment Details</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="rounded-md border p-4">
            <PaymentElement
              onReady={() => setIsReady(true)}
              options={{
                layout: 'tabs',
              }}
            />
          </div>

          {!isReady && (
            <div className="flex items-center justify-center py-8 text-muted-foreground">
              <Loader2 className="mr-2 h-5 w-5 animate-spin" />
              Loading payment form...
            </div>
          )}

          {error && (
            <div className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-600">
              {error}
            </div>
          )}

          <div className="flex items-center justify-between rounded-md bg-muted p-4">
            <div>
              <p className="text-sm text-muted-foreground">Total Amount</p>
              <p className="text-2xl font-bold">
                {currency.toUpperCase()} {amount}
              </p>
            </div>
          </div>

          <Button
            type="submit"
            className="w-full"
            disabled={!stripe || !isReady || isProcessing}
            size="lg"
          >
            {isProcessing ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Processing...
              </>
            ) : (
              `Pay ${currency.toUpperCase()} ${amount}`
            )}
          </Button>

          <p className="text-center text-xs text-muted-foreground">
            Your payment is secured by Stripe. We never see your card details.
          </p>
        </form>
      </CardContent>
    </Card>
  )
}
