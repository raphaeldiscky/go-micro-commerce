import {
  CANCEL_CHECKOUT_SESSION_MUTATION,
  CREATE_CHECKOUT_SESSION_MUTATION,
  GET_CHECKOUT_SESSION_QUERY,
  PLACE_ORDER_MUTATION,
} from '@/lib/graphql/cart'
import type {
  CancelCheckoutSessionMutation,
  CreateCheckoutSessionMutation,
  GetCheckoutSessionQuery,
  PlaceOrderMutation,
} from '@/lib/graphql/cart.generated'
import { graphClient } from '@/lib/graphql/client'
import type { Address } from '@/types/__generated__/graphql'
import type {
  OrderSummary,
  PaymentGatewayUI,
  PaymentMethodUI,
  ShippingOptionUI,
} from '@/types/cart'
import { toast } from 'sonner'
import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'

// Extract CheckoutSession type from GraphQL
type CheckoutSession = NonNullable<
  GetCheckoutSessionQuery['getCheckoutSession']
>

// Checkout session store state interface
export interface CheckoutSessionState {
  checkoutSession: CheckoutSession | null
  isLoading: boolean
  selectedAddressId: string | null
  selectedCarrierId: string | null
  selectedPaymentMethod: string | null
  selectedPaymentGateway: string | null
  orderNote?: string

  // UI-specific selections for display purposes
  selectedAddress: Address | null
  selectedShippingOption: ShippingOptionUI | null
  selectedPaymentMethodData: PaymentMethodUI | null
  selectedPaymentGatewayData: PaymentGatewayUI | null
}

// Checkout session store actions interface
export interface CheckoutSessionActions {
  // Checkout session management
  fetchCheckoutSession: (sessionId: string) => Promise<void>
  startCheckout: (
    navigateToCheckout?: (checkoutId: string) => void,
  ) => Promise<void>
  cancelCheckout: (sessionId: string) => Promise<void>

  // Set selections for checkout
  setAddress: (addressId: string, address: Address) => void
  setShippingMethod: (carrierId: string, method: ShippingOptionUI) => void
  setPaymentMethod: (method: string, methodData: PaymentMethodUI) => void
  setPaymentGateway: (gateway: string, gatewayData: PaymentGatewayUI) => void
  setOrderNote: (note: string) => void

  // Place order
  placeOrder: (sessionId: string) => Promise<{
    success: boolean
    sessionId?: string
    error?: string
  }>

  // Utility
  getOrderSummary: () => OrderSummary
  clearCheckout: () => void
}

export type CheckoutSessionStore = CheckoutSessionState & CheckoutSessionActions

export const useCheckoutSessionStore = create<CheckoutSessionStore>()(
  devtools(
    persist(
      (set, get) => ({
        // Initial state
        checkoutSession: null,
        isLoading: false,
        selectedAddressId: null,
        selectedCarrierId: null,
        selectedPaymentMethod: null,
        selectedPaymentGateway: null,
        selectedAddress: null,
        selectedShippingOption: null,
        selectedPaymentMethodData: null,
        selectedPaymentGatewayData: null,
        checkoutData: {
          orderNote: '',
        },

        // Checkout session management
        fetchCheckoutSession: async (sessionId: string) => {
          set({ isLoading: true })

          try {
            const data = await graphClient.request<GetCheckoutSessionQuery>(
              GET_CHECKOUT_SESSION_QUERY,
              { id: sessionId },
            )

            set({
              checkoutSession: data.getCheckoutSession,
              isLoading: false,
            })
          } catch (error) {
            set({ isLoading: false })
            console.error('Failed to fetch checkout session:', error)
            const errorMessage =
              error instanceof Error
                ? error.message
                : 'Failed to fetch checkout session'
            toast.error(errorMessage)
            throw error
          }
        },

        startCheckout: async (
          navigateToCheckout?: (checkoutId: string) => void,
        ) => {
          set({ isLoading: true })

          try {
            // Create checkout session via GraphQL
            const data =
              await graphClient.request<CreateCheckoutSessionMutation>(
                CREATE_CHECKOUT_SESSION_MUTATION,
                {
                  input: {
                    idempotencyKey: crypto.randomUUID(),
                  },
                },
              )

            const session = data.createCheckoutSession

            set({
              checkoutSession: session,
              isLoading: false,
              selectedAddressId: session.addressId,
              selectedCarrierId: session.carrierId,
              selectedPaymentMethod: session.paymentMethod,
              selectedPaymentGateway: session.paymentGateway,
            })

            toast.success('Proceeding to checkout')

            // Navigate to checkout page with ID if navigation function provided
            if (navigateToCheckout && session.id) {
              navigateToCheckout(session.id)
            }
          } catch (error) {
            set({ isLoading: false })
            const errorMessage =
              error instanceof Error
                ? error.message
                : 'Failed to start checkout'
            toast.error(errorMessage)
            throw error
          }
        },

        cancelCheckout: async (sessionId: string) => {
          set({ isLoading: true })

          try {
            await graphClient.request<CancelCheckoutSessionMutation>(
              CANCEL_CHECKOUT_SESSION_MUTATION,
              { sessionId },
            )

            get().clearCheckout()
            toast.success('Checkout canceled')
          } catch (error) {
            set({ isLoading: false })
            const errorMessage =
              error instanceof Error
                ? error.message
                : 'Failed to cancel checkout'
            toast.error(errorMessage)
            throw error
          }
        },

        // Set selections for checkout
        setAddress: (addressId: string, address: Address) => {
          set({
            selectedAddressId: addressId,
            selectedAddress: address,
          })
        },

        setShippingMethod: (carrierId: string, method: ShippingOptionUI) => {
          set({
            selectedCarrierId: carrierId,
            selectedShippingOption: method,
          })
        },

        setPaymentMethod: (method: string, methodData: PaymentMethodUI) => {
          set({
            selectedPaymentMethod: method,
            selectedPaymentMethodData: methodData,
            // Reset gateway when payment method changes
            selectedPaymentGateway: null,
            selectedPaymentGatewayData: null,
          })
        },

        setPaymentGateway: (gateway: string, gatewayData: PaymentGatewayUI) => {
          set({
            selectedPaymentGateway: gateway,
            selectedPaymentGatewayData: gatewayData,
          })
        },

        setOrderNote: (note: string) => {
          set({
            orderNote: note,
          })
        },

        // Place order with all required fields
        placeOrder: async (
          sessionId: string,
        ): Promise<{
          success: boolean
          sessionId?: string
          error?: string
        }> => {
          const state = get()

          // Validation
          if (!state.selectedAddressId) {
            return { success: false, error: 'Please select a delivery address' }
          }

          if (!state.selectedCarrierId) {
            return { success: false, error: 'Please select a shipping method' }
          }

          if (!state.selectedPaymentMethod) {
            return { success: false, error: 'Please select a payment method' }
          }

          if (!state.selectedPaymentGateway) {
            return { success: false, error: 'Payment gateway not configured' }
          }

          set({ isLoading: true })

          try {
            const data = await graphClient.request<PlaceOrderMutation>(
              PLACE_ORDER_MUTATION,
              {
                sessionId,
                input: {
                  idempotencyKey: crypto.randomUUID(),
                  addressId: state.selectedAddressId,
                  carrierId: state.selectedCarrierId,
                  paymentMethod: state.selectedPaymentMethod,
                  paymentGateway: state.selectedPaymentGateway,
                },
              },
            )

            const session = data.placeOrder

            // Update checkout session with new status
            set({
              checkoutSession: session,
              isLoading: false,
            })

            toast.success('Order placed successfully!')

            return {
              success: true,
              sessionId: session.id,
            }
          } catch (error) {
            set({ isLoading: false })
            const errorMessage =
              error instanceof Error ? error.message : 'Order placement failed'
            toast.error(errorMessage)

            return { success: false, error: errorMessage }
          }
        },

        // Utility
        getOrderSummary: (): OrderSummary => {
          const state = get()
          const shipping = state.selectedShippingOption?.price || 0

          // Note: Subtotal calculation requires product prices
          // This should be provided by the backend in the checkout session
          const subtotal = 0
          const total = subtotal + shipping

          return {
            subtotal,
            shipping,
            discount: 0,
            total,
          }
        },

        clearCheckout: () => {
          set({
            checkoutSession: null,
            isLoading: false,
            selectedAddressId: null,
            selectedCarrierId: null,
            selectedPaymentMethod: null,
            selectedPaymentGateway: null,
            selectedAddress: null,
            selectedShippingOption: null,
            selectedPaymentMethodData: null,
            selectedPaymentGatewayData: null,
            orderNote: '',
          })
        },
      }),
      {
        name: 'checkout-session-store',
        partialize: (state) => ({
          // Persist only the session ID and selections
          // Full session data should be refetched
          selectedAddressId: state.selectedAddressId,
          selectedCarrierId: state.selectedCarrierId,
          selectedPaymentMethod: state.selectedPaymentMethod,
          selectedPaymentGateway: state.selectedPaymentGateway,
        }),
      },
    ),
  ),
)

// Selectors for easier access
export const useCheckoutSession = () =>
  useCheckoutSessionStore((state) => state.checkoutSession)
export const useIsCheckoutLoading = () =>
  useCheckoutSessionStore((state) => state.isLoading)
export const useSelectedAddressId = () =>
  useCheckoutSessionStore((state) => state.selectedAddressId)
export const useSelectedAddress = () =>
  useCheckoutSessionStore((state) => state.selectedAddress)
export const useSelectedCarrierId = () =>
  useCheckoutSessionStore((state) => state.selectedCarrierId)
export const useSelectedShippingOption = () =>
  useCheckoutSessionStore((state) => state.selectedShippingOption)
export const useSelectedPaymentMethod = () =>
  useCheckoutSessionStore((state) => state.selectedPaymentMethod)
export const useSelectedPaymentGateway = () =>
  useCheckoutSessionStore((state) => state.selectedPaymentGateway)
export const useCheckoutOrderSummary = () =>
  useCheckoutSessionStore((state) => state.getOrderSummary())
