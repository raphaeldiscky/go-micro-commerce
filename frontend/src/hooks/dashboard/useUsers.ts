import { QUERY_KEY } from '@/constants/query-key'
import { paginateWithCursor } from '@/lib/mock-data/pagination'
import type { UserRole } from '@/lib/mock-data/users'
import { mockUsers } from '@/lib/mock-data/users'
import { useQuery } from '@tanstack/react-query'

interface UserFilters {
  searchQuery: string
  role: UserRole | 'all'
  cursor?: string
}

/**
 * Hook for fetching dashboard users with cursor-based pagination
 * Uses cursor-based pagination for better performance with large datasets
 */
export function useUsers(filters: UserFilters, limit = 10) {
  return useQuery({
    queryKey: QUERY_KEY.dashboard.users({
      searchQuery: filters.searchQuery,
      role: filters.role,
      cursor: filters.cursor,
    }),
    queryFn: () => {
      // Filter users based on search and role
      const filteredUsers = mockUsers.filter((user) => {
        const matchesSearch =
          user.firstName
            .toLowerCase()
            .includes(filters.searchQuery.toLowerCase()) ||
          user.lastName
            .toLowerCase()
            .includes(filters.searchQuery.toLowerCase()) ||
          user.email.toLowerCase().includes(filters.searchQuery.toLowerCase())

        const matchesRole = filters.role === 'all' || user.role === filters.role

        return matchesSearch && matchesRole
      })

      // Paginate results
      return paginateWithCursor(filteredUsers, limit, filters.cursor)
    },
    gcTime: 5 * 60 * 1000, // 5 minutes
    staleTime: 30 * 1000, // 30 seconds
    refetchOnWindowFocus: false,
  })
}
