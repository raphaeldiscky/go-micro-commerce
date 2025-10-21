import type {
  CreateCheckoutSessionRequest,
  CreateCheckoutSessionResponse,
} from '@/types/order'
import { useMutation } from '@tanstack/react-query'

/**
 * Hook to create Stripe checkout session
 * Creates a new checkout session on-demand when user clicks "Pay Now"
 */
export function useCreateCheckoutSession() {
  return useMutation({
    mutationFn: async (
      input: CreateCheckoutSessionRequest,
    ): Promise<CreateCheckoutSessionResponse> => {
      // For now, return mock response until backend GraphQL schema is ready
      // When backend is ready, uncomment the line below and remove mock response
      // const data = await graphClient.request<CreateCheckoutSessionMutation>(
      //   CREATE_CHECKOUT_SESSION_MUTATION,
      //   { input }
      // )
      // return data.createCheckoutSession

      // Mock response for development
      return new Promise((resolve) => {
        setTimeout(() => {
          resolve({
            sessionId: `cs_test_${Date.now()}_${Math.random().toString(36).substring(7)}`,
            checkoutUrl: `https://checkout.stripe.com/pay/cs_test_mock_${input.paymentId}`,
          })
        }, 800)
      })
    },
  })
}
