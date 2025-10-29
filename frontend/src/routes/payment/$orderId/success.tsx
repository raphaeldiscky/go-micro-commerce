import { PaymentStatusBadge } from '@/components/payment/PaymentStatusBadge'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { PATH } from '@/constants/routes'
import { usePaymentByOrderId } from '@/hooks/payment'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { AlertCircle, CheckCircle, Loader2, ShoppingBag } from 'lucide-react'
import { useEffect, useState } from 'react'
import type { Payment } from '../../../types/__generated__/graphql'

export const Route = createFileRoute('/payment/$orderId/success')({
  component: RouteComponent,
})

function RouteComponent() {
  const { orderId } = Route.useParams()
  const navigate = useNavigate()
  const [payment, setPayment] = useState<Payment | null>(null)

  const { data, isLoading, error } = usePaymentByOrderId(orderId, {
    enabled: true,
    refetchInterval: () => {
      const status = payment?.status
      // Poll every 2 seconds if payment is still processing
      return status === 'PROCESSING' || status === 'PENDING' ? 2000 : false
    },
  })

  useEffect(() => {
    if (data) {
      setPayment(data)
    }
  }, [data])

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50/40 flex items-center justify-center p-4">
        <Card className="max-w-md w-full">
          <CardContent className="pt-6">
            <div className="flex flex-col items-center gap-4 text-center">
              <Loader2 className="h-12 w-12 animate-spin text-primary" />
              <div className="space-y-2">
                <h3 className="text-lg font-semibold">Verifying Payment</h3>
                <p className="text-sm text-muted-foreground">
                  Please wait while we confirm your payment...
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

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
          <CardContent className="space-y-4">
            <p className="text-muted-foreground">
              We couldn't find this payment. It may have expired or been
              cancelled.
            </p>
            <Button
              onClick={() => navigate({ to: PATH.products.root })}
              className="w-full"
            >
              <ShoppingBag className="mr-2 h-4 w-4" />
              Return to Shop
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  const isFailed = payment.status === 'FAILED'
  const isTimeout = payment.status === 'TIMEOUT'
  const isProcessing =
    payment.status === 'PROCESSING' || payment.status === 'PENDING'

  if (isFailed || isTimeout) {
    return (
      <div className="min-h-screen bg-gray-50/40 flex items-center justify-center p-4">
        <Card className="max-w-md w-full">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-red-600">
              <AlertCircle className="h-5 w-5" />
              Payment {isTimeout ? 'Timeout' : 'Failed'}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>
                Payment {isTimeout ? 'Timeout' : 'Failed'}
              </AlertTitle>
              <AlertDescription>
                {isTimeout
                  ? 'Your payment window has expired. Please create a new order to continue.'
                  : 'Your payment could not be processed. Please try again or contact support.'}
              </AlertDescription>
            </Alert>

            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Payment ID</span>
                <span className="font-mono text-xs">{payment.id}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Order ID</span>
                <span className="font-mono text-xs">{payment.orderId}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Status</span>
                <PaymentStatusBadge status={payment.status} />
              </div>
            </div>

            <div className="flex gap-2">
              <Button
                onClick={() =>
                  navigate({ to: PATH.orders.detail(payment.orderId) })
                }
                variant="outline"
                className="flex-1"
              >
                View Order
              </Button>
              <Button
                onClick={() => navigate({ to: PATH.products.root })}
                className="flex-1"
              >
                <ShoppingBag className="mr-2 h-4 w-4" />
                Shop Again
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (isProcessing) {
    return (
      <div className="min-h-screen bg-gray-50/40 flex items-center justify-center p-4">
        <Card className="max-w-md w-full">
          <CardContent className="pt-6">
            <div className="flex flex-col items-center gap-4 text-center">
              <Loader2 className="h-12 w-12 animate-spin text-primary" />
              <div className="space-y-2">
                <h3 className="text-lg font-semibold">Processing Payment</h3>
                <p className="text-sm text-muted-foreground">
                  Your payment is being processed. This may take a few
                  moments...
                </p>
              </div>

              <Alert className="mt-4">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription className="text-xs">
                  Please don't close this page. We'll update the status
                  automatically.
                </AlertDescription>
              </Alert>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50/40">
      {/* Header */}
      <div className="border-b bg-white">
        <div className="mx-auto max-w-2xl px-4 sm:px-6 lg:px-8 py-6 text-center">
          <div className="mx-auto h-16 w-16 rounded-full bg-green-100 flex items-center justify-center mb-4">
            <CheckCircle className="h-8 w-8 text-green-600" />
          </div>
          <h1 className="text-3xl font-bold tracking-tight">
            Payment Successful!
          </h1>
          <p className="text-sm text-muted-foreground mt-2">
            Thank you for your purchase.
          </p>
        </div>
      </div>

      {/* Main Content */}
      <div className="mx-auto max-w-2xl px-4 sm:px-6 lg:px-8 py-8 space-y-6">
        {/* Success Alert */}
        <Alert className="border-green-200 bg-green-50">
          <CheckCircle className="h-4 w-4 text-green-600" />
          <AlertTitle className="text-green-600">
            Payment Completed Successfully
          </AlertTitle>
          <AlertDescription className="text-green-600">
            Your payment has been processed and confirmed. You'll receive an
            email confirmation shortly.
          </AlertDescription>
        </Alert>

        {/* Payment Summary */}
        <Card>
          <CardHeader>
            <CardTitle>Payment Summary</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex justify-between items-center py-2 border-b">
              <span className="text-sm text-muted-foreground">Amount Paid</span>
              <span className="text-2xl font-bold">
                {payment.currency.toUpperCase()} {payment.amount}
              </span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Payment ID</span>
              <span className="font-mono text-xs">{payment.id}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Order ID</span>
              <span className="font-mono text-xs">{payment.orderId}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Payment Method</span>
              <span className="capitalize">{payment.paymentGateway}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Status</span>
              <PaymentStatusBadge status={payment.status} />
            </div>
            {payment.completedAt && (
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Completed At</span>
                <span>{new Date(payment.completedAt).toLocaleString()}</span>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Action Buttons */}
        <div className="flex flex-col sm:flex-row gap-3">
          <Button
            onClick={() =>
              navigate({ to: PATH.orders.detail(payment.orderId) })
            }
            size="lg"
            className="flex-1"
          >
            View Order Details
          </Button>
          <Button
            onClick={() => navigate({ to: PATH.products.root })}
            variant="outline"
            size="lg"
            className="flex-1"
          >
            <ShoppingBag className="mr-2 h-4 w-4" />
            Continue Shopping
          </Button>
        </div>

        {/* Additional Info */}
        <Card className="bg-blue-50/50 border-blue-200">
          <CardContent className="pt-6">
            <div className="text-center space-y-2">
              <h4 className="text-sm font-medium">What's Next?</h4>
              <p className="text-xs text-muted-foreground">
                You can track your order status and view order details from your
                orders page. We'll send you updates via email as your order
                progresses.
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
