import { getConversations, joinConversation } from '@/lib/api'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

export const CONVERSATIONS_QUERY_KEY = 'conversations'

/**
 * Hook for fetching conversations list
 */
export function useConversations() {
  return useQuery({
    gcTime: 5 * 60 * 1000, // 5 minutes
    queryFn: getConversations,
    queryKey: [CONVERSATIONS_QUERY_KEY],
    refetchOnWindowFocus: false,
    retry: 3,
    staleTime: 30 * 1000, // 30 seconds - conversations don't change frequently
  })
}

/**
 * Hook for joining a conversation
 */
export function useJoinConversation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (conversationId: string) => joinConversation(conversationId),
    onError: (error) => {
      console.error('Failed to join conversation:', error)
    },
    onSuccess: () => {
      // Refetch conversations list to update join status
      queryClient.invalidateQueries({ queryKey: [CONVERSATIONS_QUERY_KEY] })
    },
  })
}
