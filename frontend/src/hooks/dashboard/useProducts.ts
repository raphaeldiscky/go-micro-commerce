import { QUERY_KEY } from '@/constants/query-key'
import { paginateWithCursor } from '@/mocks/pagination'
import type { ProductStatus } from '@/mocks/products'
import { mockProducts } from '@/mocks/products'
import { useQuery } from '@tanstack/react-query'

interface ProductFilters {
  searchQuery: string
  status: ProductStatus | 'all'
  cursor?: string
}

/**
 * Hook for fetching dashboard products with cursor-based pagination
 * Uses cursor-based pagination for better performance with large datasets
 */
export function useProducts(filters: ProductFilters, limit = 10) {
  return useQuery({
    queryKey: QUERY_KEY.dashboard.products({
      searchQuery: filters.searchQuery,
      status: filters.status,
      cursor: filters.cursor,
    }),
    queryFn: () => {
      // Filter products based on search and status
      const filteredProducts = mockProducts.filter((product) => {
        const matchesSearch =
          product.name
            .toLowerCase()
            .includes(filters.searchQuery.toLowerCase()) ||
          product.sku.toLowerCase().includes(filters.searchQuery.toLowerCase())

        const matchesStatus =
          filters.status === 'all' || product.status === filters.status

        return matchesSearch && matchesStatus
      })

      // Paginate results
      return paginateWithCursor(filteredProducts, limit, filters.cursor)
    },
    gcTime: 5 * 60 * 1000, // 5 minutes
    staleTime: 30 * 1000, // 30 seconds
    refetchOnWindowFocus: false,
  })
}
