import { PATH_AUTH, PATH_FEATURES } from '@/constants'
import { ChatWebSocketProvider } from '@/contexts/ChatWebSocketContext'
import { useIsAuthenticated, useUser } from '@/hooks/auth/useAuth'
import { useConversationDetails } from '@/hooks/chat/useConversationDetails'
import { useMessageReceipts } from '@/hooks/chat/useMessageReceipts'
import { useSendMessage } from '@/hooks/chat/useMessages'
import { usePresence } from '@/hooks/chat/usePresence'
import { useTypingIndicator } from '@/hooks/chat/useTypingIndicator'
import { getAccessToken } from '@/lib/api'
import type { SendMessageRequest } from '@/lib/api'
import { graphqlClient, REQUEST_CHAT_CONNECTION_MUTATION } from '@/lib/graphql'
import type { Message } from '@/types/__generated__/graphql'
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

interface ConversationPageContentProps extends ConversationPageProps {
  websocket: WebSocket | null
  connectionStatus: 'connected' | 'connecting' | 'disconnected' | 'error'
}

function ConversationPageContent({
  conversationId,
  isFullscreen = false,
  onToggleFullscreen,
  showToggle = true,
  websocket,
  connectionStatus,
}: ConversationPageContentProps) {
  const [showParticipants, setShowParticipants] = useState(false)
  const [replyingTo, setReplyingTo] = useState<Message | null>(null)

  const navigate = useNavigate()
  const user = useUser()

  // Hooks
  const { data: conversation, isLoading: isLoadingConversation } =
    useConversationDetails(conversationId)
  const sendMessageMutation = useSendMessage(conversationId)
  const { addTypingUser, startTyping, stopTyping, typingUsers } =
    useTypingIndicator(conversationId)
  const { addOnlineUser, isUserOnline, removeOnlineUser } = usePresence()
  const { updateMessageStatus } = useMessageReceipts(conversationId)

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

  // Handle WebSocket messages
  useEffect(() => {
    if (!websocket) return

    const handleMessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data)
        const handlers = handlersRef.current

        switch (data.type) {
          case 'message':
            // Handle incoming message - managed by useMessages hook
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

    websocket.addEventListener('message', handleMessage)

    return () => {
      websocket.removeEventListener('message', handleMessage)
    }
  }, [websocket])

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

  if (isLoadingConversation) {
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
            <Link to={PATH_FEATURES.chat.root}>Back to Conversations</Link>
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
            onClick={() => navigate({ to: PATH_FEATURES.chat.root })}
            size="sm"
            variant="ghost"
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>

          <div>
            <h1 className="text-lg font-semibold">
              {conversation.subject || 'Conversation'}
            </h1>
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

// Wrapper component that provides WebSocket context
export function ConversationPage(props: ConversationPageProps) {
  const [ws, setWs] = useState<null | WebSocket>(null)
  const [connectionStatus, setConnectionStatus] = useState<
    'connected' | 'connecting' | 'disconnected' | 'error'
  >('disconnected')

  const user = useUser()
  const isAuthenticated = useIsAuthenticated()

  const wsRef = useRef<null | WebSocket>(null)

  // WebSocket connection management
  const connectWebSocket = useCallback(async () => {
    if (!user || !isAuthenticated) return

    // Get JWT access token
    const accessToken = getAccessToken()
    if (!accessToken) {
      console.error('No access token available')
      return
    }

    // Close existing connection if any
    if (wsRef.current && wsRef.current.readyState !== WebSocket.CLOSED) {
      wsRef.current.close()
    }

    setConnectionStatus('connecting')

    try {
      // Request node address from GraphQL
      const response = await graphqlClient.request<{
        requestChatConnection: {
          nodeAddress: string
          userId: string
          userType: string
        }
      }>(REQUEST_CHAT_CONNECTION_MUTATION)

      // Connect to WebSocket with JWT token
      const websocketUrl = `${response.requestChatConnection.nodeAddress}/v1/ws?token=${accessToken}&conversation_id=${props.conversationId}`

      const websocket = new WebSocket(websocketUrl)
      wsRef.current = websocket

      websocket.onopen = () => {
        console.log('WebSocket connected')
        setConnectionStatus('connected')
        setWs(websocket)
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
    } catch (error) {
      console.error('Failed to establish WebSocket connection:', error)
      setConnectionStatus('error')
    }
  }, [user, isAuthenticated, props.conversationId])

  // Track connection attempts to prevent multiple connections
  const connectionAttemptRef = useRef<boolean>(false)

  // Connect WebSocket when component mounts
  useEffect(() => {
    if (user && isAuthenticated && !connectionAttemptRef.current) {
      connectionAttemptRef.current = true
      connectWebSocket()
    }

    return () => {
      connectionAttemptRef.current = false
      if (wsRef.current) {
        wsRef.current.close(1000) // Normal closure
      }
    }
  }, [user, isAuthenticated, props.conversationId, connectWebSocket])

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

  return (
    <ChatWebSocketProvider connectionStatus={connectionStatus} websocket={ws}>
      <ConversationPageContent
        {...props}
        connectionStatus={connectionStatus}
        websocket={ws}
      />
    </ChatWebSocketProvider>
  )
}
