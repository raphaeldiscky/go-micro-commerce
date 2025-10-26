import {
  AddressSelector,
  OrderNotes,
  OrderReview,
  OrderSummary,
  PaymentGatewaySelector,
  ShippingOptions,
} from '@/components/checkout'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { PATH } from '@/constants/routes'
import { fCurrency } from '@/lib/utils/number'
import {
  useCheckoutSession,
  useCheckoutSessionStore,
} from '@/store/checkoutSessionStore'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import {
  AlertCircle,
  ArrowLeft,
  CheckCircle,
  Clock,
  MapPin,
  Package,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { toast } from 'sonner'

export const Route = createFileRoute('/checkout/$checkoutId')({
  component: RouteComponent,
})

function RouteComponent() {
  const { checkoutId } = Route.useParams()
  const navigate = useNavigate()
  const {
    checkoutSession,
    selectedDestination,
    selectedShippingOption,
    selectedPaymentGateway,
    fetchCheckoutSession,
    placeOrder,
  } = useCheckoutSessionStore()

  const [isPlacingOrder, setIsPlacingOrder] = useState(false)

  console.log('====CHECKOUT DATA====', checkoutSession)
  const data = useCheckoutSession()
  const checkoutItems = data?.items ?? []

  // Fetch checkout session on mount or when checkoutId changes
  useEffect(() => {
    const shouldFetch = !checkoutSession || checkoutSession.id !== checkoutId

    if (shouldFetch) {
      fetchCheckoutSession(checkoutId).catch((error) => {
        console.error('Failed to load checkout session:', error)
      })
    }
    // fetchCheckoutSession is a stable Zustand action and doesn't need to be in dependencies
  }, [checkoutId, checkoutSession?.id])

  // Validate checkout session
  const isValidSession = checkoutSession?.id === checkoutId

  // Check if checkout is ready
  const isCheckoutReady =
    checkoutItems.length > 0 &&
    selectedDestination &&
    selectedShippingOption &&
    selectedPaymentGateway

  const handlePlaceOrder = async () => {
    if (!isCheckoutReady) {
      toast.error('Please complete all required fields')
      return
    }

    setIsPlacingOrder(true)
    try {
      const result = await placeOrder(checkoutId)

      if (result.success) {
        toast.success('Order placed successfully!')
        // Navigate to payment page - use checkoutId which becomes the orderId
        navigate({ to: `/orders/${checkoutId}` })
      } else {
        toast.error(result.error || 'Failed to place order')
      }
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : 'An unexpected error occurred',
      )
    } finally {
      setIsPlacingOrder(false)
    }
  }

  const handleBackToCart = () => {
    window.history.back()
  }

  // Loading state
  if (!checkoutSession && !isValidSession) {
    return (
      <div className="min-h-screen bg-gray-50/40 p-4 sm:p-6 lg:p-8">
        <div className="mx-auto max-w-6xl">
          <div className="mb-8">
            <Skeleton className="h-8 w-64 mb-2" />
            <Skeleton className="h-4 w-96" />
          </div>
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

  // Invalid or expired session
  if (!isValidSession) {
    return (
      <div className="min-h-screen bg-gray-50/40 flex items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 h-12 w-12 rounded-full bg-red-100 flex items-center justify-center">
              <AlertCircle className="h-6 w-6 text-red-600" />
            </div>
            <CardTitle className="text-red-600">
              Invalid Checkout Session
            </CardTitle>
          </CardHeader>
          <CardContent className="text-center space-y-4">
            <p className="text-muted-foreground">
              We could not find your checkout session. Please start over.
            </p>
            <div className="space-y-2">
              <Button
                onClick={handleBackToCart}
                variant="outline"
                className="w-full"
              >
                <ArrowLeft className="h-4 w-4 mr-2" />
                Back to Cart
              </Button>
              <Button
                onClick={() => navigate({ to: PATH.products.root })}
                className="w-full"
              >
                Continue Shopping
              </Button>
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
        <div className="mx-auto max-w-6xl px-4 sm:px-6 lg:px-8 py-6">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleBackToCart}
            className="text-muted-foreground hover:text-foreground"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Cart
          </Button>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div>
                <h1 className="text-2xl font-bold tracking-tight">Checkout</h1>
                <p className="text-sm text-muted-foreground">
                  Session ID: {checkoutId}
                </p>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <Badge
                variant={
                  checkoutSession.status === 'PENDING' ? 'default' : 'secondary'
                }
                className="flex items-center gap-1"
              >
                {checkoutSession.status === 'PENDING' && (
                  <Clock className="h-3 w-3" />
                )}
                {checkoutSession.status}
              </Badge>
              <div className="text-sm text-muted-foreground">
                {checkoutItems.length}{' '}
                {checkoutItems.length === 1 ? 'item' : 'items'}
              </div>
            </div>
          </div>

          {/* Progress Steps */}
          <div className="mt-6 flex items-center justify-center">
            <div className="flex items-center space-x-2 lg:space-x-4">
              {/* Address Step */}
              <div className="flex items-center">
                <div
                  className={`h-8 w-8 rounded-full flex items-center justify-center text-sm font-medium ${
                    selectedDestination
                      ? 'bg-green-600 text-white'
                      : 'bg-gray-200 text-gray-500'
                  }`}
                >
                  {selectedDestination ? (
                    <CheckCircle className="h-4 w-4" />
                  ) : (
                    <MapPin className="h-4 w-4" />
                  )}
                </div>
                <span className="ml-2 text-sm font-medium hidden sm:inline">
                  Address
                </span>
              </div>

              <div className="h-px w-4 lg:w-8 bg-gray-300" />

              {/* Shipping Step */}
              <div className="flex items-center">
                <div
                  className={`h-8 w-8 rounded-full flex items-center justify-center text-sm font-medium ${
                    selectedShippingOption
                      ? 'bg-green-600 text-white'
                      : 'bg-gray-200 text-gray-500'
                  }`}
                >
                  {selectedShippingOption ? (
                    <CheckCircle className="h-4 w-4" />
                  ) : (
                    <Package className="h-4 w-4" />
                  )}
                </div>
                <span className="ml-2 text-sm font-medium hidden sm:inline">
                  Shipping
                </span>
              </div>

              <div className="h-px w-4 lg:w-8 bg-gray-300" />

              {/* Payment Gateway Step */}
              <div className="flex items-center">
                <div
                  className={`h-8 w-8 rounded-full flex items-center justify-center text-sm font-medium ${
                    selectedPaymentGateway
                      ? 'bg-green-600 text-white'
                      : 'bg-gray-200 text-gray-500'
                  }`}
                >
                  {selectedPaymentGateway ? (
                    <CheckCircle className="h-4 w-4" />
                  ) : (
                    <span>3</span>
                  )}
                </div>
                <span className="ml-2 text-sm font-medium hidden lg:inline">
                  Payment
                </span>
              </div>

              <div className="h-px w-4 lg:w-8 bg-gray-300" />

              {/* Review Step */}
              <div className="flex items-center">
                <div
                  className={`h-8 w-8 rounded-full flex items-center justify-center text-sm font-medium ${
                    isCheckoutReady
                      ? 'bg-green-600 text-white'
                      : 'bg-gray-200 text-gray-500'
                  }`}
                >
                  {isCheckoutReady ? (
                    <CheckCircle className="h-4 w-4" />
                  ) : (
                    <span>4</span>
                  )}
                </div>
                <span className="ml-2 text-sm font-medium hidden lg:inline">
                  Review
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="mx-auto max-w-6xl px-4 sm:px-6 lg:px-8 py-8">
        {checkoutItems.length === 0 ? (
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>No Items Selected</AlertTitle>
            <AlertDescription>
              Please select items from your cart to proceed with checkout.
            </AlertDescription>
          </Alert>
        ) : (
          <div className="grid gap-6 lg:grid-cols-3">
            {/* Left Column - Checkout Forms */}
            <div className="lg:col-span-2 space-y-6">
              {/* Order Review */}
              <OrderReview />

              {/* Address Selection */}
              <AddressSelector />

              {/* Shipping Options */}
              <ShippingOptions />

              {/* Payment Gateway */}
              <PaymentGatewaySelector />

              {/* Order Notes */}
              <OrderNotes />
            </div>

            {/* Right Column - Order Summary */}
            <div className="space-y-6">
              {/* Order Summary */}
              <OrderSummary />

              {/* Checkout Button */}
              <Card>
                <CardContent className="pt-6">
                  <Button
                    onClick={handlePlaceOrder}
                    disabled={!isCheckoutReady || isPlacingOrder}
                    size="lg"
                    className="w-full h-12 text-base font-medium"
                  >
                    {isPlacingOrder ? (
                      <>
                        <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent" />
                        Placing Order...
                      </>
                    ) : (
                      <>
                        Place Order •{' '}
                        {fCurrency(
                          checkoutItems.reduce(
                            (total, item) =>
                              total + Number(item.unitPrice) * item.quantity,
                            0,
                          ) + (selectedShippingOption?.price || 0),
                        )}
                      </>
                    )}
                  </Button>

                  {!isCheckoutReady && (
                    <div className="mt-3 space-y-2 text-sm text-muted-foreground">
                      {!selectedDestination && (
                        <p>• Please select a delivery address</p>
                      )}
                      {!selectedShippingOption && (
                        <p>• Please select a shipping method</p>
                      )}
                      {!selectedPaymentGateway && (
                        <p>• Please select a payment gateway</p>
                      )}
                    </div>
                  )}
                </CardContent>
              </Card>

              {/* Security Info */}
              <Card>
                <CardContent className="pt-6">
                  <div className="text-center space-y-2">
                    <div className="mx-auto h-8 w-8 rounded-full bg-green-100 flex items-center justify-center">
                      <CheckCircle className="h-4 w-4 text-green-600" />
                    </div>
                    <h4 className="text-sm font-medium">Secure Checkout</h4>
                    <p className="text-xs text-muted-foreground">
                      Your payment information is encrypted and secure. We never
                      store your payment details.
                    </p>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
