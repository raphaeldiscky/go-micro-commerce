import type { Message } from '@/lib/api'
import { sendDeliveryReceipt, sendReadReceipt } from '@/lib/api'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useCallback } from 'react'

/**
 * Hook for managing message receipts
 */
export function useMessageReceipts(conversationId: string) {
  const queryClient = useQueryClient()

  // Mutation for delivery receipts
  const deliveryReceiptMutation = useMutation({
    mutationFn: ({ messageId }: { messageId: string }) =>
      sendDeliveryReceipt(conversationId, messageId),
    onError: (error) => {
      console.error('Failed to send delivery receipt:', error)
    },
  })

  // Mutation for read receipts
  const readReceiptMutation = useMutation({
    mutationFn: ({ messageId }: { messageId: string }) =>
      sendReadReceipt(conversationId, messageId),
    onError: (error) => {
      console.error('Failed to send read receipt:', error)
    },
    onSuccess: (_, { messageId }) => {
      // Update message status in cache
      queryClient.setQueryData(['messages', conversationId], (old: any) => {
        if (!old) return old

        return {
          ...old,
          pages: old.pages.map((page: Array<Message>) =>
            page.map((msg) =>
              msg.id === messageId
                ? { ...msg, delivery_status: 'read' as const }
                : msg,
            ),
          ),
        }
      })
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
   */
  const updateMessageStatus = useCallback(
    (messageId: string, status: 'delivered' | 'read' | 'sent') => {
      queryClient.setQueryData(['messages', conversationId], (old: any) => {
        if (!old) return old

        return {
          ...old,
          pages: old.pages.map((page: Array<Message>) =>
            page.map((msg) =>
              msg.id === messageId ? { ...msg, delivery_status: status } : msg,
            ),
          ),
        }
      })
    },
    [queryClient, conversationId],
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
