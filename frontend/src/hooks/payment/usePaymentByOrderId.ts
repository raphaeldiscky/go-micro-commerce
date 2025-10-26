import { QUERY_KEY } from '@/constants/query-key'
import { graphClient } from '@/lib/graphql/client'
import { GET_PAYMENT_BY_ORDER_ID } from '@/lib/graphql/payment'
import type {
  GetPaymentByOrderIdQuery,
  GetPaymentByOrderIdQueryVariables,
} from '@/lib/graphql/payment.generated'
import { useQuery } from '@tanstack/react-query'

/**
 * Hook to fetch payment details by order ID
 * @param orderId - The order ID to fetch payment for
 * @param options - TanStack Query options
 * @returns Payment query result with payment details including client secret for Stripe
 */
export function usePaymentByOrderId(
  orderId: string,
  options?: {
    enabled?: boolean
    refetchInterval?: number | false
  },
) {
  return useQuery({
    queryKey: QUERY_KEY.order.paymentByOrderId(orderId),
    queryFn: async () => {
      const data = await graphClient.request<
        GetPaymentByOrderIdQuery,
        GetPaymentByOrderIdQueryVariables
      >(GET_PAYMENT_BY_ORDER_ID, {
        orderId,
      })
      return data.getPaymentByOrderId
    },
    enabled: options?.enabled,
    refetchInterval: options?.refetchInterval,
  })
}
