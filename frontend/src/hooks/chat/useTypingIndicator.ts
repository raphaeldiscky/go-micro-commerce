import { graphClient, SEND_TYPING_INDICATOR_MUTATION } from '@/lib/graphql'
import type { TypingIndicator } from '@/types/__generated__/graphql'
import { useMutation } from '@tanstack/react-query'
import { useCallback, useEffect, useRef, useState } from 'react'

/**
 * Hook for managing typing indicators via GraphQL
 */
export function useTypingIndicator(conversationId: string) {
  const [typingUsers, setTypingUsers] = useState<Array<TypingIndicator>>([])
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null)

  // Mutation for sending typing indicators
  const sendTypingMutation = useMutation({
    mutationFn: async (isTyping: boolean) => {
      return graphClient.request(SEND_TYPING_INDICATOR_MUTATION, {
        input: {
          conversationId,
          isTyping,
        },
      })
    },
  })

  /**
   * Start typing indicator
   */
  const startTyping = useCallback(() => {
    sendTypingMutation.mutate(true)

    // Stop typing after 3 seconds of inactivity
    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current)
    }

    typingTimeoutRef.current = setTimeout(() => {
      sendTypingMutation.mutate(false)
    }, 3000)
    // sendTypingMutation is stable from useMutation
  }, [])

  /**
   * Stop typing indicator
   */
  const stopTyping = useCallback(() => {
    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current)
      typingTimeoutRef.current = null
    }

    sendTypingMutation.mutate(false)
    // sendTypingMutation is stable from useMutation
  }, [])

  /**
   * Add typing user from GraphQL subscription
   */
  const addTypingUser = useCallback((user: TypingIndicator) => {
    setTypingUsers((prev) => {
      const filtered = prev.filter((u) => u.userId !== user.userId)
      return user.isTyping ? [...filtered, user] : filtered
    })

    // Auto-remove typing indicator after 5 seconds
    if (user.isTyping) {
      setTimeout(() => {
        setTypingUsers((prev) => prev.filter((u) => u.userId !== user.userId))
      }, 5000)
    }
  }, [])

  /**
   * Clear all typing indicators
   */
  const clearTypingUsers = useCallback(() => {
    setTypingUsers([])
  }, [])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      stopTyping()
    }
  }, [stopTyping])

  return {
    addTypingUser,
    clearTypingUsers,
    startTyping,
    stopTyping,
    typingUsers,
  }
}
