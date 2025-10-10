import { PATH_AUTH, PATH_FEATURES } from '@/constants'
import { useIsAuthenticated, useUser } from '@/hooks/auth/useAuth'
import { useConversationDetails } from '@/hooks/chat/useConversationDetails'
import { useConversationSubscription } from '@/hooks/chat/useConversationSubscription'
import { useSendMessage } from '@/hooks/chat/useMessages'
import { usePresence } from '@/hooks/chat/usePresence'
import { useTypingIndicator } from '@/hooks/chat/useTypingIndicator'
import type { Message, SendMessageInput } from '@/types/__generated__/graphql'
import { MessageType } from '@/types/__generated__/graphql'
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
import { useCallback, useState } from 'react'
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

  const navigate = useNavigate()
  const user = useUser()
  const isAuthenticated = useIsAuthenticated()

  const { data: conversation, isLoading: isLoadingConversation } =
    useConversationDetails(conversationId)
  const sendMessageMutation = useSendMessage(conversationId, user?.id || '')
  const { startTyping, stopTyping, typingUsers } =
    useTypingIndicator(conversationId)
  const { isUserOnline } = usePresence()

  useConversationSubscription(conversationId)

  const handleSendMessage = useCallback(
    async (content: string) => {
      if (!content.trim()) return

      const messageData: SendMessageInput = {
        conversationId,
        content: content.trim(),
        messageType: MessageType.Text,
      }

      try {
        await sendMessageMutation.mutateAsync(messageData)
        setReplyingTo(null)
      } catch (error) {
        console.error('Failed to send message:', error)
      }
    },
    [sendMessageMutation],
  )

  const handleReply = useCallback((message: Message) => {
    setReplyingTo(message)
  }, [])

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
              <span className="w-2 h-2 rounded-full bg-green-500"></span>
              <span>Connected</span>
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

      <div className="flex-1 flex overflow-hidden">
        <div className="flex-1 flex flex-col">
          <MessageList
            conversationId={conversationId}
            currentUserId={user?.id || ''}
            onReply={handleReply}
            typingUsers={typingUsers}
          />

          <MessageInput
            disabled={sendMessageMutation.isPending}
            isLoading={sendMessageMutation.isPending}
            onCancelReply={() => setReplyingTo(null)}
            onSendMessage={handleSendMessage}
            onTypingStart={handleTypingStart}
            onTypingStop={handleTypingStop}
            replyingTo={replyingTo}
          />
        </div>

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
