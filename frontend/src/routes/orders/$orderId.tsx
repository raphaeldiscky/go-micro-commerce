import { OrderStatusBadge } from '@/components/orders/OrderStatusBadge'
import { PaymentStatusBadge } from '@/components/payment/PaymentStatusBadge'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { PATH } from '@/constants/routes'
import { useOrderById } from '@/hooks/order'
import { usePaymentByOrderId } from '@/hooks/payment'
import { fCurrency } from '@/lib/utils/number'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { format } from 'date-fns'
import Decimal from 'decimal.js'
import {
  AlertCircle,
  ArrowLeft,
  Box,
  Calendar,
  CheckCircle,
  CreditCard,
  MapPin,
  Package,
  Truck,
  User,
} from 'lucide-react'

export const Route = createFileRoute(PATH.orders.$orderId)({
  component: RouteComponent,
})

function RouteComponent() {
  const { orderId } = Route.useParams()
  const navigate = useNavigate()

  const {
    data: order,
    isLoading: isOrderLoading,
    error: orderError,
  } = useOrderById(orderId)

  const { data: payment, isLoading: isPaymentLoading } = usePaymentByOrderId(
    orderId,
    {
      enabled: !!order,
    },
  )

  // Loading state
  if (isOrderLoading || isPaymentLoading) {
    return (
      <div className="min-h-screen bg-gray-50/40 p-4 sm:p-6 lg:p-8">
        <div className="mx-auto max-w-6xl space-y-6">
          <Skeleton className="h-8 w-64" />
          <div className="grid gap-6 lg:grid-cols-3">
            <Skeleton className="h-96 w-full lg:col-span-2" />
            <Skeleton className="h-96 w-full" />
          </div>
          <Skeleton className="h-64 w-full" />
          <Skeleton className="h-48 w-full" />
        </div>
      </div>
    )
  }

  // Error state
  if (orderError || !order) {
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

  const canRetryPayment =
    order.status === 'PAYMENT_PENDING' && payment?.clientSecret

  return (
    <div className="min-h-screen bg-gray-50/40">
      {/* Header */}
      <div className="border-b bg-white">
        <div className="mx-auto max-w-6xl px-4 sm:px-6 lg:px-8 py-6">
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
                Order Details
              </h1>
              <p className="text-sm text-muted-foreground">
                Order ID: {order.id.slice(-12)}
              </p>
            </div>
            <OrderStatusBadge status={order.status} size="lg" />
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="mx-auto max-w-6xl px-4 sm:px-6 lg:px-8 py-8 space-y-6">
        {/* Order Status Alert */}
        {(order.status === 'PAID' || order.status === 'COMPLETED') && (
          <Alert className="border-green-200 bg-green-50">
            <CheckCircle className="h-4 w-4 text-green-600" />
            <AlertTitle className="text-green-600">Order Confirmed!</AlertTitle>
            <AlertDescription className="text-green-600">
              Your order has been successfully placed and payment has been
              processed. You will receive a confirmation email shortly.
            </AlertDescription>
          </Alert>
        )}

        {order.status === 'PAYMENT_PENDING' && (
          <Alert>
            <Package className="h-4 w-4" />
            <AlertTitle>Payment Pending</AlertTitle>
            <AlertDescription>
              Your order is created and awaiting payment completion. Please
              complete the payment to process your order.
            </AlertDescription>
          </Alert>
        )}

        {(order.status === 'PROCESSING' || order.status === 'PENDING') && (
          <Alert>
            <Package className="h-4 w-4" />
            <AlertTitle>Order Processing</AlertTitle>
            <AlertDescription>
              Your order is being processed. We'll notify you once it's ready
              for shipment.
            </AlertDescription>
          </Alert>
        )}

        {(order.status === 'FAILED' || order.status === 'CANCELED') && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>
              Order {order.status === 'CANCELED' ? 'Canceled' : 'Failed'}
            </AlertTitle>
            <AlertDescription>
              There was an issue with your order. Please contact support or try
              placing the order again.
            </AlertDescription>
          </Alert>
        )}

        <div className="grid gap-6 lg:grid-cols-3">
          {/* Order Items */}
          <div className="lg:col-span-2">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Box className="h-5 w-5" />
                  Order Items ({order.items.length || 0})
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {order.items.map((item) => (
                    <div key={item.id} className="flex gap-4">
                      {/* Product Image Placeholder */}
                      <div className="w-16 h-16 bg-gray-200 rounded-lg flex items-center justify-center">
                        <Package className="h-6 w-6 text-gray-400" />
                      </div>

                      {/* Product Details */}
                      <div className="flex-1 min-w-0">
                        <h3 className="font-medium text-sm leading-tight mb-1">
                          {item.productName}
                        </h3>
                        <p className="text-xs text-muted-foreground mb-2">
                          Product ID: {item.productId.slice(-8)}
                        </p>
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-4 text-sm text-muted-foreground">
                            <span>Qty: {item.quantity}</span>
                            <span>
                              {fCurrency(item.unitPrice, order.currency)} each
                            </span>
                          </div>
                          <div className="text-sm font-medium">
                            {fCurrency(item.totalPrice, order.currency)}
                          </div>
                        </div>

                        {/* Tax and Discount Details */}
                        {(Decimal(item.totalTax).greaterThan(0) ||
                          Decimal(item.totalDiscount).greaterThan(0)) && (
                          <div className="mt-2 pt-2 border-t text-xs space-y-1">
                            {Decimal(item.totalTax).greaterThan(0) && (
                              <div className="flex justify-between text-muted-foreground">
                                <span>Tax</span>
                                <span>
                                  {fCurrency(item.totalTax, order.currency)}
                                </span>
                              </div>
                            )}
                            {Decimal(item.totalDiscount).greaterThan(0) && (
                              <div className="flex justify-between text-green-600">
                                <span>Discount</span>
                                <span>
                                  -
                                  {fCurrency(
                                    item.totalDiscount,
                                    order.currency,
                                  )}
                                </span>
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>

                <Separator className="my-6" />

                {/* Order Summary */}
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Subtotal</span>
                    <span>{fCurrency(order.subtotal, order.currency)}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Shipping</span>
                    <span>{fCurrency(order.shippingCost, order.currency)}</span>
                  </div>
                  {Decimal(order.totalTax).greaterThan(0) && (
                    <div className="flex justify-between text-sm">
                      <span className="text-muted-foreground">Tax</span>
                      <span>{fCurrency(order.totalTax, order.currency)}</span>
                    </div>
                  )}
                  {Decimal(order.totalDiscount).greaterThan(0) && (
                    <div className="flex justify-between text-sm text-green-600">
                      <span>Discount</span>
                      <span>
                        -{fCurrency(order.totalDiscount, order.currency)}
                      </span>
                    </div>
                  )}
                  <Separator className="my-2" />
                  <div className="flex justify-between font-semibold text-base">
                    <span>Total</span>
                    <span className="text-primary">
                      {fCurrency(order.totalPrice, order.currency)}
                    </span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Order Information */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-lg">
                  <Package className="h-5 w-5" />
                  Order Information
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-3 text-sm">
                  <div className="flex items-center gap-2">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                    <span className="text-muted-foreground">Order Date:</span>
                  </div>
                  <p className="font-medium">
                    {format(new Date(order.createdAt), 'MMM dd, yyyy HH:mm')}
                  </p>

                  <div className="flex items-center gap-2">
                    <User className="h-4 w-4 text-muted-foreground" />
                    <span className="text-muted-foreground">Customer ID:</span>
                  </div>
                  <p className="font-mono text-xs">
                    {order.customerId.slice(-8)}
                  </p>

                  <div className="flex items-center gap-2">
                    <CreditCard className="h-4 w-4 text-muted-foreground" />
                    <span className="text-muted-foreground">Payment:</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className="capitalize">
                      {order.paymentGateway}
                    </Badge>
                    {payment && <PaymentStatusBadge status={payment.status} />}
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Shipping Information */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-lg">
                  <Truck className="h-5 w-5" />
                  Shipping Information
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <MapPin className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">
                      Delivery Address
                    </span>
                  </div>
                  <div className="bg-muted p-3 rounded-lg text-sm">
                    <p className="font-medium">
                      {order.destination.city}, {order.destination.state}
                    </p>
                    <p className="text-muted-foreground">
                      {order.destination.postalCode}
                    </p>
                    <p className="text-muted-foreground">
                      {order.destination.countryCode}
                    </p>
                  </div>

                  <div className="flex items-center gap-2">
                    <Box className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">Package Details</span>
                  </div>
                  <div className="bg-muted p-3 rounded-lg text-sm space-y-1">
                    <p>Weight: {order.package.weightKg} kg</p>
                    <p>
                      Dimensions: {order.package.length} × {order.package.width}{' '}
                      × {order.package.height} {order.package.unit}
                    </p>
                    <p>Courier: {order.courier.courierId}</p>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Payment Actions */}
            {canRetryPayment && (
              <Card>
                <CardContent className="pt-6">
                  <Button
                    onClick={() =>
                      navigate({ to: PATH.payment.detail(orderId) })
                    }
                    className="w-full"
                    size="lg"
                  >
                    Complete Payment
                  </Button>
                </CardContent>
              </Card>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
