import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils/index'
import { useCartItemCount, useIsCartDrawerOpen, useCartStore } from '@/store/cartStore'
import { ShoppingBag } from 'lucide-react'
import { forwardRef } from 'react'

interface CartIconProps {
  className?: string
}

export const CartIcon = forwardRef<HTMLButtonElement, CartIconProps>(
  ({ className }, ref) => {
    const itemCount = useCartItemCount()
    const isDrawerOpen = useIsCartDrawerOpen()
    const toggleDrawer = useCartStore((state) => state.toggleDrawer)

    const handleToggleCart = () => {
      toggleDrawer()
    }

    return (
      <Button
        ref={ref}
        aria-label={`Shopping cart with ${itemCount} items`}
        className={cn(
          'relative h-10 w-10 rounded-full shadow-lg transition-all duration-200 hover:scale-105 active:scale-95',
          isDrawerOpen && 'ring-2 ring-primary ring-offset-2',
          className,
        )}
        onClick={handleToggleCart}
        size="icon"
        variant="outline"
      >
        <ShoppingBag className="h-5 w-5" />

        {/* Badge showing total item count */}
        {itemCount > 0 && (
          <Badge
            className={cn(
              'absolute -right-1 -top-1 flex h-5 w-5 items-center justify-center rounded-full p-0 text-xs font-medium',
              'bg-destructive text-destructive-foreground border-2 border-background',
              'animate-pulse',
            )}
            variant="destructive"
          >
            {itemCount > 99 ? '99+' : itemCount}
          </Badge>
        )}

        {/* Visual feedback when drawer is open */}
        {isDrawerOpen && (
          <div className="absolute inset-0 rounded-full bg-primary/10" />
        )}
      </Button>
    )
  },
)

CartIcon.displayName = 'CartIcon'