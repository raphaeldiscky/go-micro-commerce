import { QUERY_KEY } from '@/constants/query-key'
import type { SendMessageRequest } from '@/lib/api'
import {
  CONVERSATION_MESSAGES_QUERY,
  SEND_MESSAGE_MUTATION,
  graphqlClient,
} from '@/lib/graphql'
import { generateTimestamp, generateUniqueId } from '@/lib/utils/date'
import type { Message, MessageConnection } from '@/types/__generated__/graphql'
import { ConversationStatus, MessageType } from '@/types/__generated__/graphql'
import {
  useInfiniteQuery,
  useMutation,
  useQueryClient,
} from '@tanstack/react-query'

interface ConversationMessagesQueryResponse {
  conversationMessages: MessageConnection
}

interface InfiniteQueryData {
  pages: Array<Array<Message>>
  pageParams: Array<string | undefined>
}

/**
 * Helper hook to add real-time message to cache
 */
export function useAddMessage(conversationId: string) {
  const queryClient = useQueryClient()

  return (message: Message) => {
    queryClient.setQueryData<InfiniteQueryData>(
      QUERY_KEY.chat.messages(conversationId),
      (old) => {
        if (!old) return old

        // Check if message already exists (prevent duplicates)
        const messageExists = old.pages.some((page) =>
          page.some((msg) => msg.id === message.id),
        )

        if (messageExists) return old

        // Add message to the last page
        const updatedPages = [...old.pages]
        if (updatedPages.length > 0) {
          updatedPages[updatedPages.length - 1] = [
            ...updatedPages[updatedPages.length - 1],
            message,
          ]
        } else {
          updatedPages.push([message])
        }

        return {
          ...old,
          pages: updatedPages,
        }
      },
    )

    // Update conversation list
    queryClient.invalidateQueries({ queryKey: QUERY_KEY.chat.conversations() })
  }
}

/**
 * Hook for fetching messages with cursor-based infinite scroll pagination
 */
export function useMessages(conversationId: string) {
  return useInfiniteQuery({
    enabled: !!conversationId,
    gcTime: 5 * 60 * 1000, // 5 minutes
    getNextPageParam: (lastPage: ConversationMessagesQueryResponse) => {
      const { pageInfo } = lastPage.conversationMessages
      return pageInfo.hasNextPage ? pageInfo.endCursor : undefined
    },
    initialPageParam: undefined as string | undefined,
    queryFn: async ({ pageParam }) => {
      const data =
        await graphqlClient.request<ConversationMessagesQueryResponse>(
          CONVERSATION_MESSAGES_QUERY,
          {
            conversationId,
            first: 50,
            after: pageParam,
          },
        )
      return data
    },
    queryKey: QUERY_KEY.chat.messages(conversationId),
    refetchOnWindowFocus: false, // Real-time updates handle this
    staleTime: 30 * 1000, // 30 seconds - messages are real-time
    // Note: Removed select function because it creates new array references on every render
    // causing infinite loops. Transform data in the component using useMemo instead.
  })
}

/**
 * Hook for sending messages via GraphQL mutation
 * Messages are sent through GraphQL and broadcasted via WebSocket subscriptions
 */
export function useSendMessage(conversationId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (message: SendMessageRequest) => {
      // Send message via GraphQL mutation
      await graphqlClient.request(SEND_MESSAGE_MUTATION, {
        input: {
          conversationId,
          content: message.content,
          messageType: message.message_type
            ? message.message_type.toUpperCase()
            : 'TEXT',
          replyToId: message.reply_to_id,
        },
      })

      // @TODO: Fix later
      // return result.sendMessage
    },
    onMutate: async (newMessage) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({
        queryKey: QUERY_KEY.chat.messages(conversationId),
      })

      // Snapshot the previous value
      const previousMessages = queryClient.getQueryData<InfiniteQueryData>([
        'messages',
        conversationId,
      ])

      // Optimistically update the cache
      queryClient.setQueryData<InfiniteQueryData>(
        QUERY_KEY.chat.messages(conversationId),
        (old) => {
          if (!old) return old

          const optimisticMessage: Message = {
            __typename: 'Message',
            content: newMessage.content,
            conversation: {
              __typename: 'Conversation',
              id: conversationId,
              subject: null,
              status: ConversationStatus.Active,
              priority: 0,
              participantCount: 0,
              participants: [],
              messages: {
                __typename: 'MessageConnection',
                edges: [],
                pageInfo: {
                  __typename: 'PageInfo',
                  hasNextPage: false,
                  hasPreviousPage: false,
                  startCursor: null,
                  endCursor: null,
                },
              },
              createdAt: '',
              updatedAt: '',
              endedAt: null,
            },
            conversationId,
            createdAt: generateTimestamp(),
            id: generateUniqueId('temp'),
            isSystem: false,
            messageType:
              newMessage.message_type === 'image'
                ? MessageType.Image
                : newMessage.message_type === 'file'
                  ? MessageType.File
                  : MessageType.Text,
            sender: null,
            senderId: 'current-user', // Will be replaced by real user ID
          }

          // Add message to the last page
          const updatedPages = [...old.pages]
          if (updatedPages.length > 0) {
            updatedPages[updatedPages.length - 1] = [
              ...updatedPages[updatedPages.length - 1],
              optimisticMessage,
            ]
          } else {
            updatedPages.push([optimisticMessage])
          }

          return {
            ...old,
            pages: updatedPages,
          }
        },
      )

      return { previousMessages }
    },
    onError: (_err, _newMessage, context) => {
      // Revert optimistic update on error
      if (context?.previousMessages) {
        queryClient.setQueryData<InfiniteQueryData>(
          QUERY_KEY.chat.messages(conversationId),
          context.previousMessages,
        )
      }
    },
    onSuccess: () => {
      // Note: We don't invalidate messages here because the GraphQL subscription
      // will receive the NewMessage event and handle the invalidation
      // This prevents double invalidations which can cause infinite loops

      // Update conversation list to reflect new message
      queryClient.invalidateQueries({
        queryKey: QUERY_KEY.chat.conversations(),
      })
    },
  })
}
