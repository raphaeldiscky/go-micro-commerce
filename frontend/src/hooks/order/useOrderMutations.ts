import { QUERY_KEY } from '@/constants/query-key'
import { graphClient } from '@/lib/graphql/client'
import { CREATE_ORDER_MUTATION } from '@/lib/graphql/order'
import type {
  CreateOrderInput,
  CreateOrderMutation,
} from '@/lib/graphql/order.generated'
import { useMutation, useQueryClient } from '@tanstack/react-query'

/**
 * Hook to create an order using GraphQL
 */
export function useCreateOrder() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (input: CreateOrderInput) => {
      const data = await graphClient.request<CreateOrderMutation>(
        CREATE_ORDER_MUTATION,
        { input },
      )

      return data.createOrder
    },
    onSuccess: (data) => {
      // Invalidate cart and checkout queries after successful order creation
      queryClient.invalidateQueries({ queryKey: QUERY_KEY.cart.all })
      queryClient.invalidateQueries({ queryKey: QUERY_KEY.checkout.all })
      queryClient.invalidateQueries({ queryKey: QUERY_KEY.order.lists() })

      // Set the order data in cache for immediate access
      queryClient.setQueryData(QUERY_KEY.order.detail(data.id), data)
    },
  })
}
