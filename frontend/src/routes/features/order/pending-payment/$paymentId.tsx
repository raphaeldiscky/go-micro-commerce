import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { PATH_FEATURES } from '@/constants/routes'
import { formatCurrency } from '@/data/mockData'
import { useCreateCheckoutSession } from '@/hooks/order'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import {
  AlertCircle,
  Clock,
  CreditCard,
  MapPin,
  Package,
  ShieldCheck,
} from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

export const Route = createFileRoute(
  '/features/order/pending-payment/$paymentId',
)({
  component: RouteComponent,
})

// Mock payment details hook (will be replaced with real hook in Phase 10)
function useMockPaymentDetails(paymentId: string) {
  return {
    data: {
      paymentId,
      orderId: 'order-123',
      amount: 359.97,
      currency: 'USD',
      paymentStatus: 'pending' as const,
      paymentMethod: 'card',
      paymentGateway: 'stripe',
      paymentDeadline: new Date(Date.now() + 23 * 60 * 60 * 1000).toISOString(),
      order: {
        orderId: 'order-123',
        items: [
          {
            productId: 'prod-1',
            productName: 'Wireless Bluetooth Headphones',
            quantity: 1,
            price: 299.99,
          },
          {
            productId: 'prod-2',
            productName: 'Smart Watch Pro',
            quantity: 1,
            price: 49.99,
          },
        ],
        subtotal: 349.98,
        shippingCost: 9.99,
        total: 359.97,
        shippingAddress: {
          receiverName: 'John Doe',
          addressLine1: '123 Main Street',
          addressLine2: 'Apt 4B',
          city: 'San Francisco',
          state: 'CA',
          postalCode: '94103',
          countryCode: 'US',
        },
      },
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    },
    isLoading: false,
    error: null,
  }
}

// Mock countdown timer (will be replaced with real component in Phase 9)
function MockCountdownTimer({ deadline }: { deadline: string }) {
  const timeRemaining = Math.max(0, new Date(deadline).getTime() - Date.now())
  const hours = Math.floor(timeRemaining / (1000 * 60 * 60))
  const minutes = Math.floor((timeRemaining % (1000 * 60 * 60)) / (1000 * 60))
  const seconds = Math.floor((timeRemaining % (1000 * 60)) / 1000)

  const isExpired = timeRemaining === 0
  const isUrgent = hours < 1
  const isWarning = hours < 2 && !isUrgent

  const colorClass = isExpired
    ? 'text-red-600'
    : isUrgent
      ? 'text-red-600'
      : isWarning
        ? 'text-yellow-600'
        : 'text-green-600'

  if (isExpired) {
    return <span className={colorClass}>Expired</span>
  }

  return (
    <span className={colorClass}>
      {hours}h {minutes}m {seconds}s remaining
    </span>
  )
}

function RouteComponent() {
  const { paymentId } = Route.useParams()
  const navigate = useNavigate()
  const { data: payment, isLoading } = useMockPaymentDetails(paymentId)
  const createCheckoutSession = useCreateCheckoutSession()
  const [isCreatingSession, setIsCreatingSession] = useState(false)

  const handlePayNow = async () => {
    setIsCreatingSession(true)
    try {
      const result = await createCheckoutSession.mutateAsync({ paymentId })
      // Redirect to Stripe Checkout
      window.location.href = result.checkoutUrl
    } catch (err) {
      toast.error('Failed to create checkout session')
      setIsCreatingSession(false)
    }
  }

  const handleCreateNewOrder = () => {
    // TODO: Pre-fill cart with same items
    navigate({ to: PATH_FEATURES.products.root })
  }

  // Loading state
  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50/40 p-4 sm:p-6 lg:p-8">
        <div className="mx-auto max-w-6xl">
          <Skeleton className="h-8 w-64 mb-8" />
          <div className="grid gap-6 lg:grid-cols-3">
            <div className="lg:col-span-2 space-y-6">
              <Skeleton className="h-64 w-full" />
              <Skeleton className="h-48 w-full" />
            </div>
            <div className="space-y-6">
              <Skeleton className="h-48 w-full" />
              <Skeleton className="h-32 w-full" />
            </div>
          </div>
        </div>
      </div>
    )
  }

  const isExpired =
    new Date(payment.paymentDeadline).getTime() < Date.now() || false

  const canPay = !isExpired && !isCreatingSession

  return (
    <div className="min-h-screen bg-gray-50/40">
      {/* Header */}
      <div className="border-b bg-white">
        <div className="mx-auto max-w-6xl px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold tracking-tight">
                Payment Pending
              </h1>
              <p className="text-sm text-muted-foreground">
                Payment ID: {payment.paymentId}
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Badge variant={'default'} className="flex items-center gap-1">
                {isExpired && <AlertCircle className="h-3 w-3" />}
                {!isExpired && <Clock className="h-3 w-3" />}
                {payment.paymentStatus}
              </Badge>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="mx-auto max-w-6xl px-4 sm:px-6 lg:px-8 py-8">
        {/* Expired Alert */}
        {isExpired && (
          <Alert className="mb-6" variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Payment Window Expired</AlertTitle>
            <AlertDescription>
              Your payment window has closed. Please create a new order to
              continue.
            </AlertDescription>
          </Alert>
        )}

        <div className="grid gap-6 lg:grid-cols-3">
          {/* Left Column - Order Details */}
          <div className="lg:col-span-2 space-y-6">
            {/* Order Summary */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Package className="h-5 w-5" />
                  Order Summary
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-3">
                  {payment.order.items.map((item) => (
                    <div
                      key={item.productId}
                      className="flex items-center justify-between py-2"
                    >
                      <div className="flex-1">
                        <p className="font-medium">{item.productName}</p>
                        <p className="text-sm text-muted-foreground">
                          Quantity: {item.quantity}
                        </p>
                      </div>
                      <p className="font-medium">
                        {formatCurrency(item.price * item.quantity)}
                      </p>
                    </div>
                  ))}
                </div>

                <div className="border-t pt-4 space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Subtotal</span>
                    <span>{formatCurrency(payment.order.subtotal)}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Shipping</span>
                    <span>{formatCurrency(payment.order.shippingCost)}</span>
                  </div>
                  <div className="flex justify-between text-base font-bold pt-2 border-t">
                    <span>Total</span>
                    <span>{formatCurrency(payment.order.total)}</span>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Shipping Address */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <MapPin className="h-5 w-5" />
                  Shipping Address
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-sm space-y-1">
                  <p className="font-medium">
                    {payment.order.shippingAddress.receiverName}
                  </p>
                  <p className="text-muted-foreground">
                    {payment.order.shippingAddress.addressLine1}
                  </p>
                  {payment.order.shippingAddress.addressLine2 && (
                    <p className="text-muted-foreground">
                      {payment.order.shippingAddress.addressLine2}
                    </p>
                  )}
                  <p className="text-muted-foreground">
                    {payment.order.shippingAddress.city},{' '}
                    {payment.order.shippingAddress.state}{' '}
                    {payment.order.shippingAddress.postalCode}
                  </p>
                  <p className="text-muted-foreground">
                    {payment.order.shippingAddress.countryCode}
                  </p>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Right Column - Payment Action */}
          <div className="space-y-6">
            {/* Payment Status */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <CreditCard className="h-5 w-5" />
                  Payment Status
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Method</span>
                    <span className="capitalize">{payment.paymentMethod}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Gateway</span>
                    <span className="capitalize">{payment.paymentGateway}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Status</span>
                    <Badge variant={'default'}>{payment.paymentStatus}</Badge>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Countdown Timer */}

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Clock className="h-5 w-5" />
                  Time Remaining
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="text-center py-4">
                  <div className="text-3xl font-bold mb-2">
                    <MockCountdownTimer deadline={payment.paymentDeadline} />
                  </div>
                  <p className="text-sm text-muted-foreground">
                    Payment expires at{' '}
                    {new Date(payment.paymentDeadline).toLocaleString()}
                  </p>
                </div>
              </CardContent>
            </Card>

            {/* Pay Now Button */}
            {canPay && (
              <Card>
                <CardContent className="pt-6">
                  <Button
                    onClick={handlePayNow}
                    disabled={isCreatingSession}
                    size="lg"
                    className="w-full h-12 text-base font-medium"
                  >
                    <CreditCard className="mr-2 h-5 w-5" />
                    Pay with {payment.paymentGateway}
                  </Button>
                </CardContent>
              </Card>
            )}

            {/* New Order Button for Expired */}
            {isExpired && (
              <Card>
                <CardContent className="pt-6">
                  <Button
                    onClick={handleCreateNewOrder}
                    size="lg"
                    className="w-full h-12 text-base font-medium"
                  >
                    Create New Order
                  </Button>
                </CardContent>
              </Card>
            )}

            {/* Payment Instructions */}
            {!isExpired && (
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">How to Pay</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="space-y-2 text-sm">
                    <div className="flex gap-3">
                      <div className="flex-shrink-0 h-6 w-6 rounded-full bg-primary/10 flex items-center justify-center text-xs font-medium">
                        1
                      </div>
                      <p className="text-muted-foreground">
                        Click "Pay with Stripe" button
                      </p>
                    </div>
                    <div className="flex gap-3">
                      <div className="flex-shrink-0 h-6 w-6 rounded-full bg-primary/10 flex items-center justify-center text-xs font-medium">
                        2
                      </div>
                      <p className="text-muted-foreground">
                        You'll be redirected to secure Stripe checkout
                      </p>
                    </div>
                    <div className="flex gap-3">
                      <div className="flex-shrink-0 h-6 w-6 rounded-full bg-primary/10 flex items-center justify-center text-xs font-medium">
                        3
                      </div>
                      <p className="text-muted-foreground">
                        Complete your payment
                      </p>
                    </div>
                    <div className="flex gap-3">
                      <div className="flex-shrink-0 h-6 w-6 rounded-full bg-primary/10 flex items-center justify-center text-xs font-medium">
                        4
                      </div>
                      <p className="text-muted-foreground">
                        You'll be redirected back automatically
                      </p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}

            {/* Security Badge */}
            <Card>
              <CardContent className="pt-6">
                <div className="text-center space-y-2">
                  <div className="mx-auto h-8 w-8 rounded-full bg-green-100 flex items-center justify-center">
                    <ShieldCheck className="h-4 w-4 text-green-600" />
                  </div>
                  <h4 className="text-sm font-medium">Secure Payment</h4>
                  <p className="text-xs text-muted-foreground">
                    Powered by Stripe. Your payment information is encrypted and
                    secure.
                  </p>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  )
}
