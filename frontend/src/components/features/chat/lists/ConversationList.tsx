import {
  useConversations,
  useJoinConversation,
} from '@/hooks/chat/useConversations'
import { formatRelativeTime } from '@/lib/utils/date'
import type { ParticipantRole } from '@/types/__generated__/graphql'
import { MessageCircle } from 'lucide-react'
import { Button } from '../../../ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '../../../ui/card'
import { ScrollArea } from '../../../ui/scroll-area'
import { Skeleton } from '../../../ui/skeleton'

interface ConversationListProps {
  onConversationSelect: (conversationId: string) => void
  selectedConversationId?: string
}

export function ConversationList({
  onConversationSelect,
  selectedConversationId,
}: ConversationListProps) {
  const { data: conversations, error, isLoading } = useConversations()

  const joinConversationMutation = useJoinConversation()

  const handleJoinConversation = (conversationId: string) => {
    joinConversationMutation.mutate(
      {
        conversationId,
        role: 'MEMBER' as ParticipantRole,
      },
      {
        onSuccess: () => {
          onConversationSelect(conversationId)
        },
      },
    )
  }

  if (isLoading) {
    return (
      <Card className="w-80 h-full">
        <CardHeader>
          <CardTitle>Conversations</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {Array.from({ length: 5 }).map((_, i) => (
              <div className="flex items-center space-x-3" key={i}>
                <Skeleton className="h-10 w-10 rounded-full" />
                <div className="space-y-2 flex-1">
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-3 w-3/4" />
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card className="w-80 h-full">
        <CardHeader>
          <CardTitle>Conversations</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center text-muted-foreground">
            <MessageCircle className="h-12 w-12 mx-auto mb-2 opacity-50" />
            <p>Failed to load conversations</p>
            <Button
              className="mt-2"
              onClick={() => window.location.reload()}
              size="sm"
              variant="outline"
            >
              Retry
            </Button>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className="w-80 h-full flex flex-col">
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2">
          <MessageCircle className="h-5 w-5" />
          Conversations
        </CardTitle>
      </CardHeader>
      <CardContent className="flex-1 overflow-hidden">
        <ScrollArea className="h-full">
          {conversations && conversations.length > 0 ? (
            <div className="space-y-2">
              {conversations.map((conversation) => (
                <div
                  className={`p-3 rounded-lg border cursor-pointer transition-colors hover:bg-accent ${
                    selectedConversationId === conversation.id
                      ? 'bg-accent border-primary'
                      : 'border-border'
                  }`}
                  key={conversation.id}
                  onClick={() => handleJoinConversation(conversation.id)}
                >
                  <div className="flex items-start justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <MessageCircle className="h-4 w-4" />
                      <span className="font-medium text-sm truncate">
                        {conversation.subject || 'Untitled Conversation'}
                      </span>
                    </div>
                    <div className="flex items-center gap-1">
                      <span className="text-xs text-muted-foreground">
                        {formatRelativeTime(conversation.updatedAt)}
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center justify-between">
                    <div className="flex-1 min-w-0">
                      <p className="text-xs text-muted-foreground">
                        Status: {conversation.status}
                      </p>
                    </div>
                    <div className="flex items-center gap-1 ml-2">
                      <span className="text-xs text-muted-foreground">
                        Priority: {conversation.priority}
                      </span>
                    </div>
                  </div>

                  {joinConversationMutation.isPending &&
                    selectedConversationId === conversation.id && (
                      <div className="mt-2 text-xs text-muted-foreground">
                        Joining...
                      </div>
                    )}
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center text-muted-foreground py-8">
              <MessageCircle className="h-12 w-12 mx-auto mb-2 opacity-50" />
              <p>No conversations found</p>
              <p className="text-xs mt-1">
                Join a conversation to start chatting
              </p>
            </div>
          )}
        </ScrollArea>
      </CardContent>
    </Card>
  )
}
