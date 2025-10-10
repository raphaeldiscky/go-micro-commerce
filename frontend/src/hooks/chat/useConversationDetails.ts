import { QUERY_KEY } from '@/constants/query-key'
import {
  CONVERSATION_PARTICIPANTS_QUERY,
  CONVERSATION_QUERY,
  graphClient,
} from '@/lib/graphql'
import type { Conversation, Participant } from '@/types/__generated__/graphql'
import { useQuery } from '@tanstack/react-query'

interface ConversationQueryResponse {
  conversation: Conversation | null
}

interface ConversationParticipantsQueryResponse {
  conversationParticipants: Array<Participant>
}

/**
 * Hook for fetching conversation details
 */
export function useConversationDetails(conversationId: string) {
  return useQuery({
    enabled: !!conversationId,
    gcTime: 10 * 60 * 1000, // 10 minutes
    queryFn: async () => {
      const data = await graphClient.request<ConversationQueryResponse>(
        CONVERSATION_QUERY,
        { id: conversationId },
      )
      return data.conversation
    },
    queryKey: QUERY_KEY.chat.conversationDetails(conversationId),
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
    queryFn: async () => {
      const data =
        await graphClient.request<ConversationParticipantsQueryResponse>(
          CONVERSATION_PARTICIPANTS_QUERY,
          { conversationId },
        )
      return data.conversationParticipants
    },
    queryKey: QUERY_KEY.chat.conversationParticipants(conversationId),
    refetchOnWindowFocus: true, // Refetch to get latest online status
    retry: 3,
    staleTime: 2 * 60 * 1000, // 2 minutes - participants change more frequently
  })
}
