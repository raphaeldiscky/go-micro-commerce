import { QUERY_KEY } from '@/constants/query-key'
import { CONVERSATION_EVENTS_SUBSCRIPTION } from '@/lib/graphql/chat'
import type { ConversationEventsSubscription } from '@/lib/graphql/chat.generated'
import { getSubscriptionClient } from '@/lib/graphql/subscription-client'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'

export function useConversationSubscription(conversationId: string) {
  const queryClient = useQueryClient()
  const unsubscribeRef = useRef<(() => void) | null>(null)
  const invalidationTimeoutRef = useRef<NodeJS.Timeout | null>(null)

  useEffect(() => {
    if (!conversationId) return

    const client = getSubscriptionClient()

    unsubscribeRef.current = client.subscribe<ConversationEventsSubscription>(
      {
        query: CONVERSATION_EVENTS_SUBSCRIPTION,
        variables: { conversationId },
      },
      {
        next: (data) => {
          if (!data.data?.conversationEvents) return

          const event = data.data.conversationEvents

          switch (event.__typename) {
            case 'NewMessage':
              if (invalidationTimeoutRef.current) {
                clearTimeout(invalidationTimeoutRef.current)
              }
              invalidationTimeoutRef.current = setTimeout(() => {
                queryClient.invalidateQueries({
                  queryKey: QUERY_KEY.chat.messages(conversationId),
                })
                invalidationTimeoutRef.current = null
              }, 100)
              break

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
          console.error('GraphQL subscription error:', error)
        },
        complete: () => {
          console.log('GraphQL subscription completed')
        },
      },
    )

    return () => {
      if (invalidationTimeoutRef.current) {
        clearTimeout(invalidationTimeoutRef.current)
        invalidationTimeoutRef.current = null
      }
      if (unsubscribeRef.current) {
        unsubscribeRef.current()
        unsubscribeRef.current = null
      }
    }
  }, [conversationId])
}
