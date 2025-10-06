import { QUERY_KEY } from '@/constants/query-key'
import { CONVERSATION_EVENTS_SUBSCRIPTION } from '@/lib/graphql/chat'
import { getSubscriptionClient } from '@/lib/graphql/subscription-client'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'

type NewMessage = {
  __typename: 'NewMessage'
  id: string
  conversationId: string
  senderId: string
  content: string
  messageType: string
  isSystem: boolean
  createdAt: string
}

type TypingIndicator = {
  __typename: 'TypingIndicator'
  userId: string
  conversationId: string
  isTyping: boolean
  timestamp: string
}

type DeliveryReceipt = {
  __typename: 'DeliveryReceipt'
  messageId: string
  conversationId: string
  recipientId: string
  deliveredAt: string
}

type ReadReceipt = {
  __typename: 'ReadReceipt'
  messageId: string
  conversationId: string
  readerId: string
  readAt: string
}

type ConversationEvent =
  | NewMessage
  | TypingIndicator
  | DeliveryReceipt
  | ReadReceipt

interface ConversationEventsData {
  conversationEvents: ConversationEvent
}

/**
 * Hook to subscribe to real-time conversation events via GraphQL subscriptions
 * Handles messages, typing indicators, and message receipts
 */
export function useConversationSubscription(conversationId: string) {
  const queryClient = useQueryClient()
  const unsubscribeRef = useRef<(() => void) | null>(null)

  useEffect(() => {
    if (!conversationId) return

    const client = getSubscriptionClient()

    // Subscribe to conversation events
    unsubscribeRef.current = client.subscribe<ConversationEventsData>(
      {
        query: CONVERSATION_EVENTS_SUBSCRIPTION,
        variables: { conversationId },
      },
      {
        next: (data) => {
          if (!data.data?.conversationEvents) return

          const event = data.data.conversationEvents

          // Handle different event types
          switch (event.__typename) {
            case 'NewMessage':
              // Invalidate messages query to refetch and show new message
              queryClient.invalidateQueries({
                queryKey: QUERY_KEY.chat.messages(conversationId),
              })
              break

            case 'TypingIndicator':
              // Update typing indicator state
              queryClient.setQueryData(
                QUERY_KEY.chat.typingIndicator(conversationId, event.userId),
                event.isTyping,
              )
              break

            case 'DeliveryReceipt':
            case 'ReadReceipt':
              // Invalidate messages to update receipt status
              queryClient.invalidateQueries({
                queryKey: QUERY_KEY.chat.messages(conversationId),
              })
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

    // Cleanup on unmount
    return () => {
      if (unsubscribeRef.current) {
        unsubscribeRef.current()
        unsubscribeRef.current = null
      }
    }
  }, [conversationId, queryClient])
}
