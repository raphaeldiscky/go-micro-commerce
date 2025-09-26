import type { Message, SendMessageRequest } from '@/lib/api'
import { getConversationMessages, sendMessage } from '@/lib/api'
import {
  useInfiniteQuery,
  useMutation,
  useQueryClient,
} from '@tanstack/react-query'

/**
 * Helper hook to add real-time message to cache
 */
export function useAddMessage(conversationId: string) {
  const queryClient = useQueryClient()

  return (message: Message) => {
    queryClient.setQueryData(['messages', conversationId], (old: any) => {
      if (!old) return old

      // Check if message already exists (prevent duplicates)
      const messageExists = old.pages.some((page: Array<Message>) =>
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
    })

    // Update conversation list
    queryClient.invalidateQueries({ queryKey: ['conversations'] })
  }
}

/**
 * Hook for fetching messages with infinite scroll pagination
 */
export function useMessages(conversationId: string) {
  return useInfiniteQuery({
    enabled: !!conversationId,
    gcTime: 5 * 60 * 1000, // 5 minutes
    getNextPageParam: (lastPage: { messages: Array<Message>; hasMore: boolean; totalPages: number }, allPages) => {
      // Use the hasMore flag from the API response
      return lastPage.hasMore ? allPages.length + 1 : undefined
    },
    initialPageParam: 1,
    queryFn: ({ pageParam = 1 }) =>
      getConversationMessages(conversationId, pageParam, 50),
    queryKey: ['messages', conversationId],
    refetchOnWindowFocus: false, // Real-time updates handle this
    staleTime: 30 * 1000, // 30 seconds - messages are real-time
    select: (data) => ({
      ...data,
      pages: data.pages.map((page: { messages: Array<Message>; hasMore: boolean; totalPages: number }) => page.messages),
    }),
  })
}

/**
 * Hook for sending messages
 */
export function useSendMessage(conversationId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (message: SendMessageRequest) =>
      sendMessage(conversationId, message),
    onMutate: async (newMessage) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({
        queryKey: ['messages', conversationId],
      })

      // Snapshot the previous value
      const previousMessages = queryClient.getQueryData([
        'messages',
        conversationId,
      ])

      // Optimistically update the cache
      queryClient.setQueryData(['messages', conversationId], (old: any) => {
        if (!old) return old

        const optimisticMessage: Message = {
          content: newMessage.content,
          conversation_id: conversationId,
          created_at: new Date().toISOString(),
          delivery_status: 'sent',
          id: `temp-${Date.now()}`,
          message_type: newMessage.message_type || 'text',
          sender_id: 'current-user', // Will be replaced by real user ID
          sender_name: 'You',
          updated_at: new Date().toISOString(),
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
      })

      return { previousMessages }
    },
    onError: (_err, _newMessage, context) => {
      // Revert optimistic update on error
      if (context?.previousMessages) {
        queryClient.setQueryData(
          ['messages', conversationId],
          context.previousMessages,
        )
      }
    },
    onSuccess: (data) => {
      // Replace optimistic message with real message
      queryClient.setQueryData(['messages', conversationId], (old: any) => {
        if (!old) return old

        return {
          ...old,
          pages: old.pages.map((page: Array<Message>, index: number) => {
            if (index === old.pages.length - 1) {
              // Replace the last optimistic message with the real one
              return page.map((msg) =>
                msg.id.startsWith('temp-') ? data : msg,
              )
            }
            return page
          }),
        }
      })

      // Update conversation list to reflect new message
      queryClient.invalidateQueries({ queryKey: ['conversations'] })
    },
  })
}
