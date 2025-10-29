import { PaymentCountdownTimer } from '@/components/payment/PaymentCountdownTimer'
import { PaymentStatusBadge } from '@/components/payment/PaymentStatusBadge'
import { StripePaymentElement } from '@/components/payment/StripePaymentElement'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { PATH } from '@/constants/routes'
import { env } from '@/env'
import { usePaymentByOrderId } from '@/hooks/payment'
import { Elements } from '@stripe/react-stripe-js'
import { loadStripe } from '@stripe/stripe-js'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { AlertCircle, CheckCircle, ShieldCheck } from 'lucide-react'
import { toast } from 'sonner'

export const Route = createFileRoute('/payment/$orderId/')({
  component: RouteComponent,
})

const stripePromise = loadStripe(env.VITE_STRIPE_PUBLISHABLE_KEY)

function RouteComponent() {
  const { orderId } = Route.useParams()
  const navigate = useNavigate()

  const {
    data: payment,
    isLoading,
    error,
    refetch,
  } = usePaymentByOrderId(orderId, {
    enabled: true,
  })

  const handleExpired = () => {
    toast.error('Payment window has expired')
    refetch()
  }

  const handlePaymentError = (errorMsg: string) => {
    toast.error(errorMsg)
  }

  const handlePaymentSuccess = () => {
    toast.success('Payment successful!')
    navigate({ to: PATH.payment.success(orderId) })
  }

  // Loading state
  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50/40 p-4 sm:p-6 lg:p-8">
        <div className="mx-auto max-w-4xl space-y-6">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-96 w-full" />
          <Skeleton className="h-32 w-full" />
        </div>
      </div>
    )
  }

  // Error state
  if (error || !payment) {
    return (
      <div className="min-h-screen bg-gray-50/40 flex items-center justify-center p-4">
        <Card className="max-w-md w-full">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-red-600">
              <AlertCircle className="h-5 w-5" />
              Payment Not Found
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground mb-4">
              We couldn't find this payment. It may have expired or been
              cancelled.
            </p>
            <Button onClick={() => navigate({ to: PATH.products.root })}>
              Return to Shop
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  const isExpired = payment.expiresAt
    ? new Date(payment.expiresAt).getTime() < Date.now()
    : false
  const isCompleted = payment.status === 'COMPLETED'
  const isFailed = payment.status === 'FAILED'
  const isTimeout = payment.status === 'TIMEOUT'
  const canPay =
    !isExpired &&
    !isCompleted &&
    !isFailed &&
    !isTimeout &&
    payment.clientSecret

  return (
    <div className="min-h-screen bg-gray-50/40">
      {/* Header */}
      <div className="border-b bg-white">
        <div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold tracking-tight">
                {isCompleted ? 'Payment Successful' : 'Complete Payment'}
              </h1>
              <p className="text-sm text-muted-foreground">
                Order ID: {payment.orderId}
              </p>
            </div>
            <PaymentStatusBadge status={payment.status} />
          </div>
        </div>
      </div>
      <Button onClick={handlePaymentSuccess}>red</Button>

      {/* Main Content */}
      <div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8 py-8 space-y-6">
        {/* Success Alert */}
        {isCompleted && (
          <Alert className="border-green-200 bg-green-50">
            <CheckCircle className="h-4 w-4 text-green-600" />
            <AlertTitle className="text-green-600">
              Payment Completed!
            </AlertTitle>
            <AlertDescription className="text-green-600">
              Your payment has been processed successfully.
            </AlertDescription>
          </Alert>
        )}

        {/* Expired Alert */}
        {(isExpired || isTimeout) && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Payment Window Expired</AlertTitle>
            <AlertDescription>
              Your payment window has closed. Please create a new order to
              continue.
              <Button
                variant="outline"
                size="sm"
                className="mt-2"
                onClick={() => navigate({ to: PATH.products.root })}
              >
                Return to Shop
              </Button>
            </AlertDescription>
          </Alert>
        )}

        {/* Failed Alert */}
        {isFailed && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Payment Failed</AlertTitle>
            <AlertDescription>
              Your payment could not be processed. Please try again or contact
              support.
            </AlertDescription>
          </Alert>
        )}

        {/* Countdown Timer */}
        {payment.expiresAt && !isCompleted && !isTimeout && (
          <PaymentCountdownTimer
            expiresAt={payment.expiresAt}
            onExpired={handleExpired}
          />
        )}

        {/* Payment Form */}
        {canPay && payment.clientSecret && (
          <Elements
            stripe={stripePromise}
            options={{
              clientSecret: payment.clientSecret,
              appearance: {
                theme: 'stripe',
                variables: {
                  colorPrimary: '#0F172A',
                  colorBackground: '#ffffff',
                  colorText: '#0F172A',
                  colorDanger: '#dc2626',
                  fontFamily: 'system-ui, sans-serif',
                  spacingUnit: '4px',
                  borderRadius: '8px',
                },
              },
            }}
          >
            <StripePaymentElement
              orderId={payment.orderId}
              amount={payment.amount}
              currency={payment.currency}
              onSuccess={handlePaymentSuccess}
              onError={handlePaymentError}
            />
          </Elements>
        )}

        {/* Payment Info */}
        {canPay && (
          <Card>
            <CardContent className="pt-6">
              <div className="text-center space-y-2">
                <div className="mx-auto h-8 w-8 rounded-full bg-green-100 flex items-center justify-center">
                  <ShieldCheck className="h-4 w-4 text-green-600" />
                </div>
                <h4 className="text-sm font-medium">Secure Payment</h4>
                <p className="text-xs text-muted-foreground">
                  Your payment is secured by Stripe. We never see your card
                  details.
                </p>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Payment Details Summary */}
        <Card>
          <CardHeader>
            <CardTitle>Payment Summary</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Payment ID</span>
              <span className="font-mono text-xs">{payment.id}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Gateway</span>
              <span className="capitalize">{payment.paymentGateway}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Amount</span>
              <span className="font-semibold">
                {payment.currency.toUpperCase()} {payment.amount}
              </span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Status</span>
              <PaymentStatusBadge status={payment.status} />
            </div>
            {payment.createdAt && (
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Created</span>
                <span>{new Date(payment.createdAt).toLocaleString()}</span>
              </div>
            )}
            {payment.completedAt && (
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Completed</span>
                <span>{new Date(payment.completedAt).toLocaleString()}</span>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
