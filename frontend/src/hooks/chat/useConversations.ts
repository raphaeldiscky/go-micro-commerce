import { QUERY_KEY } from '@/constants/query-key'
import {
  CONVERSATIONS_QUERY,
  JOIN_CONVERSATION_MUTATION,
  graphqlClient,
} from '@/lib/graphql'
import type {
  Conversation,
  Participant,
  ParticipantRole,
} from '@/types/__generated__/graphql'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

interface ConversationsQueryResponse {
  conversations: Array<Conversation>
}

interface JoinConversationMutationResponse {
  joinConversation: Participant
}

interface JoinConversationInput {
  conversationId: string
  role: ParticipantRole
}

/**
 * Hook for fetching conversations list
 */
export function useConversations() {
  return useQuery({
    gcTime: 5 * 60 * 1000, // 5 minutes
    queryFn: async () => {
      const data =
        await graphqlClient.request<ConversationsQueryResponse>(
          CONVERSATIONS_QUERY,
        )
      return data.conversations
    },
    queryKey: QUERY_KEY.chat.conversations(),
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
    mutationFn: async (input: JoinConversationInput) => {
      const data =
        await graphqlClient.request<JoinConversationMutationResponse>(
          JOIN_CONVERSATION_MUTATION,
          { input },
        )
      return data.joinConversation
    },
    onError: (error) => {
      console.error('Failed to join conversation:', error)
    },
    onSuccess: () => {
      // Refetch conversations list to update join status
      queryClient.invalidateQueries({
        queryKey: QUERY_KEY.chat.conversations(),
      })
    },
  })
}
