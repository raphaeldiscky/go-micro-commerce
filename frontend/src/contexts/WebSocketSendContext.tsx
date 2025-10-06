import { env } from '@/env'
import { useIsAuthenticated } from '@/hooks/auth/useAuth'
import { getAccessToken } from '@/lib/api/client'
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from 'react'

interface WebSocketMessage {
  type: string
  content: unknown
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
  const wsRef = useRef<WebSocket | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const reconnectAttemptsRef = useRef(0)
  const maxReconnectAttempts = 5
  const isAuthenticated = useIsAuthenticated()

  const connect = useCallback(() => {
    const token = getAccessToken()
    if (!token) {
      return
    }

    // Don't create a new connection if one already exists
    if (
      wsRef.current &&
      (wsRef.current.readyState === WebSocket.CONNECTING ||
        wsRef.current.readyState === WebSocket.OPEN)
    ) {
      return
    }

    const wsUrl = env.VITE_API_GATEWAY_WS_URL

    const ws = new WebSocket(`${wsUrl}?token=${token}`)

    ws.onopen = () => {
      setIsConnected(true)
      reconnectAttemptsRef.current = 0 // Reset reconnect attempts on successful connection
    }

    ws.onclose = (event) => {
      setIsConnected(false)
      wsRef.current = null

      // Only reconnect if we haven't exceeded max attempts and have a token
      if (
        reconnectAttemptsRef.current < maxReconnectAttempts &&
        getAccessToken()
      ) {
        reconnectAttemptsRef.current += 1
        const delay = Math.min(1000 * 2 ** reconnectAttemptsRef.current, 30000) // Exponential backoff, max 30s

        reconnectTimeoutRef.current = setTimeout(() => {
          connect()
        }, delay)
      } else if (reconnectAttemptsRef.current >= maxReconnectAttempts) {
        console.error(
          'Max WebSocket reconnection attempts reached. Please refresh the page.',
        )
      }
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }

    wsRef.current = ws
  }, []) // Empty dependencies - token is fetched inside the function

  useEffect(() => {
    // Only connect once when component mounts
    connect()

    return () => {
      // Clear any pending reconnection attempts
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      // Close the WebSocket connection
      if (wsRef.current) {
        wsRef.current.close()
        wsRef.current = null
      }
    }
  }, [])

  // Handle authentication state changes
  useEffect(() => {
    if (!isAuthenticated) {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
        reconnectTimeoutRef.current = null
      }
      // Close WebSocket
      if (wsRef.current) {
        wsRef.current.close()
        wsRef.current = null
      }
      setIsConnected(false)
      reconnectAttemptsRef.current = 0
    } else if (!wsRef.current) {
      connect()
    }
  }, [isAuthenticated, connect])

  const sendMessage = useCallback((message: WebSocketMessage) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      console.warn('WebSocket not connected, cannot send message')
      return
    }

    try {
      wsRef.current.send(JSON.stringify(message))
    } catch (error) {
      console.error('Failed to send WebSocket message:', error)
    }
  }, [])

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
