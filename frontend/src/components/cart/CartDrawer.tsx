import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { PATH_FEATURES } from '@/constants/routes'
import { formatCurrency } from '@/data/mockData'
import { useCartStore } from '@/store/cartStore'
import { useNavigate } from '@tanstack/react-router'
import { CheckCheck, Package, ShoppingBag } from 'lucide-react'
import { useMemo } from 'react'
import { CartItemRow } from './CartItemRow'

export function CartDrawer() {
  const navigate = useNavigate()
  const {
    items,
    isDrawerOpen,
    closeDrawer,
    getSelectedItems,
    getTotalItemCount,
    startCheckout,
    selectAll,
    deselectAll,
    clearCart,
  } = useCartStore()

  const selectedItems = getSelectedItems()
  const totalItemCount = getTotalItemCount()
  const selectedTotal = useMemo(() => {
    return selectedItems.reduce(
      (total, item) => total + item.product.price * item.quantity,
      0,
    )
  }, [selectedItems])

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      closeDrawer()
    }
  }

  const handleSelectAll = () => {
    const hasUnselectedItems = items.some((item) => !item.selected_for_checkout)
    if (hasUnselectedItems) {
      selectAll()
    } else {
      deselectAll()
    }
  }

  const handleCheckout = () => {
    if (selectedItems.length === 0) {
      return
    }

    // Navigate to checkout with ID
    const navigateToCheckout = (checkoutId: string) => {
      navigate({ to: PATH_FEATURES.checkout.detail(checkoutId) })
    }

    startCheckout(navigateToCheckout)
  }

  const hasSelectedItems = selectedItems.length > 0
  const allItemsSelected =
    items.length > 0 && items.every((item) => item.selected_for_checkout)

  return (
    <Sheet onOpenChange={handleOpenChange} open={isDrawerOpen}>
      <SheetContent
        className="w-full sm:max-w-md p-0 flex flex-col"
        onOpenAutoFocus={(e) => e.preventDefault()}
        side="right"
      >
        {/* Header Section */}
        <div className="flex-shrink-0 border-b px-6 py-4">
          <SheetHeader className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <ShoppingBag className="h-5 w-5" />
                <SheetTitle>Shopping Cart</SheetTitle>
              </div>
              {items.length > 0 && (
                <div className="text-sm text-muted-foreground">
                  {totalItemCount} {totalItemCount === 1 ? 'item' : 'items'}
                </div>
              )}
            </div>
            <SheetDescription>
              Review your items and proceed to checkout
            </SheetDescription>
          </SheetHeader>
        </div>

        {/* Cart Items Section */}
        <div className="flex-1 overflow-hidden flex flex-col">
          {items.length === 0 ? (
            <div className="flex-1 flex items-center justify-center p-8">
              <div className="text-center space-y-4">
                <div className="mx-auto h-16 w-16 rounded-full bg-muted flex items-center justify-center">
                  <Package className="h-8 w-8 text-muted-foreground" />
                </div>
                <div>
                  <h3 className="font-semibold">Your cart is empty</h3>
                  <p className="text-sm text-muted-foreground mt-1">
                    Add some products to get started
                  </p>
                </div>
              </div>
            </div>
          ) : (
            <>
              {/* Selection Controls */}
              <div className="flex-shrink-0 px-6 py-3 border-b">
                <div className="flex items-center justify-between">
                  <Button
                    onClick={handleSelectAll}
                    size="sm"
                    variant="ghost"
                    className="h-8 px-3"
                  >
                    {allItemsSelected ? (
                      <>
                        <CheckCheck className="h-4 w-4 mr-2" />
                        Deselect All
                      </>
                    ) : (
                      <>
                        <CheckCheck className="h-4 w-4 mr-2" />
                        Select All
                      </>
                    )}
                  </Button>
                  <Button
                    onClick={clearCart}
                    size="sm"
                    variant="ghost"
                    className="h-8 px-3 text-muted-foreground hover:text-destructive"
                  >
                    Clear Cart
                  </Button>
                </div>
              </div>

              {/* Cart Items List */}
              <ScrollArea className="flex-1">
                <div className="px-6 py-2">
                  {items.map((item) => (
                    <CartItemRow key={item.id} item={item} />
                  ))}
                </div>
              </ScrollArea>

              {/* Selected Items Summary */}
              {hasSelectedItems && (
                <div className="flex-shrink-0 border-t">
                  <div className="px-6 py-4 space-y-3">
                    <div className="space-y-2">
                      {selectedItems.slice(0, 3).map((item) => (
                        <div
                          key={item.id}
                          className="flex justify-between text-sm"
                        >
                          <span className="text-muted-foreground truncate">
                            {item.product.name} x{item.quantity}
                          </span>
                          <span className="font-medium">
                            {formatCurrency(item.product.price * item.quantity)}
                          </span>
                        </div>
                      ))}
                      {selectedItems.length > 3 && (
                        <div className="text-sm text-muted-foreground">
                          +{selectedItems.length - 3} more items
                        </div>
                      )}
                    </div>

                    <Separator />

                    <div className="flex justify-between items-center">
                      <span className="font-semibold">Selected Total:</span>
                      <span className="font-bold text-lg">
                        {formatCurrency(selectedTotal)}
                      </span>
                    </div>

                    <div className="text-xs text-muted-foreground">
                      {selectedItems.length}{' '}
                      {selectedItems.length === 1 ? 'item' : 'items'} selected
                    </div>
                  </div>
                </div>
              )}
            </>
          )}
        </div>

        {/* Footer with Checkout Button */}
        {items.length > 0 && (
          <div className="flex-shrink-0 border-t p-6">
            <Button
              disabled={!hasSelectedItems}
              onClick={handleCheckout}
              className="w-full"
              size="lg"
            >
              {hasSelectedItems ? (
                <>
                  Checkout ({selectedItems.length}{' '}
                  {selectedItems.length === 1 ? 'item' : 'items'})
                </>
              ) : (
                <>
                  <ShoppingBag className="h-4 w-4 mr-2" />
                  Select items to checkout
                </>
              )}
            </Button>
            {!hasSelectedItems && items.length > 0 && (
              <p className="text-xs text-muted-foreground text-center mt-2">
                Select items above to proceed with checkout
              </p>
            )}
          </div>
        )}
      </SheetContent>
    </Sheet>
  )
}
