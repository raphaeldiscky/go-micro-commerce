import { useChatWebSocket } from '@/contexts/ChatWebSocketContext'
import { useUser } from '@/hooks/auth/useAuth'
import { useCallback } from 'react'

/**
 * Hook for managing message receipts
 */
export function useMessageReceipts(conversationId: string) {
  const { sendMessage, isConnected } = useChatWebSocket()
  const user = useUser()

  /**
   * Send delivery receipt for a message
   */
  const markAsDelivered = useCallback(
    (messageId: string) => {
      if (!isConnected || !user) return

      sendMessage({
        type: 'delivery_receipt',
        content: {
          message_id: messageId,
          conversation_id: conversationId,
          recipient_id: user.id,
          delivered_at: Math.floor(Date.now() / 1000),
        },
      })
    },
    [sendMessage, isConnected, conversationId, user],
  )

  /**
   * Send read receipt for a message
   */
  const markAsRead = useCallback(
    (messageId: string) => {
      if (!isConnected || !user) return

      sendMessage({
        type: 'read_receipt',
        content: {
          message_id: messageId,
          conversation_id: conversationId,
          reader_id: user.id,
          read_at: Math.floor(Date.now() / 1000),
        },
      })
    },
    [sendMessage, isConnected, conversationId, user],
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
