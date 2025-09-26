import {
  useConversationDetails,
  useConversationParticipants,
} from '@/hooks/chat/useConversationDetails'
import { cn } from '@/lib/utils'
import { Crown, MoreVertical, Settings, UserMinus, X } from 'lucide-react'
import { useState } from 'react'
import { Avatar, AvatarFallback, AvatarImage } from '../../ui/avatar'
import { Button } from '../../ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '../../ui/card'
import { ScrollArea } from '../../ui/scroll-area'
import { Skeleton } from '../../ui/skeleton'

interface ParticipantsListProps {
  conversationId: string
  currentUserId: string
  isUserOnline: (userId: string) => boolean
  onClose: () => void
}

export function ParticipantsList({
  conversationId,
  currentUserId,
  isUserOnline,
  onClose,
}: ParticipantsListProps) {
  const [selectedParticipant, setSelectedParticipant] = useState<null | string>(
    null,
  )

  const {
    data: conversation,
    error: conversationError,
    isLoading: conversationLoading,
  } = useConversationDetails(conversationId)

  const {
    data: participants,
    error: participantsError,
    isLoading: participantsLoading,
  } = useConversationParticipants(conversationId)

  const isLoading = conversationLoading || participantsLoading
  const error = conversationError || participantsError

  const getInitials = (userId: string) => {
    // Use user_id as fallback since we don't have display names
    return userId.slice(0, 2).toUpperCase()
  }

  const getRoleIcon = (userType: string) => {
    switch (userType) {
      case 'admin':
        return <Crown className="h-3 w-3 text-yellow-500" />
      case 'moderator':
        return <Settings className="h-3 w-3 text-blue-500" />
      default:
        return null
    }
  }

  const getRoleColor = (userType: string) => {
    switch (userType) {
      case 'admin':
        return 'text-yellow-600 dark:text-yellow-400'
      case 'moderator':
        return 'text-blue-600 dark:text-blue-400'
      default:
        return 'text-gray-500 dark:text-gray-400'
    }
  }

  if (isLoading) {
    return (
      <Card className="w-full h-full flex flex-col">
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg">Participants</CardTitle>
            <Button onClick={onClose} size="sm" variant="ghost">
              <X className="h-4 w-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent className="flex-1">
          <div className="space-y-3">
            {Array.from({ length: 8 }).map((_, i) => (
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

  if (error || !conversation) {
    return (
      <Card className="w-full h-full flex flex-col">
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg">Participants</CardTitle>
            <Button onClick={onClose} size="sm" variant="ghost">
              <X className="h-4 w-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="text-center text-muted-foreground py-8">
            <p>Failed to load participants</p>
          </div>
        </CardContent>
      </Card>
    )
  }

  const participantsList = participants || []
  const sortedParticipants = [...participantsList].sort((a, b) => {
    // Sort by online status first, then by user_type, then by user_id
    const aOnline = isUserOnline(a.user_id)
    const bOnline = isUserOnline(b.user_id)

    if (aOnline !== bOnline) return bOnline ? 1 : -1

    const roleOrder = { admin: 0, moderator: 1, user: 2 }
    const aRoleOrder = roleOrder[a.user_type as keyof typeof roleOrder] || 2
    const bRoleOrder = roleOrder[b.user_type as keyof typeof roleOrder] || 2

    if (aRoleOrder !== bRoleOrder) return aRoleOrder - bRoleOrder

    return a.user_id.localeCompare(b.user_id)
  })

  const onlineCount = participantsList.filter((p: { user_id: string }) =>
    isUserOnline(p.user_id),
  ).length

  return (
    <Card className="w-full h-full flex flex-col">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-lg">Participants</CardTitle>
            <p className="text-sm text-muted-foreground">
              {onlineCount} online • {participantsList.length} total
            </p>
          </div>
          <Button onClick={onClose} size="sm" variant="ghost">
            <X className="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>

      <CardContent className="flex-1 overflow-hidden p-0">
        <ScrollArea className="h-full">
          <div className="p-4 space-y-1">
            {sortedParticipants.map((participant) => {
              const isOnline = isUserOnline(participant.user_id)
              const isCurrentUser = participant.user_id === currentUserId

              return (
                <div
                  className={cn(
                    'flex items-center space-x-3 p-2 rounded-lg cursor-pointer transition-colors',
                    'hover:bg-gray-100 dark:hover:bg-gray-700',
                    selectedParticipant === participant.user_id &&
                      'bg-gray-100 dark:bg-gray-700',
                  )}
                  key={participant.user_id}
                  onClick={() =>
                    setSelectedParticipant(
                      selectedParticipant === participant.user_id
                        ? null
                        : participant.user_id,
                    )
                  }
                >
                  {/* Avatar with Online Status */}
                  <div className="relative">
                    <Avatar className="h-10 w-10">
                      <AvatarImage alt={participant.user_id} src={undefined} />
                      <AvatarFallback className="text-sm bg-gradient-to-br from-blue-500 to-purple-600 text-white">
                        {getInitials(participant.user_id)}
                      </AvatarFallback>
                    </Avatar>

                    {/* Online Status Indicator */}
                    <div
                      className={cn(
                        'absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-white dark:border-gray-800',
                        isOnline ? 'bg-green-500' : 'bg-gray-400',
                      )}
                    />
                  </div>

                  {/* User Info */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center space-x-2">
                      <p
                        className={cn(
                          'font-medium text-sm truncate',
                          isCurrentUser && 'text-blue-600 dark:text-blue-400',
                        )}
                      >
                        {participant.user_id}
                        {isCurrentUser && ' (You)'}
                      </p>
                      {getRoleIcon(participant.user_type)}
                    </div>

                    <div className="flex items-center space-x-2">
                      <p
                        className={cn(
                          'text-xs capitalize',
                          getRoleColor(participant.user_type),
                        )}
                      >
                        {participant.user_type}
                      </p>
                      {isOnline && (
                        <span className="text-xs text-green-600 dark:text-green-400">
                          Online
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Actions Menu */}
                  {!isCurrentUser && (
                    <div className="flex items-center space-x-1">
                      <Button
                        className="h-6 w-6 p-0 opacity-0 group-hover:opacity-100 transition-opacity"
                        onClick={(e) => {
                          e.stopPropagation()
                          // Handle participant actions
                        }}
                        size="sm"
                        variant="ghost"
                      >
                        <MoreVertical className="h-3 w-3" />
                      </Button>
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </ScrollArea>
      </CardContent>

      {/* Actions Footer */}
      <div className="border-t p-3">
        <div className="flex justify-between items-center">
          <Button className="text-xs" size="sm" variant="outline">
            Invite People
          </Button>
          <Button
            className="text-xs text-red-600 hover:text-red-700"
            size="sm"
            variant="ghost"
          >
            <UserMinus className="h-3 w-3 mr-1" />
            Leave
          </Button>
        </div>
      </div>
    </Card>
  )
}
