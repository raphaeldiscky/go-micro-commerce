import type { TypingIndicator } from '@/lib/api'
import { sendTypingIndicator } from '@/lib/api'
import { useMutation } from '@tanstack/react-query'
import { useCallback, useEffect, useRef, useState } from 'react'

/**
 * Hook for managing typing indicators
 */
export function useTypingIndicator(conversationId: string) {
  const [typingUsers, setTypingUsers] = useState<Array<TypingIndicator>>([])
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null)

  const sendTypingMutation = useMutation({
    mutationFn: (isTyping: boolean) =>
      sendTypingIndicator(conversationId, isTyping),
    onError: (error) => {
      console.error('Failed to send typing indicator:', error)
    },
  })

  /**
   * Start typing indicator
   */
  const startTyping = useCallback(() => {
    sendTypingMutation.mutate(true)

    // Stop typing after 3 seconds of inactivity
    typingTimeoutRef.current = setTimeout(() => {
      sendTypingMutation.mutate(false)
    }, 3000)
  }, [sendTypingMutation])

  /**
   * Stop typing indicator
   */
  const stopTyping = useCallback(() => {
    sendTypingMutation.mutate(false)
  }, [sendTypingMutation])

  /**
   * Add typing user from WebSocket
   */
  const addTypingUser = useCallback((user: TypingIndicator) => {
    setTypingUsers((prev) => {
      const filtered = prev.filter((u) => u.user_id !== user.user_id)
      return user.is_typing ? [...filtered, user] : filtered
    })

    // Auto-remove typing indicator after 5 seconds
    if (user.is_typing) {
      setTimeout(() => {
        setTypingUsers((prev) => prev.filter((u) => u.user_id !== user.user_id))
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
    isLoading: sendTypingMutation.isPending,
    startTyping,
    stopTyping,
    typingUsers,
  }
}
