import React, { useCallback, useEffect, useRef, useState } from 'react'
import type { ChatMessage } from './ChatMessage'
import { ChatMessageComponent } from './ChatMessage'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
import { Card, CardContent, CardHeader, CardTitle } from './ui/card'
import { Input } from './ui/input'

interface User {
  id: string
  name: string
  ticket: string
}

interface UserChatPanelProps {
  user: User
  onRemove?: () => void
}

type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error'

export function UserChatPanel({ user, onRemove }: UserChatPanelProps) {
  const [messages, setMessages] = useState<Array<ChatMessage>>([])
  const [inputMessage, setInputMessage] = useState('')
  const [connectionStatus, setConnectionStatus] =
    useState<ConnectionStatus>('disconnected')
  const [ws, setWs] = useState<WebSocket | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const reconnectAttemptsRef = useRef(0)

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

    setConnectionStatus('connecting')
    const websocket = new WebSocket(
      `ws://192.168.0.107:9088/v1/ws?ticket=${user.ticket}`,
    )

    websocket.onopen = () => {
      console.log(`WebSocket connected for ${user.name}`)
      setConnectionStatus('connected')
      setWs(websocket)
      reconnectAttemptsRef.current = 0
    }

    websocket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        const newMessage: ChatMessage = {
          id: `${Date.now()}-${Math.random()}`,
          content: data.message || data.content || event.data,
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
          id: `${Date.now()}-${Math.random()}`,
          content: event.data,
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
        `WebSocket closed for ${user.name}:`,
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
            `Attempting to reconnect ${user.name} (attempt ${reconnectAttemptsRef.current})`,
          )
          connectWebSocket()
        }, timeout)
      }
    }

    websocket.onerror = (error) => {
      console.error(`WebSocket error for ${user.name}:`, error)
      setConnectionStatus('error')
    }

    return websocket
  }, [user.id, user.name, user.ticket, ws])

  useEffect(() => {
    const websocket = connectWebSocket()

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }

      if (websocket && websocket.readyState === WebSocket.OPEN) {
        websocket.close()
      }
    }
  }, [connectWebSocket])

  const sendMessage = useCallback(() => {
    if (!inputMessage.trim() || !ws || ws.readyState !== WebSocket.OPEN) {
      return
    }

    const messageData = {
      message: inputMessage.trim(),
      senderId: user.id,
      senderName: user.name,
      timestamp: Date.now(),
    }

    try {
      ws.send(JSON.stringify(messageData))

      // Add message to local state for immediate feedback
      const newMessage: ChatMessage = {
        id: `${Date.now()}-${Math.random()}`,
        content: inputMessage.trim(),
        senderId: user.id,
        senderName: user.name,
        timestamp: new Date(),
        type: 'sent',
      }

      setMessages((prev) => [...prev, newMessage])
      setInputMessage('')
    } catch (error) {
      console.error(`Error sending message for ${user.name}:`, error)
    }
  }, [inputMessage, user.id, user.name, ws])

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
        return 'Error'
      default:
        return 'Disconnected'
    }
  }

  return (
    <Card className="w-full max-w-md mx-auto h-[500px] flex flex-col">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg">{user.name}</CardTitle>
          <div className="flex items-center gap-2">
            <Badge variant={getStatusBadgeVariant()}>{getStatusText()}</Badge>
            {onRemove && (
              <Button
                variant="outline"
                size="sm"
                onClick={onRemove}
                className="h-6 w-6 p-0"
              >
                ×
              </Button>
            )}
          </div>
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
                key={message.id}
                message={message}
                currentUserId={user.id}
              />
            ))
          )}
          <div ref={messagesEndRef} />
        </div>

        {/* Input area */}
        <div className="flex gap-2">
          <Input
            value={inputMessage}
            onChange={(e) => setInputMessage(e.target.value)}
            onKeyDown={handleKeyPress}
            placeholder="Type a message..."
            disabled={connectionStatus !== 'connected'}
            className="flex-1"
          />
          <Button
            onClick={sendMessage}
            disabled={connectionStatus !== 'connected' || !inputMessage.trim()}
            size="sm"
          >
            Send
          </Button>
        </div>

        {connectionStatus !== 'connected' && (
          <div className="text-xs text-gray-500 dark:text-gray-400 mt-1">
            {connectionStatus === 'error'
              ? 'Connection failed. Check WebSocket server.'
              : connectionStatus === 'connecting'
                ? 'Establishing connection...'
                : 'Click to reconnect'}
            {connectionStatus === 'disconnected' && (
              <Button
                variant="link"
                size="sm"
                onClick={connectWebSocket}
                className="h-auto p-0 ml-2 text-xs"
              >
                Reconnect
              </Button>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
