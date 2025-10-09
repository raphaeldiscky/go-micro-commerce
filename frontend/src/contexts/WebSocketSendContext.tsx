import { env } from '@/env'
import { useIsAuthenticated } from '@/hooks/auth/useAuth'
import { getAccessToken } from '@/lib/api/client'
import { createContext, useCallback, useContext, useMemo } from 'react'
import useWebSocket, { ReadyState } from 'react-use-websocket'

interface WebSocketMessage {
  id?: string
  type: string
  content: unknown
  timestamp?: string
  sender_id?: string
  channel?: string
}

interface WebSocketSendContextValue {
  sendMessage: (message: WebSocketMessage) => void
  isConnected: boolean
}

const WebSocketSendContext = createContext<
  WebSocketSendContextValue | undefined
>(undefined)

export function WebSocketSendProvider({
  children,
}: {
  children: React.ReactNode
}) {
  const isAuthenticated = useIsAuthenticated()
  const token = getAccessToken()

  // Build WebSocket URL with token as query parameter
  const socketUrl = useMemo(() => {
    if (!token) return null
    return `${env.VITE_CHAT_WEBSOCKET_URL}?token=${token}`
  }, [token])

  const { sendJsonMessage, readyState } = useWebSocket(
    socketUrl,
    {
      // Share WebSocket instance across all components using this context
      // This prevents double-mounting issues in React StrictMode
      share: true,

      // Only connect when authenticated and have a token
      shouldReconnect: () => isAuthenticated && !!token,

      // Reconnection configuration
      reconnectAttempts: 5,
      reconnectInterval: 3000, // 3 seconds

      // Connection lifecycle callbacks
      onOpen: () => {
        console.log('WebSocket connected')
      },
      onClose: () => {
        console.log('WebSocket disconnected')
      },
      onError: (error) => {
        console.error('WebSocket error:', error)
      },

      // Retry on error
      retryOnError: true,
    },
    // Only enable WebSocket when authenticated and have token
    isAuthenticated && !!token,
  )

  // Map readyState to simple boolean
  const isConnected = readyState === ReadyState.OPEN

  // Wrap sendJsonMessage to format messages according to backend expectations
  const sendMessage = useCallback(
    (message: WebSocketMessage) => {
      if (readyState !== ReadyState.OPEN) {
        console.warn('WebSocket not connected, cannot send message')
        return
      }

      // Ensure message has required fields matching backend format
      const completeMessage: WebSocketMessage = {
        id: message.id || crypto.randomUUID(),
        type: message.type,
        content: message.content,
        timestamp: message.timestamp || new Date().toISOString(),
        ...(message.sender_id && { sender_id: message.sender_id }),
        ...(message.channel && { channel: message.channel }),
      }

      try {
        sendJsonMessage(completeMessage)
      } catch (error) {
        console.error('Failed to send WebSocket message:', error)
      }
    },
    [sendJsonMessage, readyState],
  )

  return (
    <WebSocketSendContext.Provider value={{ sendMessage, isConnected }}>
      {children}
    </WebSocketSendContext.Provider>
  )
}

export function useWebSocketSend() {
  const context = useContext(WebSocketSendContext)
  if (!context) {
    throw new Error(
      'useWebSocketSend must be used within WebSocketSendProvider',
    )
  }
  return context
}
