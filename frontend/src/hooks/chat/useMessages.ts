import { QUERY_KEY } from '@/constants/query-key'
import {
  CONVERSATION_MESSAGES_QUERY,
  SEND_MESSAGE_MUTATION,
  graphqlClient,
} from '@/lib/graphql'
import { generateTimestamp, generateUniqueId } from '@/lib/utils/date'
import type {
  Message,
  MessageConnection,
  SendMessageInput,
} from '@/types/__generated__/graphql'
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
  pages: Array<ConversationMessagesQueryResponse>
  pageParams: Array<string | undefined>
}

export function useAddMessage(conversationId: string) {
  const queryClient = useQueryClient()

  return (message: Message) => {
    queryClient.setQueryData<InfiniteQueryData>(
      QUERY_KEY.chat.messages(conversationId),
      (old) => {
        if (!old) return old

        const messageExists = old.pages.some((page) =>
          page.conversationMessages.edges.some(
            (edge) => edge.node.id === message.id,
          ),
        )

        if (messageExists) return old

        const updatedPages = [...old.pages]
        if (updatedPages.length > 0) {
          const lastPage = updatedPages[updatedPages.length - 1]
          updatedPages[updatedPages.length - 1] = {
            ...lastPage,
            conversationMessages: {
              ...lastPage.conversationMessages,
              edges: [
                ...lastPage.conversationMessages.edges,
                {
                  __typename: 'MessageEdge',
                  node: message,
                  cursor: message.id,
                },
              ],
            },
          }
        } else {
          updatedPages.push({
            conversationMessages: {
              __typename: 'MessageConnection',
              edges: [
                {
                  __typename: 'MessageEdge',
                  node: message,
                  cursor: message.id,
                },
              ],
              pageInfo: {
                __typename: 'PageInfo',
                hasNextPage: false,
                hasPreviousPage: false,
                startCursor: message.id,
                endCursor: message.id,
              },
            },
          })
        }

        return {
          ...old,
          pages: updatedPages,
        }
      },
    )

    queryClient.invalidateQueries({ queryKey: QUERY_KEY.chat.conversations() })
  }
}

export function useMessages(conversationId: string) {
  return useInfiniteQuery({
    enabled: !!conversationId,
    gcTime: 5 * 60 * 1000,
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
    refetchOnWindowFocus: false,
    staleTime: 30 * 1000,
  })
}

export function useSendMessage(conversationId: string, currentUserId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (input: SendMessageInput) => {
      await graphqlClient.request(SEND_MESSAGE_MUTATION, {
        input: {
          ...input,
          conversationId,
        },
      })
    },
    onMutate: async (newMessage) => {
      await queryClient.cancelQueries({
        queryKey: QUERY_KEY.chat.messages(conversationId),
      })

      const previousMessages = queryClient.getQueryData<InfiniteQueryData>([
        'messages',
        conversationId,
      ])

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
            messageType: newMessage.messageType || MessageType.Text,
            sender: null,
            senderId: currentUserId,
          }

          const updatedPages = [...old.pages]
          if (updatedPages.length > 0) {
            const lastPage = updatedPages[updatedPages.length - 1]
            updatedPages[updatedPages.length - 1] = {
              ...lastPage,
              conversationMessages: {
                ...lastPage.conversationMessages,
                edges: [
                  ...lastPage.conversationMessages.edges,
                  {
                    __typename: 'MessageEdge',
                    node: optimisticMessage,
                    cursor: optimisticMessage.id,
                  },
                ],
              },
            }
          } else {
            updatedPages.push({
              conversationMessages: {
                __typename: 'MessageConnection',
                edges: [
                  {
                    __typename: 'MessageEdge',
                    node: optimisticMessage,
                    cursor: optimisticMessage.id,
                  },
                ],
                pageInfo: {
                  __typename: 'PageInfo',
                  hasNextPage: false,
                  hasPreviousPage: false,
                  startCursor: optimisticMessage.id,
                  endCursor: optimisticMessage.id,
                },
              },
            })
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
      if (context?.previousMessages) {
        queryClient.setQueryData<InfiniteQueryData>(
          QUERY_KEY.chat.messages(conversationId),
          context.previousMessages,
        )
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: QUERY_KEY.chat.conversations(),
      })
    },
  })
}
