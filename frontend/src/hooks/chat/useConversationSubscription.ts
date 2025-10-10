import { QUERY_KEY } from '@/constants/query-key'
import { CONVERSATION_EVENTS_SUBSCRIPTION } from '@/lib/graphql/chat'
import type { ConversationEventsSubscription } from '@/lib/graphql/chat.generated'
import { getSubscriptionClient } from '@/lib/graphql/subscription-client'
import type { Message, MessageConnection } from '@/types/__generated__/graphql'
import { ConversationStatus } from '@/types/__generated__/graphql'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'

interface ConversationMessagesQueryResponse {
  conversationMessages: MessageConnection
}

interface InfiniteQueryData {
  pages: Array<ConversationMessagesQueryResponse>
  pageParams: Array<string | undefined>
}

export function useConversationSubscription(conversationId: string) {
  const queryClient = useQueryClient()
  const unsubscribeRef = useRef<(() => void) | null>(null)

  useEffect(() => {
    if (!conversationId) return

    console.log('🔔 Initializing GraphQL subscription', {
      conversationId,
      timestamp: new Date().toISOString(),
    })

    const client = getSubscriptionClient()

    unsubscribeRef.current = client.subscribe<ConversationEventsSubscription>(
      {
        query: CONVERSATION_EVENTS_SUBSCRIPTION,
        variables: { conversationId },
      },
      {
        next: (data) => {
          console.log('📨 Subscription event received', {
            conversationId,
            event: data.data?.conversationEvents,
            timestamp: new Date().toISOString(),
          })

          if (!data.data?.conversationEvents) return

          const event = data.data.conversationEvents

          switch (event.__typename) {
            case 'NewMessage': {
              console.log('🔄 Processing NewMessage event', {
                messageId: event.id,
                content: event.content,
                conversationId,
              })

              const message: Message = {
                __typename: 'Message',
                id: event.id,
                conversationId: event.conversationId,
                senderId: event.senderId,
                content: event.content,
                messageType: event.messageType,
                isSystem: event.isSystem,
                createdAt: event.createdAt,
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
                sender: null,
              }

              const queryKey = QUERY_KEY.chat.messages(conversationId)
              console.log('🔑 Query key for cache update:', queryKey)

              queryClient.setQueryData<InfiniteQueryData>(queryKey, (old) => {
                console.log('📦 Current cache state:', {
                  hasOldData: !!old,
                  pageCount: old?.pages.length ?? 0,
                  totalMessages:
                    old?.pages.reduce(
                      (sum, page) =>
                        sum + page.conversationMessages.edges.length,
                      0,
                    ) ?? 0,
                })

                if (!old) {
                  console.warn('⚠️  No existing query data, cannot add message')
                  return old
                }

                const messageExists = old.pages.some((page) =>
                  page.conversationMessages.edges.some(
                    (edge) => edge.node.id === message.id,
                  ),
                )

                if (messageExists) {
                  console.log('ℹ️  Message already exists in cache, skipping')
                  return old
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
                          node: message,
                          cursor: message.id,
                        },
                      ],
                    },
                  }
                } else {
                  console.log('📝 Creating first page with new message')
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

                const newData = {
                  ...old,
                  pages: updatedPages,
                }

                console.log('✅ Cache updated successfully', {
                  newPageCount: newData.pages.length,
                  newTotalMessages: newData.pages.reduce(
                    (sum, page) => sum + page.conversationMessages.edges.length,
                    0,
                  ),
                })

                return newData
              })

              queryClient.invalidateQueries({
                queryKey: QUERY_KEY.chat.conversations(),
              })
              break
            }

            case 'TypingIndicator':
              queryClient.setQueryData(
                QUERY_KEY.chat.typingIndicator(conversationId, event.userId),
                event.isTyping,
              )
              break

            case 'DeliveryReceipt':
            case 'ReadReceipt':
              break
          }
        },
        error: (error) => {
          console.error('❌ GraphQL subscription error:', {
            conversationId,
            error,
            timestamp: new Date().toISOString(),
          })
        },
        complete: () => {
          console.log('✅ GraphQL subscription completed', {
            conversationId,
            timestamp: new Date().toISOString(),
          })
        },
      },
    )

    return () => {
      if (unsubscribeRef.current) {
        unsubscribeRef.current()
        unsubscribeRef.current = null
      }
    }
  }, [conversationId])
}
