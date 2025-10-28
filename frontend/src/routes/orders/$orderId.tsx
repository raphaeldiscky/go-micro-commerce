import { PaymentStatusBadge } from '@/components/payment/PaymentStatusBadge'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { PATH } from '@/constants/routes'
import {
  usePaymentByOrderId,
  usePaymentStatusSubscription,
} from '@/hooks/payment'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import {
  AlertCircle,
  CheckCircle,
  Package,
  Truck,
  MapPin,
  CreditCard,
  ArrowLeft,
} from 'lucide-react'
import { useEffect } from 'react'
import { toast } from 'sonner'

export const Route = createFileRoute(PATH.orders.$orderId)({
  component: RouteComponent,
})

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
    refetchInterval: false, // Don't poll - use SSE instead
  })

  // Subscribe to real-time order status updates via SSE
  usePaymentStatusSubscription(orderId, {
    enabled: !!payment,
    onPaymentSuccess: () => {
      toast.success('Payment completed successfully!')
      refetch()
    },
    onPaymentFailed: (errorMsg) => {
      toast.error(errorMsg || 'Payment failed. Please try again.')
      refetch()
    },
    onPaymentTimeout: () => {
      toast.error('Payment window expired. Please create a new order.')
      refetch()
    },
  })

  // Show success message when payment is completed
  useEffect(() => {
    if (payment?.status === 'COMPLETED') {
      toast.success('Order confirmed! Your payment has been processed.')
    }
  }, [payment?.status])

  const handleBackToShop = () => {
    navigate({ to: PATH.products.root })
  }

  const handleTrackOrder = () => {
    // TODO: Navigate to order tracking page when implemented
    toast.info('Order tracking feature coming soon!')
  }

  // Loading state
  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50/40 p-4 sm:p-6 lg:p-8">
        <div className="mx-auto max-w-4xl space-y-6">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-32 w-full" />
          <Skeleton className="h-48 w-full" />
          <Skeleton className="h-48 w-full" />
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
              Order Not Found
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground mb-4">
              We couldn't find your order. It may have been cancelled or there
              was an error.
            </p>
            <Button onClick={() => navigate({ to: PATH.products.root })}>
              Return to Shop
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  const isCompleted = payment.status === 'COMPLETED'
  const isFailed = payment.status === 'FAILED'
  const isPending =
    payment.status === 'PENDING' || payment.status === 'PROCESSING'
  const isExpired = payment.expiresAt
    ? new Date(payment.expiresAt).getTime() < Date.now()
    : false
  const canRetryPayment = isPending && !isExpired && payment.clientSecret

  return (
    <div className="min-h-screen bg-gray-50/40">
      {/* Header */}
      <div className="border-b bg-white">
        <div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between">
            <div>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => navigate({ to: PATH.products.root })}
                className="text-muted-foreground hover:text-foreground mb-4"
              >
                <ArrowLeft className="h-4 w-4 mr-2" />
                Back to Shop
              </Button>
              <h1 className="text-2xl font-bold tracking-tight">
                Order Confirmation
              </h1>
              <p className="text-sm text-muted-foreground">
                Order ID: {payment.orderId}
              </p>
            </div>
            <PaymentStatusBadge status={payment.status} />
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8 py-8 space-y-6">
        {/* Order Status Alert */}
        {isCompleted && (
          <Alert className="border-green-200 bg-green-50">
            <CheckCircle className="h-4 w-4 text-green-600" />
            <AlertTitle className="text-green-600">Order Confirmed!</AlertTitle>
            <AlertDescription className="text-green-600">
              Your order has been successfully placed and payment has been
              processed. You will receive a confirmation email shortly.
            </AlertDescription>
          </Alert>
        )}

        {isPending && (
          <Alert>
            <Package className="h-4 w-4" />
            <AlertTitle>Order Processing</AlertTitle>
            <AlertDescription>
              Your order is being processed. We'll notify you once it's ready
              for shipment.
            </AlertDescription>
          </Alert>
        )}

        {isFailed && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Order Failed</AlertTitle>
            <AlertDescription>
              There was an issue with your order. Please contact support or try
              placing the order again.
            </AlertDescription>
          </Alert>
        )}

        <div className="grid gap-6 lg:grid-cols-2">
          {/* Order Details */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Package className="h-5 w-5" />
                Order Details
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Order ID</span>
                  <span className="font-mono text-xs">{payment.orderId}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Status</span>
                  <PaymentStatusBadge status={payment.status} />
                </div>
                {payment.createdAt && (
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Order Date</span>
                    <span>
                      {new Date(payment.createdAt).toLocaleDateString()}
                    </span>
                  </div>
                )}
                {payment.completedAt && (
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">
                      Payment Completed
                    </span>
                    <span>
                      {new Date(payment.completedAt).toLocaleString()}
                    </span>
                  </div>
                )}
              </div>

              <hr />

              {/* Order Summary */}
              <div className="space-y-2">
                <h4 className="font-medium">Order Summary</h4>
                <div className="space-y-1 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Subtotal</span>
                    <span>
                      {payment.currency.toUpperCase()} {payment.amount}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Shipping</span>
                    <span>Calculated at checkout</span>
                  </div>
                  <div className="flex justify-between font-semibold">
                    <span>Total</span>
                    <span>
                      {payment.currency.toUpperCase()} {payment.amount}
                    </span>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Payment Information */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <CreditCard className="h-5 w-5" />
                Payment Information
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Payment Method</span>
                  <span className="capitalize">{payment.paymentGateway}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Payment ID</span>
                  <span className="font-mono text-xs">{payment.id}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Amount Paid</span>
                  <span className="font-semibold">
                    {payment.currency.toUpperCase()} {payment.amount}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Payment Status</span>
                  <PaymentStatusBadge status={payment.status} />
                </div>
              </div>

              <hr />

              <div className="text-center space-y-2">
                <div className="mx-auto h-8 w-8 rounded-full bg-green-100 flex items-center justify-center">
                  <CheckCircle className="h-4 w-4 text-green-600" />
                </div>
                <p className="text-xs text-muted-foreground">
                  Your payment was processed securely via{' '}
                  {payment.paymentGateway}.
                </p>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Shipping Information */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Truck className="h-5 w-5" />
              Shipping Information
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <h4 className="font-medium flex items-center gap-2">
                  <MapPin className="h-4 w-4" />
                  Delivery Address
                </h4>
                <div className="text-sm text-muted-foreground bg-muted p-3 rounded">
                  <p>
                    Address details will be available once order processing
                    begins.
                  </p>
                </div>
              </div>
              <div className="space-y-2">
                <h4 className="font-medium">Estimated Delivery</h4>
                <div className="text-sm text-muted-foreground bg-muted p-3 rounded">
                  <p>
                    Delivery information will be available once order processing
                    begins.
                  </p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Action Buttons */}
        <Card>
          <CardContent className="pt-6">
            <div className="flex flex-col gap-4">
              {/* Payment Retry Button */}
              {canRetryPayment && (
                <Button
                  onClick={() => navigate({ to: PATH.payment.detail(orderId) })}
                  className="w-full"
                  size="lg"
                >
                  Complete Payment
                </Button>
              )}

              {/* Default Action Buttons */}
              <div className="flex flex-col sm:flex-row gap-4">
                <Button
                  onClick={handleBackToShop}
                  variant="outline"
                  className="flex-1"
                >
                  Continue Shopping
                </Button>
                <Button onClick={handleTrackOrder} className="flex-1">
                  Track Order
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
