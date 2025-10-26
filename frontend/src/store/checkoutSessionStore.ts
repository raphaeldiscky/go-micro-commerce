import {
  CANCEL_CHECKOUT_SESSION_MUTATION,
  CREATE_CHECKOUT_SESSION_MUTATION,
  GET_CHECKOUT_SESSION_QUERY,
  PLACE_ORDER_MUTATION,
  UPDATE_CHECKOUT_SESSION_MUTATION,
} from '@/lib/graphql/checkout'
import type {
  CancelCheckoutSessionMutation,
  CreateCheckoutSessionMutation,
  GetCheckoutSessionQuery,
  PlaceOrderMutation,
  UpdateCheckoutSessionMutation,
} from '@/lib/graphql/checkout.generated'
import { graphClient } from '@/lib/graphql/client'
import type { Address, CheckoutSession } from '@/types/__generated__/graphql'
import type {
  OrderSummary,
  PaymentGatewayUI,
  ShippingOptionUI,
} from '@/types/cart'
import { toast } from 'sonner'
import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import { useShallow } from 'zustand/react/shallow'

// Checkout session store state interface
export interface CheckoutSessionState {
  checkoutSession: CheckoutSession | null
  isLoading: boolean
  selectedAddressId: string | null
  selectedCarrierId: string | null
  selectedPaymentGateway: string | null
  orderNote?: string

  // UI-specific selections for display purposes
  selectedAddress: Address | null
  selectedShippingOption: ShippingOptionUI | null
  selectedPaymentGatewayData: PaymentGatewayUI | null
}

// Checkout session store actions interface
export interface CheckoutSessionActions {
  // Checkout session management
  fetchCheckoutSession: (sessionId: string) => Promise<void>
  createCheckoutSession: (
    cartId: string,
    navigateToCheckout: (checkoutId: string) => void,
  ) => Promise<void>
  cancelCheckout: (sessionId: string) => Promise<void>

  // Set selections for checkout (auto-saves to backend)
  setAddress: (addressId: string, address: Address) => Promise<void>
  setShippingMethod: (
    carrierId: string,
    method: ShippingOptionUI,
  ) => Promise<void>
  setPaymentGateway: (
    gateway: string,
    gatewayData: PaymentGatewayUI,
  ) => Promise<void>
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
        selectedPaymentGateway: null,
        selectedAddress: null,
        selectedShippingOption: null,
        selectedPaymentGatewayData: null,
        orderNote: '',

        // Checkout session management
        fetchCheckoutSession: async (sessionId: string) => {
          set({ isLoading: true })

          try {
            const data = await graphClient.request<GetCheckoutSessionQuery>(
              GET_CHECKOUT_SESSION_QUERY,
              { id: sessionId },
            )
            const session = data.getCheckoutSession
            set({
              checkoutSession: session,
              selectedAddressId: session?.addressId,
              selectedCarrierId: session?.carrierId,
              selectedPaymentGateway: session?.paymentGateway,
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

        createCheckoutSession: async (
          cartId: string,
          navigateToCheckout: (checkoutId: string) => void,
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
                    cartId,
                  },
                },
              )

            const session = data.createCheckoutSession

            set({
              checkoutSession: session,
              isLoading: false,
              selectedAddressId: null,
              selectedCarrierId: null,
              selectedPaymentGateway: null,
            })

            toast.success('Proceeding to checkout')

            // Navigate to checkout page with ID if navigation function provided
            if (session.id) {
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

        // Set selections for checkout (auto-saves to backend)
        setAddress: async (addressId: string, address: Address) => {
          const sessionId = get().checkoutSession?.id

          // Update local state immediately for responsive UI
          set({
            selectedAddressId: addressId,
            selectedAddress: address,
          })

          // Persist to backend
          if (sessionId) {
            try {
              await graphClient.request<UpdateCheckoutSessionMutation>(
                UPDATE_CHECKOUT_SESSION_MUTATION,
                {
                  sessionId,
                  input: { addressId },
                },
              )
            } catch (error) {
              console.error(
                'Failed to update address in checkout session:',
                error,
              )
              // Silent failure - user selection is saved locally
              // Backend will be synced when placing order
            }
          }
        },

        setShippingMethod: async (
          carrierId: string,
          method: ShippingOptionUI,
        ) => {
          const sessionId = get().checkoutSession?.id

          // Update local state immediately for responsive UI
          set({
            selectedCarrierId: carrierId,
            selectedShippingOption: method,
          })

          // Persist to backend
          if (sessionId) {
            try {
              await graphClient.request<UpdateCheckoutSessionMutation>(
                UPDATE_CHECKOUT_SESSION_MUTATION,
                {
                  sessionId,
                  input: { carrierId },
                },
              )
            } catch (error) {
              console.error(
                'Failed to update carrier in checkout session:',
                error,
              )
              // Silent failure - user selection is saved locally
              // Backend will be synced when placing order
            }
          }
        },

        setPaymentGateway: async (
          gateway: string,
          gatewayData: PaymentGatewayUI,
        ) => {
          const sessionId = get().checkoutSession?.id

          // Update local state immediately for responsive UI
          set({
            selectedPaymentGateway: gateway,
            selectedPaymentGatewayData: gatewayData,
          })

          // Persist to backend
          if (sessionId) {
            try {
              await graphClient.request<UpdateCheckoutSessionMutation>(
                UPDATE_CHECKOUT_SESSION_MUTATION,
                {
                  sessionId,
                  input: { paymentGateway: gateway },
                },
              )
            } catch (error) {
              console.error(
                'Failed to update payment gateway in checkout session:',
                error,
              )
              // Silent failure - user selection is saved locally
              // Backend will be synced when placing order
            }
          }
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

          const items = state.checkoutSession?.items || []

          const subtotal = items.reduce(
            (total, item) => total + Number(item.unitPrice) * item.quantity,
            0,
          )
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
            selectedPaymentGateway: null,
            selectedAddress: null,
            selectedShippingOption: null,
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
export const useSelectedPaymentGateway = () =>
  useCheckoutSessionStore((state) => state.selectedPaymentGateway)
export const useCheckoutOrderSummary = () =>
  useCheckoutSessionStore(useShallow((state) => state.getOrderSummary()))
