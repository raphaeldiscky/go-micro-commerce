import { env } from '@/env'
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

  const connect = useCallback(() => {
    const token = getAccessToken()
    if (!token) {
      console.warn('No auth token available for WebSocket connection')
      return
    }

    // Create WebSocket connection to the old /v1/ws endpoint for typing/presence
    // This is separate from the GraphQL subscription WebSocket
    const wsUrl = env.VITE_GRAPHQL_GATEWAY_WS_URL

    const ws = new WebSocket(`${wsUrl}?token=${token}`)

    ws.onopen = () => {
      console.log('WebSocket connected for typing/presence')
      setIsConnected(true)
    }

    ws.onclose = () => {
      console.log('WebSocket disconnected')
      setIsConnected(false)
      wsRef.current = null

      // Attempt to reconnect after 3 seconds
      reconnectTimeoutRef.current = setTimeout(() => {
        console.log('Attempting to reconnect WebSocket...')
        connect()
      }, 3000)
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }

    wsRef.current = ws
  }, [])

  useEffect(() => {
    connect()

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [connect])

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
