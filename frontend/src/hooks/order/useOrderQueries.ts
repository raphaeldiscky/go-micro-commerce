import { QUERY_KEY } from '@/constants/query-key'
import { graphClient } from '@/lib/graphql/client'
import { LIST_MY_ORDERS_QUERY } from '@/lib/graphql/order'
import type { ListMyOrdersQuery } from '@/lib/graphql/order.generated'
import { useQuery } from '@tanstack/react-query'

/**
 * Hook to fetch authenticated user's orders with cursor pagination
 */
export function useMyOrders(limit: number = 20, cursor?: string) {
  return useQuery({
    queryKey: QUERY_KEY.order.list(limit, cursor),
    queryFn: async () => {
      const data = await graphClient.request<ListMyOrdersQuery>(
        LIST_MY_ORDERS_QUERY,
        { limit, cursor },
      )
      return data.listMyOrders
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
  })
}
