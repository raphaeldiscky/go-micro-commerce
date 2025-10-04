import { PATH_AUTH, PATH_DASHBOARD } from '@/constants'
import { useIsAuthenticated, useUser } from '@/hooks/auth/useAuth'
import { useChatTicket } from '@/hooks/chat/useChatTicket'
import { useConversationDetails } from '@/hooks/chat/useConversationDetails'
import { useMessageReceipts } from '@/hooks/chat/useMessageReceipts'
import { useSendMessage } from '@/hooks/chat/useMessages'
import { usePresence } from '@/hooks/chat/usePresence'
import { useTypingIndicator } from '@/hooks/chat/useTypingIndicator'
import type { Message, SendMessageRequest } from '@/lib/api'
import { isExpired } from '@/lib/utils/date'
import { Link, useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  Maximize2,
  Minimize2,
  Phone,
  Settings,
  Users,
  Video,
} from 'lucide-react'
import { useCallback, useEffect, useRef, useState } from 'react'
import { Button } from '../../../ui/button'
import { Card } from '../../../ui/card'
import { MessageInput } from '../input/MessageInput'
import { MessageList } from '../lists/MessageList'
import { ParticipantsList } from '../participants/ParticipantsList'

interface ConversationPageProps {
  conversationId: string
  isFullscreen?: boolean
  onToggleFullscreen?: () => void
  showToggle?: boolean
}

export function ConversationPage({
  conversationId,
  isFullscreen = false,
  onToggleFullscreen,
  showToggle = true,
}: ConversationPageProps) {
  const [showParticipants, setShowParticipants] = useState(false)
  const [replyingTo, setReplyingTo] = useState<Message | null>(null)
  const [_ws, setWs] = useState<null | WebSocket>(null)
  const [connectionStatus, setConnectionStatus] = useState<
    'connected' | 'connecting' | 'disconnected' | 'error'
  >('disconnected')

  const navigate = useNavigate()
  const user = useUser()
  const isAuthenticated = useIsAuthenticated()

  // Hooks
  const { data: conversation, isLoading: isLoadingConversation } =
    useConversationDetails(conversationId)
  const sendMessageMutation = useSendMessage(conversationId)
  const { addTypingUser, startTyping, stopTyping, typingUsers } =
    useTypingIndicator(conversationId)
  const { addOnlineUser, isUserOnline, removeOnlineUser } = usePresence()
  const { updateMessageStatus } = useMessageReceipts(conversationId)

  // Chat ticket for WebSocket connection
  const {
    data: ticketData,
    isLoading: isLoadingTicket,
    refetch: refetchTicket,
  } = useChatTicket(user?.id || '')

  const wsRef = useRef<null | WebSocket>(null)

  // Stable references for WebSocket event handlers
  const handlersRef = useRef({
    addTypingUser,
    addOnlineUser,
    removeOnlineUser,
    updateMessageStatus,
  })

  // Update handlers ref when they change
  useEffect(() => {
    handlersRef.current = {
      addTypingUser,
      addOnlineUser,
      removeOnlineUser,
      updateMessageStatus,
    }
  }, [addTypingUser, addOnlineUser, removeOnlineUser, updateMessageStatus])

  // WebSocket connection management
  const connectWebSocket = useCallback(() => {
    if (!ticketData || !user || !isAuthenticated) return

    // Check if ticket is expired
    if (isExpired(ticketData.expires_at)) {
      console.log('Ticket expired, refetching...')
      refetchTicket()
      return
    }

    // Close existing connection if any
    if (wsRef.current && wsRef.current.readyState !== WebSocket.CLOSED) {
      wsRef.current.close()
    }

    setConnectionStatus('connecting')
    const websocketUrl = `${ticketData.node_address}/v1/ws?ticket=${ticketData.ticket}&conversation_id=${conversationId}`

    const websocket = new WebSocket(websocketUrl)
    wsRef.current = websocket

    websocket.onopen = () => {
      console.log('WebSocket connected')
      setConnectionStatus('connected')
      setWs(websocket)
    }

    websocket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        const handlers = handlersRef.current

        switch (data.type) {
          case 'message':
            // Handle incoming message - this would be managed by useMessages hook
            break
          case 'presence':
            if (data.data.is_online) {
              handlers.addOnlineUser(data.data.user_id)
            } else {
              handlers.removeOnlineUser(data.data.user_id)
            }
            break
          case 'receipt':
            handlers.updateMessageStatus(data.data.message_id, data.data.status)
            break
          case 'typing':
            handlers.addTypingUser(data.data)
            break
          default:
            console.log('Unknown message type:', data.type)
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error)
      }
    }

    websocket.onclose = (event) => {
      console.log('WebSocket closed:', event.code, event.reason)
      setConnectionStatus('disconnected')
      setWs(null)
      wsRef.current = null

      // Try to reconnect after a delay if not closed intentionally
      if (event.code !== 1000) {
        setTimeout(() => {
          connectWebSocket()
        }, 3000)
      }
    }

    websocket.onerror = (error) => {
      console.error('WebSocket error:', error)
      setConnectionStatus('error')
    }
  }, [ticketData, user, isAuthenticated, conversationId, refetchTicket])

  // Track connection attempts to prevent multiple connections
  const connectionAttemptRef = useRef<boolean>(false)

  // Connect WebSocket when component mounts or ticket changes
  useEffect(() => {
    if (
      ticketData &&
      user &&
      isAuthenticated &&
      !connectionAttemptRef.current
    ) {
      connectionAttemptRef.current = true
      connectWebSocket()
    }

    return () => {
      connectionAttemptRef.current = false
      if (wsRef.current) {
        wsRef.current.close(1000) // Normal closure
      }
    }
  }, [ticketData, user, isAuthenticated, conversationId])

  // Handle sending messages
  const handleSendMessage = useCallback(
    async (content: string) => {
      if (!content.trim()) return

      const messageData: SendMessageRequest = {
        content: content.trim(),
        message_type: 'text',
      }

      try {
        await sendMessageMutation.mutateAsync(messageData)
        setReplyingTo(null) // Clear reply after sending
      } catch (error) {
        console.error('Failed to send message:', error)
      }
    },
    [sendMessageMutation, replyingTo],
  )

  // Handle reply to message
  const handleReply = useCallback((message: Message) => {
    setReplyingTo(message)
  }, [])

  // Handle typing events
  const handleTypingStart = useCallback(() => {
    startTyping()
  }, [startTyping])

  const handleTypingStop = useCallback(() => {
    stopTyping()
  }, [stopTyping])

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="p-8 text-center">
          <h2 className="text-xl font-semibold mb-4">
            Authentication Required
          </h2>
          <p className="text-muted-foreground mb-6">
            Please log in to access this conversation.
          </p>
          <Button asChild>
            <Link to={PATH_AUTH.login}>Sign In</Link>
          </Button>
        </Card>
      </div>
    )
  }

  if (isLoadingConversation || isLoadingTicket) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">Loading conversation...</p>
        </div>
      </div>
    )
  }

  if (!conversation) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="p-8 text-center">
          <h2 className="text-xl font-semibold mb-4">Conversation Not Found</h2>
          <p className="text-muted-foreground mb-6">
            This conversation might not exist or you don't have access to it.
          </p>
          <Button asChild>
            <Link to={PATH_DASHBOARD.chat.root}>Back to Conversations</Link>
          </Button>
        </Card>
      </div>
    )
  }

  return (
    <div
      className={`${isFullscreen ? 'h-screen' : 'h-full'} flex flex-col bg-gray-50 dark:bg-gray-900`}
    >
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b bg-white dark:bg-gray-800">
        <div className="flex items-center space-x-4">
          <Button
            onClick={() => navigate({ to: PATH_DASHBOARD.chat.root })}
            size="sm"
            variant="ghost"
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>

          <div>
            <h1 className="text-lg font-semibold">{conversation.name}</h1>
            <div className="flex items-center space-x-2 text-sm text-muted-foreground">
              <span
                className={`w-2 h-2 rounded-full ${connectionStatus === 'connected' ? 'bg-green-500' : connectionStatus === 'connecting' ? 'bg-yellow-500' : 'bg-red-500'}`}
              ></span>
              <span>
                {connectionStatus === 'connected'
                  ? 'Connected'
                  : connectionStatus === 'connecting'
                    ? 'Connecting...'
                    : connectionStatus === 'error'
                      ? 'Connection error'
                      : 'Disconnected'}
              </span>
            </div>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <Button size="sm" variant="ghost">
            <Phone className="h-4 w-4" />
          </Button>
          <Button size="sm" variant="ghost">
            <Video className="h-4 w-4" />
          </Button>
          <Button
            onClick={() => setShowParticipants(!showParticipants)}
            size="sm"
            variant="ghost"
          >
            <Users className="h-4 w-4" />
          </Button>
          <Button size="sm" variant="ghost">
            <Settings className="h-4 w-4" />
          </Button>
          {onToggleFullscreen && showToggle && (
            <Button
              onClick={onToggleFullscreen}
              size="sm"
              variant="ghost"
              title={isFullscreen ? 'Exit fullscreen' : 'Enter fullscreen'}
            >
              {isFullscreen ? (
                <Minimize2 className="h-4 w-4" />
              ) : (
                <Maximize2 className="h-4 w-4" />
              )}
            </Button>
          )}
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Messages Area */}
        <div className="flex-1 flex flex-col">
          <MessageList
            conversationId={conversationId}
            currentUserId={user?.id || ''}
            onReply={handleReply}
            typingUsers={typingUsers}
          />

          <MessageInput
            disabled={
              connectionStatus !== 'connected' || sendMessageMutation.isPending
            }
            isLoading={sendMessageMutation.isPending}
            onCancelReply={() => setReplyingTo(null)}
            onSendMessage={handleSendMessage}
            onTypingStart={handleTypingStart}
            onTypingStop={handleTypingStop}
            replyingTo={replyingTo}
          />
        </div>

        {/* Participants Sidebar */}
        {showParticipants && (
          <div className="w-80 border-l bg-white dark:bg-gray-800">
            <ParticipantsList
              conversationId={conversationId}
              currentUserId={user?.id || ''}
              isUserOnline={isUserOnline}
              onClose={() => setShowParticipants(false)}
            />
          </div>
        )}
      </div>
    </div>
  )
}
