import type {
  Cart,
  CartItem,
  CartStore,
  CheckoutSession,
  MockProduct,
  OrderSummary,
  PaymentMethod,
  ShippingOption,
} from '@/types/cart'
import { toast } from 'sonner'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

const generateId = () =>
  `cart-${Date.now()}-${Math.random().toString(36).substring(7)}`

const generateCartId = () =>
  `cart-${Date.now()}-${Math.random().toString(36).substring(7)}`

// Helper function to compute derived state
const computeDerivedState = (
  items: Array<CartItem>,
  selectedShippingOption: ShippingOption | null
) => {
  const selectedItems = items.filter((item) => item.selected_for_checkout)
  const totalItemCount = items.reduce((total, item) => total + item.quantity, 0)
  const subtotal = selectedItems.reduce(
    (total, item) => total + item.product.price * item.quantity,
    0,
  )
  const shipping = selectedShippingOption?.price || 0
  const total = subtotal + shipping

  return {
    selectedItems,
    totalItemCount,
    orderSummary: {
      subtotal,
      shipping,
      discount: 0,
      total,
    },
  }
}

export const useCartStore = create<CartStore>()(
  persist(
    (set, get) => ({
      // Initial state
      cart: null,
      items: [],
      checkoutSession: null,
      isLoading: false,
      isDrawerOpen: false,
      isCheckoutLoading: false,
      checkoutData: {
        orderNote: '',
        shippingMethod: '',
        paymentMethod: '',
      },
      selectedShippingOption: null,
      selectedPaymentMethod: null,

      // Computed state (cached to prevent infinite loops)
      selectedItems: [],
      orderSummary: {
        subtotal: 0,
        shipping: 0,
        discount: 0,
        total: 0,
      },
      totalItemCount: 0,

      // Cart and item management
      initializeCart: (customerId: string) => {
        const state = get()

        // Initialize cart if it doesn't exist
        if (!state.cart) {
          const newCart: Cart = {
            id: generateCartId(),
            customer_id: customerId,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          }

          set({ cart: newCart })
        }
      },

      addItem: (product: MockProduct, quantity = 1) => {
        const state = get()
        if (!state.cart) {
          toast.error('Cart not initialized')
          return
        }

        // Check if product is available
        const availableQuantity = product.quantity - product.reservedQuantity
        if (availableQuantity < quantity) {
          toast.error(`Only ${availableQuantity} items available`)
          return
        }

        // Check if item already exists in cart
        const existingItemIndex = state.items.findIndex(
          (item) => item.product_id === product.id,
        )

        let updatedItems: Array<CartItem>

        if (existingItemIndex >= 0) {
          // Update existing item quantity
          const existingItem = state.items[existingItemIndex]
          const newQuantity = existingItem.quantity + quantity

          if (availableQuantity < newQuantity) {
            toast.error(`Only ${availableQuantity} items available`)
            return
          }

          updatedItems = state.items.map((item, index) =>
            index === existingItemIndex
              ? {
                  ...item,
                  quantity: newQuantity,
                  selected_for_checkout: true, // Auto-select on add
                }
              : item,
          )
        } else {
          // Add new item
          const newItem: CartItem = {
            id: generateId(),
            cart_id: state.cart.id,
            product_id: product.id,
            quantity,
            selected_for_checkout: true,
            added_at: new Date().toISOString(),
            product,
          }

          updatedItems = [...state.items, newItem]
        }

        const derivedState = computeDerivedState(updatedItems, state.selectedShippingOption)

        set({
          items: updatedItems,
          ...derivedState,
          cart: {
            ...state.cart,
            updated_at: new Date().toISOString(),
          },
        })

        toast.success(`Added ${product.name} to cart`)
      },

      removeItem: (itemId: string) => {
        const state = get()
        const itemToRemove = state.items.find((item) => item.id === itemId)

        if (!itemToRemove) return

        const updatedItems = state.items.filter((item) => item.id !== itemId)
        const derivedState = computeDerivedState(updatedItems, state.selectedShippingOption)

        set({
          items: updatedItems,
          ...derivedState,
          cart: state.cart
            ? {
                ...state.cart,
                updated_at: new Date().toISOString(),
              }
            : null,
        })

        toast.success(`Removed ${itemToRemove.product.name} from cart`)
      },

      updateQuantity: (itemId: string, quantity: number) => {
        const state = get()
        const item = state.items.find((x) => x.id === itemId)

        if (!item) return

        // Validate quantity
        if (quantity <= 0) {
          get().removeItem(itemId)
          return
        }

        // Check if quantity is available
        const availableQuantity =
          item.product.quantity - item.product.reservedQuantity
        if (availableQuantity < quantity) {
          toast.error(`Only ${availableQuantity} items available`)
          return
        }

        const updatedItems = state.items.map((x) =>
          x.id === itemId ? { ...x, quantity } : x,
        )
        const derivedState = computeDerivedState(updatedItems, state.selectedShippingOption)

        set({
          items: updatedItems,
          ...derivedState,
          cart: state.cart
            ? {
                ...state.cart,
                updated_at: new Date().toISOString(),
              }
            : null,
        })
      },

      toggleSelection: (itemId: string) => {
        const state = get()
        const updatedItems = state.items.map((item) =>
          item.id === itemId
            ? { ...item, selected_for_checkout: !item.selected_for_checkout }
            : item,
        )
        const derivedState = computeDerivedState(updatedItems, state.selectedShippingOption)

        set({ items: updatedItems, ...derivedState })
      },

      selectAll: () => {
        const state = get()
        const updatedItems = state.items.map((item) => ({
          ...item,
          selected_for_checkout: true,
        }))
        const derivedState = computeDerivedState(updatedItems, state.selectedShippingOption)

        set({ items: updatedItems, ...derivedState })
      },

      deselectAll: () => {
        const state = get()
        const updatedItems = state.items.map((item) => ({
          ...item,
          selected_for_checkout: false,
        }))
        const derivedState = computeDerivedState(updatedItems, state.selectedShippingOption)

        set({ items: updatedItems, ...derivedState })
      },

      clearCart: () => {
        const state = get()
        set({
          items: [],
          selectedItems: [],
          totalItemCount: 0,
          orderSummary: {
            subtotal: 0,
            shipping: 0,
            discount: 0,
            total: 0,
          },
          cart: state.cart
            ? {
                ...state.cart,
                updated_at: new Date().toISOString(),
              }
            : null,
        })

        // Clear checkout data
        set({
          checkoutData: {
            orderNote: '',
            shippingMethod: '',
            paymentMethod: '',
          },
          selectedShippingOption: null,
          selectedPaymentMethod: null,
          checkoutSession: null,
        })
      },

      // Drawer management
      openDrawer: () => set({ isDrawerOpen: true }),
      closeDrawer: () => set({ isDrawerOpen: false }),
      toggleDrawer: () =>
        set((state) => ({ isDrawerOpen: !state.isDrawerOpen })),

      // Checkout flow
      startCheckout: (navigateToCheckout?: (checkoutId: string) => void) => {
        const state = get()
        const selectedItems = state.getSelectedItems()

        if (selectedItems.length === 0) {
          toast.error('Please select items to checkout')
          return
        }

        // Create checkout session
        const checkoutSession: CheckoutSession = {
          id: generateId(),
          customer_id: state.cart?.customer_id || '',
          cart_id: state.cart?.id || '',
          status: 'pending',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        }

        set({ checkoutSession, isDrawerOpen: false })
        toast.success('Proceeding to checkout')

        // Navigate to checkout page with ID if navigation function provided
        if (navigateToCheckout && checkoutSession.id) {
          navigateToCheckout(checkoutSession.id)
        }
      },

      
      setShippingMethod: (method: ShippingOption) => {
        const state = get()
        const derivedState = computeDerivedState(state.items, method)

        set({
          selectedShippingOption: method,
          ...derivedState,
          checkoutData: {
            ...state.checkoutData,
            shippingMethod: method.id,
          },
        })
      },

      setPaymentMethod: (method: PaymentMethod) => {
        set({
          selectedPaymentMethod: method,
          checkoutData: {
            ...get().checkoutData,
            paymentMethod: method.id,
          },
        })
      },

      setOrderNote: (note: string) => {
        set({
          checkoutData: {
            ...get().checkoutData,
            orderNote: note,
          },
        })
      },

      placeOrder: async (): Promise<{
        success: boolean
        orderId?: string
        error?: string
      }> => {
        const state = get()
        const selectedItems = state.getSelectedItems()

        if (selectedItems.length === 0) {
          return { success: false, error: 'No items selected' }
        }

        if (!state.selectedShippingOption) {
          return { success: false, error: 'Please select a shipping method' }
        }

        if (!state.selectedPaymentMethod) {
          return { success: false, error: 'Please select a payment method' }
        }

        set({ isCheckoutLoading: true })

        try {
          // Simulate API call
          await new Promise((resolve) => setTimeout(resolve, 2000))

          // Generate order ID
          const orderId = `order-${Date.now()}-${Math.random().toString(36).substring(7)}`

          // Update checkout session
          if (state.checkoutSession) {
            set({
              checkoutSession: {
                ...state.checkoutSession,
                status: 'placed',
                updated_at: new Date().toISOString(),
              },
            })
          }

          // Clear cart after successful order
          state.clearCart()

          set({ isCheckoutLoading: false })
          toast.success('Order placed successfully!')

          return { success: true, orderId }
        } catch (error) {
          set({ isCheckoutLoading: false })
          const errorMessage =
            error instanceof Error ? error.message : 'Order placement failed'
          toast.error(errorMessage)

          // Update checkout session as failed
          if (state.checkoutSession) {
            set({
              checkoutSession: {
                ...state.checkoutSession,
                status: 'failed',
                updated_at: new Date().toISOString(),
              },
            })
          }

          return { success: false, error: errorMessage }
        }
      },

      // Utility selectors
      getTotalItemCount: () => {
        const state = get()
        return state.items.reduce((total, item) => total + item.quantity, 0)
      },

      getSelectedItems: () => {
        const state = get()
        return state.items.filter((item) => item.selected_for_checkout)
      },

      getSelectedTotal: () => {
        const state = get()
        return state
          .getSelectedItems()
          .reduce(
            (total, item) => total + item.product.price * item.quantity,
            0,
          )
      },

      getSubtotal: () => {
        const state = get()
        return state.items.reduce(
          (total, item) => total + item.product.price * item.quantity,
          0,
        )
      },

      getOrderSummary: (): OrderSummary => {
        const state = get()
        const subtotal = state.getSelectedTotal()
        const shipping = state.selectedShippingOption?.price || 0
        const total = subtotal + shipping

        return {
          subtotal,
          shipping,
          discount: 0,
          total,
        }
      },
    }),
    {
      name: 'cart-store',
      partialize: (state) => ({
        cart: state.cart,
        items: state.items,
        selectedItems: state.selectedItems,
        orderSummary: state.orderSummary,
        totalItemCount: state.totalItemCount,
        checkoutSession: state.checkoutSession,
        checkoutData: state.checkoutData,
        selectedShippingOption: state.selectedShippingOption,
        selectedPaymentMethod: state.selectedPaymentMethod,
      }),
    },
  ),
)

// Selectors for easier access (using cached computed state to prevent infinite loops)
export const useCartItems = () => useCartStore((state) => state.items)
export const useCartItemCount = () => useCartStore((state) => state.totalItemCount)
export const useSelectedItems = () => useCartStore((state) => state.selectedItems)
export const useCartTotal = () => useCartStore((state) => state.orderSummary)
export const useIsCartDrawerOpen = () =>
  useCartStore((state) => state.isDrawerOpen)
export const useIsCheckoutLoading = () =>
  useCartStore((state) => state.isCheckoutLoading)
export const useSelectedShippingOption = () =>
  useCartStore((state) => state.selectedShippingOption)
export const useSelectedPaymentMethod = () =>
  useCartStore((state) => state.selectedPaymentMethod)
export const useCheckoutSession = () =>
  useCartStore((state) => state.checkoutSession)
