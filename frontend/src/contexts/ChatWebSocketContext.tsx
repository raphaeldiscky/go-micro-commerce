import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from 'react'

interface WebSocketMessage {
  id?: string
  type: string
  content: unknown
  timestamp?: string
}

interface ChatWebSocketContextValue {
  isConnected: boolean
  sendMessage: (message: WebSocketMessage) => void
  connectionStatus: 'connected' | 'connecting' | 'disconnected' | 'error'
}

const ChatWebSocketContext = createContext<
  ChatWebSocketContextValue | undefined
>(undefined)

interface ChatWebSocketProviderProps {
  children: React.ReactNode
  websocket: WebSocket | null
  connectionStatus: 'connected' | 'connecting' | 'disconnected' | 'error'
}

export function ChatWebSocketProvider({
  children,
  websocket,
  connectionStatus,
}: ChatWebSocketProviderProps) {
  const [isConnected, setIsConnected] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    wsRef.current = websocket
    setIsConnected(connectionStatus === 'connected')
  }, [websocket, connectionStatus])

  const sendMessage = useCallback((message: WebSocketMessage) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      console.error('WebSocket is not connected')
      return
    }

    // Add ID and timestamp if not provided
    const messageWithDefaults: WebSocketMessage = {
      id: message.id || crypto.randomUUID(),
      type: message.type,
      content: message.content,
      timestamp: message.timestamp || new Date().toISOString(),
    }

    try {
      wsRef.current.send(JSON.stringify(messageWithDefaults))
    } catch (error) {
      console.error('Failed to send WebSocket message:', error)
    }
  }, [])

  const value: ChatWebSocketContextValue = {
    isConnected,
    sendMessage,
    connectionStatus,
  }

  return (
    <ChatWebSocketContext.Provider value={value}>
      {children}
    </ChatWebSocketContext.Provider>
  )
}

export function useChatWebSocket() {
  const context = useContext(ChatWebSocketContext)
  if (!context) {
    throw new Error(
      'useChatWebSocket must be used within ChatWebSocketProvider',
    )
  }
  return context
}
