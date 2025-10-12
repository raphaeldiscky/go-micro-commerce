import { PATH_AUTH, PATH_FEATURES } from '@/constants/routes'
import { useIsAuthenticated } from '@/hooks/auth'
import { useConversations } from '@/hooks/chat'
import { formatRelativeTime } from '@/lib/utils/date'
import type { Conversation } from '@/types/__generated__/graphql'
import { Link } from '@tanstack/react-router'
import {
  ChevronRight,
  Clock,
  MessageCircle,
  Plus,
  Search,
  Users,
} from 'lucide-react'
import { useState } from 'react'
import { Button } from '../../../ui/button'
import { Card, CardContent } from '../../../ui/card'
import { Input } from '../../../ui/input'
import { Skeleton } from '../../../ui/skeleton'

export function ChatListPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const isAuthenticated = useIsAuthenticated()

  const { data: conversations, error, isLoading, refetch } = useConversations()

  const getConversationName = (conversation: Conversation) => {
    return conversation.subject || 'Untitled Conversation'
  }

  const filteredConversations = conversations?.filter((conv) => {
    const name = getConversationName(conv)
    return name.toLowerCase().includes(searchQuery.toLowerCase())
  })

  const getConversationIcon = (status: Conversation['status']) => {
    switch (status) {
      case 'WAITING':
        return <Clock className="h-5 w-5 text-yellow-500" />
      case 'ACTIVE':
        return <MessageCircle className="h-5 w-5 text-green-500" />
      case 'ENDED':
        return <MessageCircle className="h-5 w-5 text-gray-500" />
      default:
        return <MessageCircle className="h-5 w-5 text-gray-500" />
    }
  }

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardContent className="flex flex-col items-center justify-center py-12">
            <MessageCircle className="h-16 w-16 text-muted-foreground mb-4" />
            <h2 className="text-2xl font-semibold mb-2">Welcome to Chat</h2>
            <p className="text-muted-foreground text-center mb-6">
              Please log in to access your conversations and start chatting.
            </p>
            <Button asChild>
              <Link to={PATH_AUTH.login}>Sign In</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 p-4">
        <div className="max-w-4xl mx-auto">
          <div className="mb-6">
            <Skeleton className="h-8 w-64 mb-2" />
            <Skeleton className="h-4 w-96" />
          </div>
          <div className="mb-6">
            <Skeleton className="h-10 w-full max-w-md" />
          </div>
          <div className="space-y-4">
            {Array.from({ length: 6 }).map((_, i) => (
              <Card key={i}>
                <CardContent className="p-4">
                  <div className="flex items-center space-x-4">
                    <Skeleton className="h-12 w-12 rounded-full" />
                    <div className="flex-1 space-y-2">
                      <Skeleton className="h-4 w-48" />
                      <Skeleton className="h-3 w-32" />
                    </div>
                    <Skeleton className="h-6 w-12" />
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 p-4">
        <div className="max-w-4xl mx-auto">
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <MessageCircle className="h-16 w-16 text-muted-foreground mb-4" />
              <h2 className="text-xl font-semibold mb-2">
                Unable to Load Conversations
              </h2>
              <p className="text-muted-foreground text-center mb-6">
                We're having trouble loading your conversations. Please try
                again.
              </p>
              <Button onClick={() => refetch()}>Try Again</Button>
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 p-4">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
            Conversations
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Choose a conversation to start messaging
          </p>
        </div>

        {/* Search and Actions */}
        <div className="flex flex-col sm:flex-row gap-4 mb-6">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
            <Input
              className="pl-10"
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search conversations..."
              value={searchQuery}
            />
          </div>
          <Button className="flex items-center gap-2" variant="outline">
            <Plus className="h-4 w-4" />
            New Conversation
          </Button>
        </div>

        {/* Conversations List */}
        <div className="space-y-3">
          {filteredConversations && filteredConversations.length > 0 ? (
            filteredConversations.map((conversation) => (
              <Card
                className="hover:shadow-md transition-shadow cursor-pointer group"
                key={conversation.id}
              >
                <Link
                  params={{ conversationId: conversation.id }}
                  to={PATH_FEATURES.chat.detail(conversation.id)}
                >
                  <CardContent className="p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-4 flex-1 min-w-0">
                        {/* Conversation Icon */}
                        <div className="flex-shrink-0">
                          {getConversationIcon(conversation.status)}
                        </div>

                        {/* Conversation Info */}
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center justify-between mb-1">
                            <h3 className="font-semibold text-gray-900 dark:text-white truncate">
                              {getConversationName(conversation)}
                            </h3>
                            <div className="flex items-center space-x-2 flex-shrink-0 ml-2">
                              <span className="text-xs text-gray-500 dark:text-gray-400 flex items-center">
                                <Clock className="h-3 w-3 mr-1" />
                                {formatRelativeTime(conversation.updatedAt)}
                              </span>
                            </div>
                          </div>

                          {/* Participants Count */}
                          <div className="flex items-center justify-between">
                            <div className="flex-1 min-w-0">
                              <p className="text-sm text-gray-500 dark:text-gray-500 italic">
                                {conversation.status === 'ACTIVE'
                                  ? 'Active conversation'
                                  : conversation.status === 'WAITING'
                                    ? 'Waiting for agent'
                                    : 'Conversation ended'}
                              </p>
                            </div>

                            {/* Participants Count */}
                            <div className="flex items-center text-xs text-gray-500 dark:text-gray-400 ml-2">
                              <Users className="h-3 w-3 mr-1" />
                              {conversation.participantCount}
                            </div>
                          </div>
                        </div>
                      </div>

                      {/* Arrow Icon */}
                      <ChevronRight className="h-5 w-5 text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300 transition-colors flex-shrink-0 ml-2" />
                    </div>
                  </CardContent>
                </Link>
              </Card>
            ))
          ) : (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12">
                <MessageCircle className="h-16 w-16 text-muted-foreground mb-4" />
                <h3 className="text-lg font-semibold mb-2">
                  {searchQuery
                    ? 'No conversations found'
                    : 'No conversations yet'}
                </h3>
                <p className="text-muted-foreground text-center">
                  {searchQuery
                    ? 'Try adjusting your search terms'
                    : 'Join a conversation or create one to start chatting'}
                </p>
                {searchQuery && (
                  <Button
                    className="mt-4"
                    onClick={() => setSearchQuery('')}
                    variant="outline"
                  >
                    Clear Search
                  </Button>
                )}
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  )
}
