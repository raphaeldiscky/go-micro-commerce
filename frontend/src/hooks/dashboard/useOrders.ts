import { QUERY_KEY } from '@/constants/query-key'
import type { OrderStatus } from '@/lib/mock-data/orders'
import { mockOrders } from '@/lib/mock-data/orders'
import { paginateWithCursor } from '@/lib/mock-data/pagination'
import { useQuery } from '@tanstack/react-query'

interface OrderFilters {
  searchQuery: string
  status: OrderStatus | 'all'
  cursor?: string
}

/**
 * Hook for fetching dashboard orders with cursor-based pagination
 * Uses cursor-based pagination for better performance with large datasets
 */
export function useOrders(filters: OrderFilters, limit = 10) {
  return useQuery({
    queryKey: QUERY_KEY.dashboard.orders({
      searchQuery: filters.searchQuery,
      status: filters.status,
      cursor: filters.cursor,
    }),
    queryFn: () => {
      // Filter orders based on search and status
      const filteredOrders = mockOrders.filter((order) => {
        const matchesSearch =
          order.id.toLowerCase().includes(filters.searchQuery.toLowerCase()) ||
          order.customerName
            .toLowerCase()
            .includes(filters.searchQuery.toLowerCase()) ||
          order.customerEmail
            .toLowerCase()
            .includes(filters.searchQuery.toLowerCase())

        const matchesStatus =
          filters.status === 'all' || order.status === filters.status

        return matchesSearch && matchesStatus
      })

      // Paginate results
      return paginateWithCursor(filteredOrders, limit, filters.cursor)
    },
    gcTime: 5 * 60 * 1000, // 5 minutes
    staleTime: 30 * 1000, // 30 seconds
    refetchOnWindowFocus: false,
  })
}
