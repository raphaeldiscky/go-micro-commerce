import { useUser } from '@/hooks/auth'
import {
  SEND_DELIVERY_RECEIPT_MUTATION,
  SEND_READ_RECEIPT_MUTATION,
  graphClient,
} from '@/lib/graphql'
import { useCallback } from 'react'

/**
 * Hook for managing message receipts via GraphQL mutations
 */
export function useMessageReceipts(conversationId: string) {
  const user = useUser()

  /**
   * Send delivery receipt for a message via GraphQL
   */
  const markAsDelivered = useCallback(
    async (messageId: string) => {
      if (!user) return

      try {
        await graphClient.request(SEND_DELIVERY_RECEIPT_MUTATION, {
          input: {
            messageId,
            conversationId,
          },
        })
      } catch (error) {
        console.error('Failed to send delivery receipt:', error)
      }
    },
    [conversationId, user],
  )

  /**
   * Send read receipt for a message via GraphQL
   */
  const markAsRead = useCallback(
    async (messageId: string) => {
      if (!user) return

      try {
        await graphClient.request(SEND_READ_RECEIPT_MUTATION, {
          input: {
            messageId,
            conversationId,
          },
        })
      } catch (error) {
        console.error('Failed to send read receipt:', error)
      }
    },
    [conversationId, user],
  )

  /**
   * Mark multiple messages as read (for bulk operations)
   */
  const markMultipleAsRead = useCallback(
    (messageIds: Array<string>) => {
      messageIds.forEach((messageId) => {
        markAsRead(messageId)
      })
    },
    [markAsRead],
  )

  /**
   * Update message delivery status from WebSocket events
   * Note: GraphQL Message type doesn't include delivery_status field
   * Status updates would need to be handled differently if this field is added to the schema
   */
  const updateMessageStatus = useCallback(
    (_messageId: string, _status: 'delivered' | 'read' | 'sent') => {
      // GraphQL Message type doesn't have delivery_status field
      // This function is kept for API compatibility but currently does nothing
    },
    [],
  )

  return {
    markAsDelivered,
    markAsRead,
    markMultipleAsRead,
    updateMessageStatus,
  }
}
