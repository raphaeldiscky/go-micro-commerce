import { useChatTicket } from '@/hooks/chat/useChatTicket'
import { useIsAuthenticated, useUser } from '@/hooks/useAuth'
import { MessageCircle } from 'lucide-react'
import React, { useCallback, useEffect, useRef, useState } from 'react'
import { Badge } from '../../ui/badge'
import { Button } from '../../ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '../../ui/card'
import { Input } from '../../ui/input'
import type { ChatMessage } from './ChatMessage'
import { ChatMessageComponent } from './ChatMessage'

interface ChatInterfaceProps {
  conversationId: string
  conversationName: string
}

type ConnectionStatus =
  | 'connected'
  | 'connecting'
  | 'disconnected'
  | 'error'
  | 'fetching_ticket'

export function ChatInterface({
  conversationId,
  conversationName,
}: ChatInterfaceProps) {
  const [messages, setMessages] = useState<Array<ChatMessage>>([])
  const [inputMessage, setInputMessage] = useState('')
  const [connectionStatus, setConnectionStatus] =
    useState<ConnectionStatus>('disconnected')
  const [ws, setWs] = useState<null | WebSocket>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const reconnectAttemptsRef = useRef(0)

  const isAuthenticated = useIsAuthenticated()
  const user = useUser()

  // Use TanStack Query to fetch chat ticket with complete data
  const {
    data: ticketData,
    error: ticketError,
    isLoading: isLoadingTicket,
    refetch: refetchTicket,
  } = useChatTicket(user?.id || '')

  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [])

  useEffect(() => {
    scrollToBottom()
  }, [messages, scrollToBottom])

  const connectWebSocket = useCallback(() => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      return
    }

    // Don't try to connect if we don't have ticket data yet or user is not authenticated
    if (!ticketData || !user || !isAuthenticated) {
      return
    }

    // Check if ticket is expired
    const expiresAt = new Date(ticketData.expires_at)
    if (expiresAt <= new Date()) {
      console.log(
        `Ticket expired for conversation ${conversationId}, refetching...`,
      )
      refetchTicket()
      return
    }

    setConnectionStatus('connecting')
    const websocketUrl = `${ticketData.node_address}/v1/ws?ticket=${ticketData.ticket}&conversation_id=${conversationId}`
    console.log(
      `Connecting to WebSocket for conversation ${conversationId}:`,
      websocketUrl,
    )

    const websocket = new WebSocket(websocketUrl)

    websocket.onopen = () => {
      console.log(`WebSocket connected for conversation ${conversationId}`)
      setConnectionStatus('connected')
      setWs(websocket)
      reconnectAttemptsRef.current = 0
    }

    websocket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        const newMessage: ChatMessage = {
          content: data.message || data.content || event.data,
          id: `${Date.now()}-${Math.random()}`,
          senderId: data.senderId || data.userId || 'unknown',
          senderName:
            data.senderName ||
            data.userName ||
            `User ${data.senderId || 'Unknown'}`,
          timestamp: new Date(data.timestamp || Date.now()),
          type: data.senderId === user.id ? 'sent' : 'received',
        }

        setMessages((prev) => [...prev, newMessage])
      } catch (error) {
        // If it's not JSON, treat as plain text message
        const newMessage: ChatMessage = {
          content: event.data,
          id: `${Date.now()}-${Math.random()}`,
          senderId: 'unknown',
          senderName: 'Unknown User',
          timestamp: new Date(),
          type: 'received',
        }

        setMessages((prev) => [...prev, newMessage])
      }
    }

    websocket.onclose = (event) => {
      console.log(
        `WebSocket closed for conversation ${conversationId}:`,
        event.code,
        event.reason,
      )
      setConnectionStatus('disconnected')
      setWs(null)

      // Attempt reconnection with exponential backoff
      if (reconnectAttemptsRef.current < 5) {
        const timeout = Math.pow(2, reconnectAttemptsRef.current) * 1000
        reconnectAttemptsRef.current += 1

        reconnectTimeoutRef.current = setTimeout(() => {
          console.log(
            `Attempting to reconnect conversation ${conversationId} (attempt ${reconnectAttemptsRef.current})`,
          )
          connectWebSocket()
        }, timeout)
      }
    }

    websocket.onerror = (error) => {
      console.error(
        `WebSocket error for conversation ${conversationId}:`,
        error,
      )
      setConnectionStatus('error')
    }

    return websocket
  }, [conversationId, ticketData, refetchTicket, user, isAuthenticated, ws])

  useEffect(() => {
    // Set connection status based on ticket loading state
    if (isLoadingTicket) {
      setConnectionStatus('fetching_ticket')
    } else if (ticketError) {
      setConnectionStatus('error')
    } else if (ticketData && connectionStatus === 'fetching_ticket') {
      setConnectionStatus('disconnected') // Ready to connect
    }
  }, [isLoadingTicket, ticketError, ticketData, connectionStatus])

  useEffect(() => {
    // Only attempt to connect if we have ticket data and user is authenticated
    if (!ticketData || !user || !isAuthenticated) {
      return
    }

    // Clear previous messages when switching conversations
    setMessages([])

    const websocket = connectWebSocket()

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }

      if (websocket && websocket.readyState === WebSocket.OPEN) {
        websocket.close()
      }
    }
  }, [connectWebSocket, ticketData, conversationId, user, isAuthenticated])

  const sendMessage = useCallback(() => {
    if (
      !inputMessage.trim() ||
      !ws ||
      ws.readyState !== WebSocket.OPEN ||
      !user
    ) {
      return
    }

    const messageData = {
      conversationId: conversationId,
      message: inputMessage.trim(),
      senderId: user.id,
      senderName:
        `${user.first_name} ${user.last_name}`.trim() || user.username,
      timestamp: Date.now(),
    }

    try {
      ws.send(JSON.stringify(messageData))

      // Add message to local state for immediate feedback
      const newMessage: ChatMessage = {
        content: inputMessage.trim(),
        id: `${Date.now()}-${Math.random()}`,
        senderId: user.id,
        senderName:
          `${user.first_name} ${user.last_name}`.trim() || user.username,
        timestamp: new Date(),
        type: 'sent',
      }

      setMessages((prev) => [...prev, newMessage])
      setInputMessage('')
    } catch (error) {
      console.error(
        `Error sending message for conversation ${conversationId}:`,
        error,
      )
    }
  }, [inputMessage, user, conversationId, ws])

  const handleKeyPress = useCallback(
    (event: React.KeyboardEvent) => {
      if (event.key === 'Enter' && !event.shiftKey) {
        event.preventDefault()
        sendMessage()
      }
    },
    [sendMessage],
  )

  const getStatusBadgeVariant = () => {
    switch (connectionStatus) {
      case 'connected':
        return 'success'
      case 'connecting':
      case 'fetching_ticket':
        return 'warning'
      case 'error':
        return 'destructive'
      default:
        return 'secondary'
    }
  }

  const getStatusText = () => {
    switch (connectionStatus) {
      case 'connected':
        return 'Connected'
      case 'connecting':
        return 'Connecting...'
      case 'error':
        return ticketError ? 'Ticket error' : 'Connection error'
      case 'fetching_ticket':
        return 'Getting ticket...'
      default:
        return 'Disconnected'
    }
  }

  if (!isAuthenticated || !user) {
    return (
      <Card className="flex-1 h-full flex flex-col">
        <CardContent className="flex-1 flex items-center justify-center">
          <div className="text-center text-muted-foreground">
            <MessageCircle className="h-12 w-12 mx-auto mb-2 opacity-50" />
            <p>Please log in to use chat</p>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className="flex-1 h-full flex flex-col">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg flex items-center gap-2">
            <MessageCircle className="h-5 w-5" />
            {conversationName}
          </CardTitle>
          <Badge variant={getStatusBadgeVariant()}>{getStatusText()}</Badge>
        </div>
      </CardHeader>

      <CardContent className="flex-1 flex flex-col p-4 pt-0">
        {/* Messages area */}
        <div className="flex-1 overflow-y-auto mb-4 border rounded-lg p-3 bg-gray-50 dark:bg-gray-900">
          {messages.length === 0 ? (
            <div className="text-center text-gray-500 dark:text-gray-400 text-sm">
              No messages yet. Start typing to begin the conversation!
            </div>
          ) : (
            messages.map((message) => (
              <ChatMessageComponent
                currentUserId={user.id}
                key={message.id}
                message={message}
              />
            ))
          )}
          <div ref={messagesEndRef} />
        </div>

        {/* Input area */}
        <div className="flex gap-2">
          <Input
            className="flex-1"
            disabled={connectionStatus !== 'connected'}
            onChange={(e) => setInputMessage(e.target.value)}
            onKeyDown={handleKeyPress}
            placeholder="Type a message..."
            value={inputMessage}
          />
          <Button
            disabled={connectionStatus !== 'connected' || !inputMessage.trim()}
            onClick={sendMessage}
            size="sm"
          >
            Send
          </Button>
        </div>

        {connectionStatus !== 'connected' && (
          <div className="text-xs text-gray-500 dark:text-gray-400 mt-1">
            {connectionStatus === 'error' && ticketError
              ? `Failed to get ticket: ${ticketError.message}`
              : connectionStatus === 'error'
                ? 'Connection failed. Check WebSocket server.'
                : connectionStatus === 'connecting'
                  ? 'Establishing connection...'
                  : connectionStatus === 'fetching_ticket'
                    ? 'Getting authentication ticket...'
                    : 'Click to reconnect'}
            {(connectionStatus === 'disconnected' ||
              (connectionStatus === 'error' && !ticketError)) && (
              <Button
                className="h-auto p-0 ml-2 text-xs"
                onClick={connectWebSocket}
                size="sm"
                variant="link"
              >
                Reconnect
              </Button>
            )}
            {connectionStatus === 'error' && ticketError && (
              <Button
                className="h-auto p-0 ml-2 text-xs"
                onClick={() => refetchTicket()}
                size="sm"
                variant="link"
              >
                Retry
              </Button>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
