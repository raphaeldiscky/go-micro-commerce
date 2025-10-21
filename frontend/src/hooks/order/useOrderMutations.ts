import { QUERY_KEY } from '@/constants/query-key'
import type { PlaceOrderRequest, PlaceOrderResponse } from '@/types/order'
import { useMutation, useQueryClient } from '@tanstack/react-query'

/**
 * Hook to create an order
 */
export function useCreateOrder() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (
      input: PlaceOrderRequest,
    ): Promise<PlaceOrderResponse> => {
      // For now, return mock response until backend GraphQL schema is ready
      // When backend is ready, uncomment the line below and remove mock response
      // const data = await graphClient.request<CreateOrderMutation>(CREATE_ORDER_MUTATION, { input })
      // return data.createOrder

      // Mock response for development
      return new Promise((resolve) => {
        setTimeout(() => {
          const now = new Date()
          const paymentDeadline = new Date(now.getTime() + 24 * 60 * 60 * 1000) // 24 hours from now

          resolve({
            orderId: `order-${Date.now()}-${Math.random().toString(36).substring(7)}`,
            paymentId: `payment-${Date.now()}-${Math.random().toString(36).substring(7)}`,
            status: 'payment_pending',
            paymentStatus: 'pending',
            paymentGateway: input.paymentGateway,
            totalAmount: input.items.reduce(
              (total, item) => total + item.quantity * 100,
              0,
            ), // Mock calculation
            currency: input.currency,
            paymentDeadline: paymentDeadline.toISOString(),
            createdAt: now.toISOString(),
          })
        }, 1500)
      })
    },
    onSuccess: (data) => {
      // Invalidate cart and checkout queries after successful order creation
      queryClient.invalidateQueries({ queryKey: QUERY_KEY.cart.all })
      queryClient.invalidateQueries({ queryKey: QUERY_KEY.checkout.all })

      // Set the order data in cache for immediate access
      queryClient.setQueryData(QUERY_KEY.order.detail(data.orderId), data)
    },
  })
}
