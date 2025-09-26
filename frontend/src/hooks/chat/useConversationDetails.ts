import { getConversationDetails, getConversationParticipants } from '@/lib/api'
import { useQuery } from '@tanstack/react-query'

/**
 * Hook for fetching conversation details
 */
export function useConversationDetails(conversationId: string) {
  return useQuery({
    enabled: !!conversationId,
    gcTime: 10 * 60 * 1000, // 10 minutes
    queryFn: () => getConversationDetails(conversationId),
    queryKey: ['conversation-details', conversationId],
    retry: 3,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

/**
 * Hook for fetching conversation participants
 */
export function useConversationParticipants(conversationId: string) {
  return useQuery({
    enabled: !!conversationId,
    gcTime: 5 * 60 * 1000, // 5 minutes
    queryFn: () => getConversationParticipants(conversationId),
    queryKey: ['conversation-participants', conversationId],
    refetchOnWindowFocus: true, // Refetch to get latest online status
    retry: 3,
    staleTime: 2 * 60 * 1000, // 2 minutes - participants change more frequently
  })
}
