import { getChatTicket } from '@/lib/api'
import { useQuery } from '@tanstack/react-query'

export const CHAT_TICKET_QUERY_KEY = 'chatTicket'

/**
 * Custom hook for fetching chat tickets using TanStack Query
 * @param userId - The user ID to get a ticket for
 * @returns TanStack Query result with complete ticket data including node_address, loading state, and error handling
 */
export function useChatTicket(userId: string) {
  return useQuery({
    gcTime: 10 * 60 * 1000, // 10 minutes - keep in cache for potential reuse
    queryFn: () => getChatTicket(userId),
    queryKey: [CHAT_TICKET_QUERY_KEY, userId],
    refetchOnReconnect: true, // Refetch when network reconnects
    refetchOnWindowFocus: false, // Don't refetch when window regains focus
    retry: 3,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    staleTime: 5 * 60 * 1000, // 5 minutes - tickets don't expire quickly
  })
}
