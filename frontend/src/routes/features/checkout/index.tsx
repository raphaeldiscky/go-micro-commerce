import { OrderNotes } from '@/components/checkout/OrderNotes'
import { OrderReview } from '@/components/checkout/OrderReview'
import { OrderSummary } from '@/components/checkout/OrderSummary'
import { PaymentMethods } from '@/components/checkout/PaymentMethods'
import { ShippingOptions } from '@/components/checkout/ShippingOptions'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { PATH_AUTH, PATH_FEATURES } from '@/constants/routes'
import { useUser } from '@/store/authStore'
import { useCartStore } from '@/store/cartStore'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { ArrowLeft, CheckCircle, Loader2 } from 'lucide-react'
import { toast } from 'sonner'

export const Route = createFileRoute('/features/checkout/')({
  component: CheckoutPage,
})

function CheckoutPage() {
  const navigate = useNavigate()
  const user = useUser()
  const {
    getSelectedItems,
    selectedShippingOption,
    selectedPaymentMethod,
    isCheckoutLoading,
    placeOrder,
  } = useCartStore()

  const selectedItems = getSelectedItems()

  if (!user) {
    navigate({ to: PATH_AUTH.login })
    toast.error('Please login to proceed with checkout')
    return null
  }

  if (selectedItems.length === 0) {
    navigate({ to: PATH_FEATURES.products.root })
    toast.error('No items selected for checkout')
    return null
  }

  const handlePlaceOrder = async () => {
    if (!selectedShippingOption) {
      toast.error('Please select a shipping method')
      return
    }

    if (!selectedPaymentMethod) {
      toast.error('Please select a payment method')
      return
    }

    const result = await placeOrder()

    if (result.success) {
      toast.success('Order placed successfully!')
      navigate({ to: PATH_FEATURES.products.root })
    } else {
      toast.error(result.error || 'Failed to place order')
    }
  }

  const handleBackToProducts = () => {
    navigate({ to: PATH_FEATURES.products.root })
  }

  const canPlaceOrder =
    selectedItems.length > 0 && selectedShippingOption && selectedPaymentMethod

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center gap-4 mb-4">
          <Button
            onClick={handleBackToProducts}
            variant="ghost"
            size="sm"
            className="flex items-center gap-2"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to Products
          </Button>
        </div>
        <div className="space-y-2">
          <h1 className="text-3xl font-bold tracking-tight">Checkout</h1>
          <p className="text-muted-foreground">
            Review your order and complete your purchase
          </p>
        </div>
      </div>

      {/* Progress Indicator */}
      <div className="mb-8">
        <div className="flex items-center justify-center space-x-4 md:space-x-8">
          <div className="flex items-center space-x-2">
            <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary text-primary-foreground text-sm font-medium">
              1
            </div>
            <span className="text-sm font-medium hidden sm:inline">Review</span>
          </div>
          <Separator className="w-12 md:w-20" />
          <div className="flex items-center space-x-2">
            <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary text-primary-foreground text-sm font-medium">
              2
            </div>
            <span className="text-sm font-medium hidden sm:inline">
              Details
            </span>
          </div>
          <Separator className="w-12 md:w-20" />
          <div className="flex items-center space-x-2">
            <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary text-primary-foreground text-sm font-medium">
              3
            </div>
            <span className="text-sm font-medium hidden sm:inline">
              Complete
            </span>
          </div>
        </div>
      </div>

      {/* Checkout Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Order Review */}
          <OrderReview />

          {/* Order Notes */}
          <OrderNotes />
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Shipping Options */}
          <ShippingOptions />

          {/* Payment Methods */}
          <PaymentMethods />

          {/* Order Summary */}
          <OrderSummary />

          {/* Place Order Button */}
          <div className="space-y-4">
            <Button
              disabled={!canPlaceOrder || isCheckoutLoading}
              onClick={handlePlaceOrder}
              className="w-full"
              size="lg"
            >
              {isCheckoutLoading ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Processing Order...
                </>
              ) : (
                <>
                  <CheckCircle className="h-4 w-4 mr-2" />
                  Place Order
                </>
              )}
            </Button>

            <div className="text-xs text-center text-muted-foreground">
              By placing this order, you agree to our Terms of Service and
              Privacy Policy
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
