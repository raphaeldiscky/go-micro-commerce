import {
  DEFAULT_PRODUCT_DIMENSIONS,
  DEFAULT_PRODUCT_WEIGHT_KG,
  DEFAULT_WAREHOUSE_ADDRESS,
  mapShippingOptionToCarrier,
} from '@/data/mockData'
import { graphClient } from '@/lib/graphql/client'
import { CREATE_ORDER_MUTATION } from '@/lib/graphql/order'
import type { CreateOrderMutation } from '@/lib/graphql/order.generated'
import { toDecimalString } from '@/lib/utils/decimal'
import type {
  Address,
  CreateOrderInput,
  PaymentGateway as GraphQLPaymentGateway,
  PaymentMethod as GraphQLPaymentMethod,
} from '@/types/__generated__/graphql'
import type {
  Cart,
  CartItem,
  CartStore,
  CheckoutSession,
  MockProduct,
  OrderSummary,
  PaymentGateway,
  PaymentMethod,
  ShippingOption,
} from '@/types/cart'
import { toast } from 'sonner'
import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import { useShallow } from 'zustand/shallow'

export const useCartStore = create<CartStore>()(
  devtools(
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
        selectedAddress: null,
        selectedShippingOption: null,
        selectedPaymentMethod: null,
        selectedPaymentGateway: null,

        // Cart and item management
        initializeCart: (customerId: string) => {
          const state = get()

          // Initialize cart if it doesn't exist
          if (!state.cart) {
            const newCart: Cart = {
              id: crypto.randomUUID(),
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
              id: crypto.randomUUID(),
              cart_id: state.cart.id,
              product_id: product.id,
              quantity,
              selected_for_checkout: true,
              added_at: new Date().toISOString(),
              product,
            }

            updatedItems = [...state.items, newItem]
          }

          set({
            items: updatedItems,
            cart: state.cart,
          })

          toast.success(`Added ${product.name} to cart`)
        },

        removeItem: (itemId: string) => {
          const state = get()
          const itemToRemove = state.items.find((item) => item.id === itemId)

          if (!itemToRemove) return

          const updatedItems = state.items.filter((item) => item.id !== itemId)

          set({
            items: updatedItems,
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

          set({
            items: updatedItems,
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
          set({ items: updatedItems })
        },

        selectAll: () => {
          const state = get()
          const updatedItems = state.items.map((item) => ({
            ...item,
            selected_for_checkout: true,
          }))
          set({ items: updatedItems })
        },

        deselectAll: () => {
          const state = get()
          const updatedItems = state.items.map((item) => ({
            ...item,
            selected_for_checkout: false,
          }))
          set({ items: updatedItems })
        },

        clearCart: () => {
          const state = get()
          set({
            items: [],
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

          // Clear any existing checkout state for clean start
          set({
            selectedAddress: null,
            selectedShippingOption: null,
            selectedPaymentMethod: null,
            selectedPaymentGateway: null,
            checkoutData: {
              orderNote: '',
              shippingMethod: '',
              paymentMethod: '',
            },
          })

          // Create checkout session
          const checkoutSession: CheckoutSession = {
            id: crypto.randomUUID(),
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

        setAddress: (address: Address) => {
          set({ selectedAddress: address })
        },

        setShippingMethod: (method: ShippingOption) => {
          const state = get()
          set({
            selectedShippingOption: method,
            checkoutData: {
              ...state.checkoutData,
              shippingMethod: method.id,
            },
          })
        },

        setPaymentMethod: (method: PaymentMethod) => {
          set({
            selectedPaymentMethod: method,
            selectedPaymentGateway: null, // Reset gateway when payment method changes
            checkoutData: {
              ...get().checkoutData,
              paymentMethod: method.id,
            },
          })
        },

        setPaymentGateway: (gateway: PaymentGateway) => {
          set({ selectedPaymentGateway: gateway })
        },

        setOrderNote: (note: string) => {
          set({
            checkoutData: {
              ...get().checkoutData,
              orderNote: note,
            },
          })
        },

        clearCheckoutState: () => {
          set({
            selectedAddress: null,
            selectedShippingOption: null,
            selectedPaymentMethod: null,
            selectedPaymentGateway: null,
            checkoutData: {
              orderNote: '',
              shippingMethod: '',
              paymentMethod: '',
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

          if (!state.selectedAddress) {
            return { success: false, error: 'Please select a delivery address' }
          }

          if (!state.selectedShippingOption) {
            return { success: false, error: 'Please select a shipping method' }
          }

          if (!state.selectedPaymentMethod) {
            return { success: false, error: 'Please select a payment method' }
          }

          if (!state.selectedPaymentGateway) {
            return { success: false, error: 'Payment gateway not configured' }
          }

          set({ isCheckoutLoading: true })

          try {
            // Build CreateOrderInput for GraphQL mutation
            const input: CreateOrderInput = {
              idempotencyKey: crypto.randomUUID(),
              items: selectedItems.map((item) => ({
                productId: item.product_id,
                quantity: item.quantity,
              })),
              paymentMethod:
                state.selectedPaymentMethod.type.toUpperCase() as GraphQLPaymentMethod,
              paymentGateway:
                state.selectedPaymentGateway.type.toUpperCase() as GraphQLPaymentGateway,
              currency: 'USD',
              shipping: {
                fromAddress: {
                  city: DEFAULT_WAREHOUSE_ADDRESS.city,
                  state: DEFAULT_WAREHOUSE_ADDRESS.state,
                  postalCode: DEFAULT_WAREHOUSE_ADDRESS.postalCode,
                  country: DEFAULT_WAREHOUSE_ADDRESS.country,
                },
                toAddress: {
                  city: state.selectedAddress.city,
                  state: state.selectedAddress.state || '',
                  postalCode: state.selectedAddress.postalCode,
                  country: state.selectedAddress.countryCode,
                },
                dimensions: {
                  length: toDecimalString(DEFAULT_PRODUCT_DIMENSIONS.length),
                  height: toDecimalString(DEFAULT_PRODUCT_DIMENSIONS.height),
                  width: toDecimalString(DEFAULT_PRODUCT_DIMENSIONS.width),
                  unit: DEFAULT_PRODUCT_DIMENSIONS.unit,
                },
                weightKg: toDecimalString(DEFAULT_PRODUCT_WEIGHT_KG),
                carrierId: mapShippingOptionToCarrier(
                  state.selectedShippingOption.id,
                ),
              },
            }

            console.log('Placing order:', input)

            // Create order using GraphQL mutation
            const data = await graphClient.request<CreateOrderMutation>(
              CREATE_ORDER_MUTATION,
              { input },
            )

            const order = data.createOrder

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

            return {
              success: true,
              orderId: order.id,
            }
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
          // Only persist cart items, not checkout state
        }),
      },
    ),
  ),
)

// Selectors for easier access
export const useCartItems = () => useCartStore((state) => state.items)
export const useCartItemCount = () =>
  useCartStore((state) =>
    state.items.reduce((total, item) => total + item.quantity, 0),
  )
export const useSelectedItems = () =>
  useCartStore(
    useShallow((state) =>
      state.items.filter((item) => item.selected_for_checkout),
    ),
  )
export const useCartTotal = () =>
  useCartStore(
    useShallow((state) => {
      const subtotal = state.items
        .filter((item) => item.selected_for_checkout)
        .reduce((total, item) => total + item.product.price * item.quantity, 0)
      const shipping = state.selectedShippingOption?.price || 0
      return {
        subtotal,
        shipping,
        discount: 0,
        total: subtotal + shipping,
      }
    }),
  )
export const useIsCartDrawerOpen = () =>
  useCartStore((state) => state.isDrawerOpen)
export const useIsCheckoutLoading = () =>
  useCartStore((state) => state.isCheckoutLoading)
export const useSelectedAddress = () =>
  useCartStore((state) => state.selectedAddress)
export const useSelectedShippingOption = () =>
  useCartStore((state) => state.selectedShippingOption)
export const useSelectedPaymentMethod = () =>
  useCartStore((state) => state.selectedPaymentMethod)
export const useSelectedPaymentGateway = () =>
  useCartStore((state) => state.selectedPaymentGateway)
export const useCheckoutSession = () =>
  useCartStore((state) => state.checkoutSession)
export const useOrderSummary = () =>
  useCartStore(
    useShallow((state) => {
      const subtotal = state.getSelectedTotal()
      const shipping = state.selectedShippingOption?.price || 0
      const total = subtotal + shipping
      return {
        subtotal,
        shipping,
        discount: 0,
        total,
      }
    }),
  )
