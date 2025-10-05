import { useChatWebSocket } from '@/contexts/ChatWebSocketContext'
import type { TypingIndicator } from '@/lib/api'
import { useCallback, useEffect, useRef, useState } from 'react'

/**
 * Hook for managing typing indicators
 */
export function useTypingIndicator(_conversationId: string) {
  const [typingUsers, setTypingUsers] = useState<Array<TypingIndicator>>([])
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const { sendMessage, isConnected } = useChatWebSocket()

  /**
   * Start typing indicator
   */
  const startTyping = useCallback(() => {
    if (!isConnected) return

    sendMessage({
      type: 'typing',
      content: { is_typing: true },
    })

    // Stop typing after 3 seconds of inactivity
    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current)
    }

    typingTimeoutRef.current = setTimeout(() => {
      sendMessage({
        type: 'typing',
        content: { is_typing: false },
      })
    }, 3000)
  }, [sendMessage, isConnected])

  /**
   * Stop typing indicator
   */
  const stopTyping = useCallback(() => {
    if (!isConnected) return

    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current)
      typingTimeoutRef.current = null
    }

    sendMessage({
      type: 'typing',
      content: { is_typing: false },
    })
  }, [sendMessage, isConnected])

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

  // Cleanup on unmount - no dependency needed for cleanup functions
  useEffect(() => {
    return () => {
      stopTyping()
    }
  }, [])

  return {
    addTypingUser,
    clearTypingUsers,
    startTyping,
    stopTyping,
    typingUsers,
  }
}
