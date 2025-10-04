import { sendDeliveryReceipt, sendReadReceipt } from '@/lib/api'
import { useMutation } from '@tanstack/react-query'
import { useCallback } from 'react'

/**
 * Hook for managing message receipts
 */
export function useMessageReceipts(conversationId: string) {
  // Mutation for delivery receipts
  const deliveryReceiptMutation = useMutation({
    mutationFn: ({ messageId }: { messageId: string }) =>
      sendDeliveryReceipt(conversationId, messageId),
    onError: (error) => {
      console.error('Failed to send delivery receipt:', error)
    },
  })

  // Mutation for read receipts
  // Note: GraphQL Message type doesn't have delivery_status field
  // Message status updates are handled through WebSocket events
  const readReceiptMutation = useMutation({
    mutationFn: ({ messageId }: { messageId: string }) =>
      sendReadReceipt(conversationId, messageId),
    onError: (error) => {
      console.error('Failed to send read receipt:', error)
    },
  })

  /**
   * Send delivery receipt for a message
   */
  const markAsDelivered = useCallback(
    (messageId: string) => {
      deliveryReceiptMutation.mutate({ messageId })
    },
    [deliveryReceiptMutation],
  )

  /**
   * Send read receipt for a message
   */
  const markAsRead = useCallback(
    (messageId: string) => {
      readReceiptMutation.mutate({ messageId })
    },
    [readReceiptMutation],
  )

  /**
   * Mark multiple messages as read (for bulk operations)
   */
  const markMultipleAsRead = useCallback(
    (messageIds: Array<string>) => {
      messageIds.forEach((messageId) => {
        readReceiptMutation.mutate({ messageId })
      })
    },
    [readReceiptMutation],
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
    isMarkingDelivered: deliveryReceiptMutation.isPending,
    isMarkingRead: readReceiptMutation.isPending,
    markAsDelivered,
    markAsRead,
    markMultipleAsRead,
    updateMessageStatus,
  }
}
