import { useConversations } from '@/hooks/chat/useConversations'
import { useIsAuthenticated } from '@/hooks/useAuth'
import { MessageCircle } from 'lucide-react'
import { useState } from 'react'
import { Card, CardContent } from '../../ui/card'
import { ConversationList } from '../lists/ConversationList'
import { ChatInterface } from './ChatInterface'

export function ChatApp() {
  const [selectedConversationId, setSelectedConversationId] = useState<string>()
  const isAuthenticated = useIsAuthenticated()
  const { data: conversations } = useConversations()

  const handleConversationSelect = (conversationId: string) => {
    setSelectedConversationId(conversationId)
  }

  const selectedConversation = conversations?.find(
    (conv) => conv.id === selectedConversationId,
  )

  if (!isAuthenticated) {
    return (
      <div className="h-screen flex items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardContent className="flex flex-col items-center justify-center py-8">
            <MessageCircle className="h-16 w-16 text-muted-foreground mb-4" />
            <h2 className="text-xl font-semibold mb-2">Welcome to Chat</h2>
            <p className="text-muted-foreground text-center mb-4">
              Please log in to access conversations and start chatting with
              others.
            </p>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="h-screen flex p-4 gap-4">
      {/* Conversations Sidebar */}
      <ConversationList
        onConversationSelect={handleConversationSelect}
        selectedConversationId={selectedConversationId}
      />

      {/* Main Chat Area */}
      <div className="flex-1 flex flex-col">
        {selectedConversation ? (
          <ChatInterface
            conversationId={selectedConversation.id}
            conversationName={selectedConversation.subject}
          />
        ) : (
          <Card className="flex-1 h-full flex flex-col">
            <CardContent className="flex-1 flex items-center justify-center">
              <div className="text-center text-muted-foreground">
                <MessageCircle className="h-16 w-16 mx-auto mb-4 opacity-50" />
                <h3 className="text-lg font-semibold mb-2">
                  No Conversation Selected
                </h3>
                <p className="text-sm max-w-md">
                  Choose a conversation from the sidebar to start chatting, or
                  join a new conversation to begin messaging.
                </p>
                {conversations && conversations.length === 0 && (
                  <p className="text-xs mt-4 opacity-75">
                    No conversations available. Contact an administrator to join
                    conversations.
                  </p>
                )}
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}
